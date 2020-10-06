package smith

import (
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	// BlockchainProcessorInterface represents interface for the blockchain processor's implementations
	BlockchainProcessorInterface interface {
		Start(sleepPeriod time.Duration)
		Stop()
		StartSmithing() error
		FakeSmithing(numberOfBlocks int, fromGenesis bool, chainType chaintype.ChainType) error
		GetBlockChainprocessorStatus() (isSmithing bool, err error)
	}

	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		ChainType               chaintype.ChainType
		Generator               *model.Blocksmith
		BlockService            service.BlockServiceInterface
		LastBlockID             int64
		Logger                  *log.Logger
		smithError              error
		BlockchainStatusService service.BlockchainStatusServiceInterface
		NodeRegistrationService service.NodeRegistrationServiceInterface
	}
)

var (
	stopSmith = make(chan bool)
)

// NewBlockchainProcessor create new instance of BlockchainProcessor
func NewBlockchainProcessor(
	ct chaintype.ChainType,
	blocksmith *model.Blocksmith,
	blockService service.BlockServiceInterface,
	logger *log.Logger,
	blockchainStatusService service.BlockchainStatusServiceInterface,
	nodeRegistrationService service.NodeRegistrationServiceInterface,
) *BlockchainProcessor {
	return &BlockchainProcessor{
		ChainType:               ct,
		Generator:               blocksmith,
		BlockService:            blockService,
		Logger:                  logger,
		BlockchainStatusService: blockchainStatusService,
		NodeRegistrationService: nodeRegistrationService,
	}
}

// FakeSmithing should only be used in testing the blockchain, it's not meant to be used in production, and could cause
// errors
// todo: @andy-shi need to adjust this function to newest state of smithing process.
func (bp *BlockchainProcessor) FakeSmithing(numberOfBlocks int, fromGenesis bool, ct chaintype.ChainType) error {
	// todo: if debug mode, allow, else no
	var (
		timeNow int64
	)
	err := bp.BlockService.UpdateLastBlockCache(nil)
	if err != nil {
		return err
	}
	// creating a virtual time
	if !fromGenesis {
		lastBlock, err := bp.BlockService.GetLastBlock()
		if err != nil {
			return err
		}
		timeNow = lastBlock.Timestamp
	} else {
		timeNow = constant.MainchainGenesisBlockTimestamp
	}
	for i := 0; i < numberOfBlocks; i++ {
		lastBlock, err := bp.BlockService.GetLastBlock()
		if err != nil {
			return blocker.NewBlocker(
				blocker.SmithingErr, "genesis block has not been applied")
		}
		// simulating real condition, calculating the smith time of current last block
		if lastBlock.GetID() != bp.LastBlockID {
			bp.LastBlockID = lastBlock.GetID()
			err = bp.BlockService.GetBlocksmithStrategy().CalculateScore(bp.Generator, 1)
			if err != nil {
				return err
			}
		}
		// speed up the virtual time if smith time has not reach the needed smithing maximum time
		for timeNow < lastBlock.GetTimestamp()+ct.GetSmithingPeriod() {
			timeNow++ // speed up bro
		}
		previousBlock, err := bp.BlockService.GetLastBlock()
		if err != nil {
			return err
		}
		block, err := bp.BlockService.GenerateBlock(
			previousBlock,
			bp.Generator.SecretPhrase,
			timeNow,
			true,
		)
		if err != nil {
			return err
		}
		// validate
		err = bp.BlockService.ValidateBlock(block, previousBlock) // err / !err
		if err != nil {
			blockerUsed := blocker.ValidateMainBlockErr
			if chaintype.IsSpineChain(bp.ChainType) {
				blockerUsed = blocker.ValidateSpineBlockErr
			}
			bp.Logger.Warnf("FakeSmithing: %v\n", blocker.NewBlocker(blockerUsed, err.Error(), block, previousBlock))
			return err
		}
		// if validated push
		err = bp.BlockService.PushBlock(previousBlock, block, false, true)
		if err != nil {
			blockerUsed := blocker.PushMainBlockErr
			if chaintype.IsSpineChain(bp.ChainType) {
				blockerUsed = blocker.PushSpineBlockErr
			}
			bp.Logger.Errorf("FakeSmithing pushBlock fail: %v", blocker.NewBlocker(blockerUsed, err.Error(), block, previousBlock))
			return err
		}
	}
	return nil
}

// StartSmithing start smithing loop
func (bp *BlockchainProcessor) StartSmithing() error {
	if bp.Generator.NodeID == 0 {
		node, err := bp.NodeRegistrationService.GetNodeRegistrationByNodePublicKey(bp.Generator.NodePublicKey)
		if err != nil {
			return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("fail-GetNodeRegistrationByNodePublicKey: %v", err))
		} else if node == nil {
			return blocker.NewBlocker(blocker.ValidationErr, "BlocksmithNotInRegistry")
		}
		bp.Generator.NodeID = node.NodeID
	}
	// Securing smithing process
	// will pause another process that used block service lock until this process done
	bp.BlockService.ChainWriteLock(constant.BlockchainStatusGeneratingBlock)
	defer bp.BlockService.ChainWriteUnlock(constant.BlockchainStatusGeneratingBlock)

	var (
		blocksmithIndex int64
		lastBlock, err  = bp.BlockService.GetLastBlock()
	)
	if err != nil {
		return blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}
	// todo: move this piece of code to service layer
	// caching: only calculate smith time once per new block
	bp.LastBlockID, blocksmithIndex, err = bp.BlockService.WillSmith(
		bp.Generator, bp.LastBlockID,
	)
	if err != nil {
		return err
	}
	timestamp := time.Now().Unix()
	block, err := bp.BlockService.GenerateBlock(
		lastBlock,
		bp.Generator.SecretPhrase,
		timestamp,
		blocksmithIndex >= constant.EmptyBlockSkippedBlocksmithLimit,
	)
	if err != nil {
		return err
	}
	// validate
	err = bp.BlockService.ValidateBlock(block, lastBlock)
	if err != nil {
		blockerErr, ok := err.(blocker.Blocker)
		if ok && blockerErr.Type != blocker.InvalidBlockTimestamp {
			blockerUsed := blocker.ValidateMainBlockErr
			if chaintype.IsSpineChain(bp.ChainType) {
				blockerUsed = blocker.ValidateSpineBlockErr
			}
			bp.Logger.Warnf("StartSmithing: %v\n", blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()))
		}
		return err
	}
	// if validated push
	err = bp.BlockService.PushBlock(lastBlock, block, true, false)
	if err != nil {
		blockerUsed := blocker.PushMainBlockErr
		if chaintype.IsSpineChain(bp.ChainType) {
			blockerUsed = blocker.PushSpineBlockErr
		}
		bp.Logger.Errorf("StartSmithing pushBlock fail: %v", blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()))
		return err
	}
	return nil
}

// Start starts the blockchainProcessor
func (bp *BlockchainProcessor) Start(sleepPeriod time.Duration) {
	ticker := time.NewTicker(sleepPeriod)
	go func() {
		for {
			select {
			case <-stopSmith:
				bp.Logger.Infof("Stopped smithing %s", bp.BlockService.GetChainType().GetName())
				bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, false)
				bp.smithError = nil
				ticker.Stop()
				return
			case <-ticker.C:
				// when starting a node, do not start smithing until the main blocks have been fully downloaded
				if !bp.BlockchainStatusService.IsSmithingLocked() && bp.BlockchainStatusService.IsBlocksmith() {
					err := bp.StartSmithing()
					if err != nil {
						bp.Logger.Debugf("Smith Error for %s. %s", bp.BlockService.GetChainType().GetName(), err.Error())
						bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, false)
						bp.smithError = err
						if blockErr, ok := err.(blocker.Blocker); ok && blockErr.Type == blocker.ZeroParticipationScoreErr {
							bp.BlockchainStatusService.SetIsBlocksmith(false)
						}
					} else {
						bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, true)
						bp.smithError = nil
					}
				} else {
					bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, false)
				}
			}
		}
	}()
}

// Stop stops the blockchainProcessor
func (*BlockchainProcessor) Stop() {
	stopSmith <- true
}

// GetBlockChainprocessorStatus return the smithing status for this blockchain processor
func (bp *BlockchainProcessor) GetBlockChainprocessorStatus() (isSmithing bool, err error) {
	return bp.BlockchainStatusService.IsSmithing(bp.ChainType), bp.smithError
}

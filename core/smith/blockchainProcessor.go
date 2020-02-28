package smith

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"time"

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
		FakeSmithing(numberOfBlocks int, fromGenesis bool) error
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
) *BlockchainProcessor {
	return &BlockchainProcessor{
		ChainType:               ct,
		Generator:               blocksmith,
		BlockService:            blockService,
		Logger:                  logger,
		BlockchainStatusService: blockchainStatusService,
	}
}

// FakeSmithing should only be used in testing the blockchain, it's not meant to be used in production, and could cause
// errors
func (bp *BlockchainProcessor) FakeSmithing(numberOfBlocks int, fromGenesis bool) error {
	// todo: if debug mode, allow, else no
	var (
		timeNow int64
	)
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
			err = bp.BlockService.GetBlocksmithStrategy().CalculateSmith(lastBlock, 0, bp.Generator, 1)
			if err != nil {
				return err
			}
		}
		// speed up the virtual time if smith time has not reach the needed smithing maximum time
		for bp.Generator.SmithTime > timeNow {
			timeNow++ // speed up bro
		}
		// todo: replace to smithing time >= timestamp
		if bp.Generator.SmithTime > timeNow {
			return blocker.NewBlocker(
				blocker.SmithingErr, "verify seed return false",
			)
		}
		previousBlock, err := bp.BlockService.GetLastBlock()
		if err != nil {
			return err
		}
		block, err := bp.BlockService.GenerateBlock(
			previousBlock,
			bp.Generator.SecretPhrase,
			timeNow,
		)
		if err != nil {
			return err
		}
		// validate
		err = bp.BlockService.ValidateBlock(block, previousBlock, timeNow) // err / !err
		if err != nil {
			return err
		}
		// if validated push
		err = bp.BlockService.PushBlock(previousBlock, block, false, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// StartSmithing start smithing loop
func (bp *BlockchainProcessor) StartSmithing() error {
	// Securing smithing process
	// will pause another process that used block service lock until this process done
	bp.BlockService.ChainWriteLock(constant.BlockchainStatusGeneratingBlock)
	defer bp.BlockService.ChainWriteUnlock(constant.BlockchainStatusGeneratingBlock)

	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}
	// todo: move this piece of code to service layer
	// caching: only calculate smith time once per new block
	bp.LastBlockID, err = bp.BlockService.WillSmith(
		bp.Generator, bp.LastBlockID,
	)
	if err != nil {
		return err
	}
	timestamp := time.Now().Unix()
	if bp.Generator.SmithTime > timestamp {
		return nil
	}
	block, err := bp.BlockService.GenerateBlock(
		lastBlock,
		bp.Generator.SecretPhrase,
		timestamp,
	)
	if err != nil {
		return err
	}
	// validate
	err = bp.BlockService.ValidateBlock(block, lastBlock, timestamp)
	if err != nil {
		return err
	}
	// if validated push
	err = bp.BlockService.PushBlock(lastBlock, block, true, false)
	if err != nil {
		return err
	}
	return nil
}

// Start starts the blockchainProcessor
func (bp *BlockchainProcessor) Start(sleepPeriod time.Duration) {
	ticker := time.NewTicker(sleepPeriod)
	stopSmith = make(chan bool)
	go func() {
		for {
			select {
			case <-stopSmith:
				ticker.Stop()
				bp.Logger.Infof("Stopped smithing %s", bp.BlockService.GetChainType().GetName())
				bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, false)
				bp.smithError = nil
				return
			case <-ticker.C:
				// when starting a node, do not start smithing until the main blocks have been fully downloaded
				if !bp.BlockchainStatusService.IsSmithingLocked() {
					err := bp.StartSmithing()
					if err != nil {
						bp.Logger.Debugf("Smith Error for %s. %s", bp.BlockService.GetChainType().GetName(), err.Error())
						bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, false)
						bp.smithError = err
					}
					bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, true)
					bp.smithError = nil
				} else {
					bp.BlockchainStatusService.SetIsSmithing(bp.ChainType, true)
					bp.Logger.Debug("Smithing process is locked...")
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

package smith

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	// BlockchainProcessorInterface represents interface for the blockchain processor's implementations
	BlockchainProcessorInterface interface {
		Start(sleepPeriod int)
		Stop()
		StartSmithing() error
		FakeSmithing(numberOfBlocks int, fromGenesis bool) error
		GetBlockChainprocessorStatus() (isSmithing bool, err error)
	}

	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		Generator    *model.Blocksmith
		BlockService service.BlockServiceInterface
		LastBlockID  int64
		canSmith     bool
		Logger       *log.Logger
		isSmithing   bool
		smithError   error
	}
)

var (
	stopSmith = make(chan bool)
)

// NewBlockchainProcessor create new instance of BlockchainProcessor
func NewBlockchainProcessor(
	blocksmith *model.Blocksmith,
	blockService service.BlockServiceInterface,
	logger *log.Logger,
) *BlockchainProcessor {
	return &BlockchainProcessor{
		Generator:    blocksmith,
		BlockService: blockService,
		Logger:       logger,
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
		err = bp.BlockService.PushBlock(previousBlock, block, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// StartSmithing start smithing loop
func (bp *BlockchainProcessor) StartSmithing() error {
	var (
		blocksmithScore int64
	)
	// Securing smithing process
	// will pause another process that used block service lock until this process done
	bp.BlockService.ChainWriteLock(constant.BlockchainStatusGeneratingBlock)
	defer bp.BlockService.ChainWriteUnlock(constant.BlockchainStatusGeneratingBlock)

	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}
	// caching: only calculate smith time once per new block
	if lastBlock.GetID() != bp.LastBlockID {
		bp.LastBlockID = lastBlock.GetID()
		blockSmithStrategy := bp.BlockService.GetBlocksmithStrategy()
		blockSmithStrategy.SortBlocksmiths(lastBlock, true)
		// check if eligible to create block in this round
		blocksmithsMap := blockSmithStrategy.GetSortedBlocksmithsMap(lastBlock)
		if blocksmithsMap[string(bp.Generator.NodePublicKey)] == nil {
			bp.canSmith = false
			return blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
		}
		bp.canSmith = true
		ct := bp.BlockService.GetChainType()
		// calculate blocksmith score for the block type
		switch ct.(type) {
		case *chaintype.MainChain:
			// get the concrete type for BlockService so we can use mainchain specific methods
			blockMainService, ok := bp.BlockService.(*service.BlockService)
			if !ok {
				return blocker.NewBlocker(blocker.AppErr, "InvalidChaintype")
			}
			// try to get the node's participation score (ps) from node public key
			// if node is not registered, ps will be 0 and this node won't be able to smith
			// the default ps is 100000, smithing could be slower than when using account balances
			// since default balance was 1000 times higher than default ps
			blocksmithScore, err = blockMainService.GetParticipationScore(bp.Generator.NodePublicKey)
			if blocksmithScore <= 0 {
				bp.Logger.Info("Node has participation score <= 0. Either is not registered or has been expelled from node registry")
			}
			if err != nil || blocksmithScore < 0 {
				// no negative scores allowed
				blocksmithScore = 0
				bp.Logger.Errorf("Participation score calculation: %s", err)
			}
		case *chaintype.SpineChain:
			// FIXME: ask @barton how to compute score for spine blocksmiths, since we don't have participation score and receipts attached to them?
			blocksmithScore = constant.DefaultParticipationScore
		default:
			return blocker.NewBlocker(blocker.SmithingErr, fmt.Sprintf("undefined chaintype %s", ct.GetName()))
		}
		err = blockSmithStrategy.CalculateSmith(
			lastBlock,
			*(blocksmithsMap[string(bp.Generator.NodePublicKey)]),
			bp.Generator,
			blocksmithScore,
		)
		if err != nil {
			return err
		}
		monitoring.SetBlockchainSmithTime(ct.GetTypeInt(), bp.Generator.SmithTime-lastBlock.Timestamp)
	}
	if !bp.canSmith {
		return blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
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
	err = bp.BlockService.PushBlock(lastBlock, block, true)
	if err != nil {
		return err
	}
	return nil
}

// Start starts the blockchainProcessor
func (bp *BlockchainProcessor) Start(sleepPeriod int) {
	ticker := time.NewTicker(time.Duration(sleepPeriod) * time.Millisecond)
	stopSmith = make(chan bool)
	go func() {
		for {
			select {
			case <-stopSmith:
				ticker.Stop()
				bp.Logger.Infof("Stopped smithing %s", bp.BlockService.GetChainType().GetName())
				bp.isSmithing = false
				bp.smithError = nil
				return
			case <-ticker.C:
				err := bp.StartSmithing()
				if err != nil {
					bp.Logger.Debugf("Smith Error for %s. %s", bp.BlockService.GetChainType().GetName(), err.Error())
					bp.isSmithing = false
					bp.smithError = err
				}
				bp.isSmithing = true
				bp.smithError = nil
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
	return bp.isSmithing, bp.smithError
}

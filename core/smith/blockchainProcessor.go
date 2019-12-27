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
	"github.com/zoobc/zoobc-core/core/smith/strategy"
)

type (
	// BlockchainProcessorInterface represents interface for the blockchain processor's implementations
	BlockchainProcessorInterface interface {
		StartSmithing() error
		FakeSmithing(numberOfBlocks int, fromGenesis bool) error
	}

	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		Chaintype               chaintype.ChainType
		Generator               *model.Blocksmith
		BlockService            service.BlockServiceInterface
		BlocksmithService       strategy.BlocksmithStrategyInterface
		NodeRegistrationService service.NodeRegistrationServiceInterface
		LastBlockID             int64
		canSmith                bool
		Logger                  *log.Logger
	}
)

// NewBlockchainProcessor create new instance of BlockchainProcessor
func NewBlockchainProcessor(
	ct chaintype.ChainType,
	blocksmith *model.Blocksmith,
	blockService service.BlockServiceInterface,
	blocksmithStrategy strategy.BlocksmithStrategyInterface,
	nodeRegistrationService service.NodeRegistrationServiceInterface,
	logger *log.Logger,
) *BlockchainProcessor {
	return &BlockchainProcessor{
		Chaintype:               ct,
		Generator:               blocksmith,
		BlockService:            blockService,
		BlocksmithService:       blocksmithStrategy,
		NodeRegistrationService: nodeRegistrationService,
		Logger:                  logger,
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
			err = bp.BlocksmithService.CalculateSmith(lastBlock, 0, bp.Generator, 1)
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
		bp.BlocksmithService.SortBlocksmiths(lastBlock)
		// check if eligible to create block in this round
		blocksmithsMap := bp.BlocksmithService.GetSortedBlocksmithsMap(lastBlock)
		if blocksmithsMap[string(bp.Generator.NodePublicKey)] == nil {
			bp.canSmith = false
			return blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
		}
		bp.canSmith = true
		// calculate blocksmith score for the block type
		switch bp.Chaintype.(type) {
		case *chaintype.MainChain:
			// try to get the node's participation score (ps) from node public key
			// if node is not registered, ps will be 0 and this node won't be able to smith
			// the default ps is 100000, smithing could be slower than when using account balances
			// since default balance was 1000 times higher than default ps
			blocksmithScore, err = bp.BlockService.GetParticipationScore(bp.Generator.NodePublicKey)
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
			return blocker.NewBlocker(blocker.SmithingErr, fmt.Sprintf("undefined chaintype %s", bp.Chaintype.GetName()))
		}
		err = bp.BlocksmithService.CalculateSmith(
			lastBlock,
			*(blocksmithsMap[string(bp.Generator.NodePublicKey)]),
			bp.Generator,
			blocksmithScore,
		)
		if err != nil {
			return err
		}
		monitoring.SetBlockchainSmithTime(bp.Chaintype.GetTypeInt(), bp.Generator.SmithTime-lastBlock.Timestamp)
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

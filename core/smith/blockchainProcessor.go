package smith

import (
	"math"
	"math/big"
	"reflect"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/blocker"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	// BlockchainProcessorInterface represents interface for the blockchain processor's implementations
	BlockchainProcessorInterface interface {
		CalculateSmith(lastBlock *model.Block, generator *model.Blocksmith) *model.Blocksmith
		StartSmithing() error
		FakeSmithing(numberOfBlocks int, fromGenesis bool) error
	}

	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		Chaintype               chaintype.ChainType
		Generator               *model.Blocksmith
		BlockService            service.BlockServiceInterface
		NodeRegistrationService service.NodeRegistrationServiceInterface
		LastBlockID             int64
		canSmith                *bool
	}
)

// NewBlockchainProcessor create new instance of BlockchainProcessor
func NewBlockchainProcessor(
	ct chaintype.ChainType,
	blocksmith *model.Blocksmith,
	blockService service.BlockServiceInterface,
	nodeRegistrationService service.NodeRegistrationServiceInterface,
) *BlockchainProcessor {
	return &BlockchainProcessor{
		Chaintype:               ct,
		Generator:               blocksmith,
		BlockService:            blockService,
		NodeRegistrationService: nodeRegistrationService,
	}
}

// CalculateSmith calculate seed, smithTime, and Deadline
func (bp *BlockchainProcessor) CalculateSmith(lastBlock *model.Block, generator *model.Blocksmith) *model.Blocksmith {
	// try to get the node's participation score (ps) from node public key
	// if node is not registered, ps will be 0 and this node won't be able to smith
	// the default ps is 100000, smithing could be slower than when using account balances
	// since default balance was 1000 times higher than default ps
	ps, err := bp.BlockService.GetParticipationScore(generator.NodePublicKey)
	if ps == 0 {
		log.Info("Node has participation score = 0. Either is not registered or has been expelled from node registry")
	}
	if err != nil {
		log.Errorf("Participation score calculation: %s", err)
		generator.Score = big.NewInt(0)
	} else {
		generator.Score = big.NewInt(ps / int64(constant.ScalarReceiptScore))
	}
	if generator.Score.Sign() == 0 {
		generator.SmithTime = 0
		generator.BlockSeed = big.NewInt(0)
	}

	generator.BlockSeed, _ = coreUtil.GetBlockSeed(generator.NodePublicKey, lastBlock, generator.SecretPhrase)
	generator.SmithTime = coreUtil.GetSmithTime(generator.Score, generator.BlockSeed, lastBlock)
	generator.Deadline = uint32(math.Max(0, float64(generator.SmithTime-lastBlock.GetTimestamp())))
	return generator
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
		smithMax := timeNow - bp.Chaintype.GetChainSmithingDelayTime()
		if lastBlock.GetID() != bp.LastBlockID {
			bp.LastBlockID = lastBlock.GetID()
			bp.Generator = bp.CalculateSmith(lastBlock, bp.Generator)
		}
		// speed up the virtual time if smith time has not reach the needed smithing maximum time
		for bp.Generator.SmithTime > smithMax {
			timeNow++ // speed up bro
			smithMax = timeNow - bp.Chaintype.GetChainSmithingDelayTime()
		}
		// smith time reached
		timestamp := bp.Generator.GetTimestamp(smithMax)
		if !bp.BlockService.VerifySeed(bp.Generator.BlockSeed, bp.Generator.Score, lastBlock, timestamp) {
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
			timestamp,
		)
		if err != nil {
			return err
		}
		// validate
		err = bp.BlockService.ValidateBlock(block, previousBlock, timestamp) // err / !err
		if err != nil {
			return err
		}
		// if validated push
		err = bp.BlockService.PushBlock(previousBlock, block, true, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// StartSmithing start smithing loop
func (bp *BlockchainProcessor) StartSmithing() error {
	var blocksmithIndex = -1
	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}
	smithMax := time.Now().Unix() - bp.Chaintype.GetChainSmithingDelayTime()
	if lastBlock.GetID() != bp.LastBlockID {
		bp.LastBlockID = lastBlock.GetID()
		bp.BlockService.SortBlocksmiths(lastBlock)
		// check if eligible to create block in this round
		for i, bs := range *(bp.BlockService.GetSortedBlocksmiths()) {
			if reflect.DeepEqual(bs.NodePublicKey, bp.Generator.NodePublicKey) {
				blocksmithIndex = i
				break
			}
		}
		if blocksmithIndex < 0 {
			*(bp.canSmith) = false
			return blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
		}
		*(bp.canSmith) = true
		// if lastBlock.Timestamp > time.Now().Unix()-bp.Chaintype.GetChainSmithingDelayTime()*10 {
		// TODO: andy-shi88
		// pop off last block if has been absent for 10*delay
		// put processed transaction to process later
		// }
		// caching: only calculate smith time once per new block
		bp.Generator = bp.CalculateSmith(lastBlock, bp.Generator)
	}
	if !*(bp.canSmith) {
		return blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
	}
	if bp.Generator.SmithTime > smithMax {
		return nil
	}
	timestamp := bp.Generator.GetTimestamp(smithMax)
	if !bp.BlockService.VerifySeed(bp.Generator.BlockSeed, bp.Generator.Score, lastBlock, timestamp) {
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
		timestamp,
	)
	if err != nil {
		return err
	}
	// validate
	err = bp.BlockService.ValidateBlock(block, previousBlock, timestamp)
	if err != nil {
		return err
	}
	// if validated push
	err = bp.BlockService.PushBlock(previousBlock, block, true, true)
	if err != nil {
		return err
	}
	return nil
}

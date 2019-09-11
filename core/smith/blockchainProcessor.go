package smith

import (
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (

	// Blocksmith is wrapper for the account in smithing process
	Blocksmith struct {
		NodePublicKey  []byte
		AccountAddress string
		Score          *big.Int
		SmithTime      int64
		BlockSeed      *big.Int
		SecretPhrase   string
		deadline       uint32
	}

	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		Chaintype    chaintype.ChainType
		Generator    *Blocksmith
		BlockService service.BlockServiceInterface
		LastBlockID  int64
	}
)

// NewBlockchainProcessor create new instance of BlockchainProcessor
func NewBlockchainProcessor(
	ct chaintype.ChainType,
	blocksmith *Blocksmith,
	blockService service.BlockServiceInterface,
) *BlockchainProcessor {
	return &BlockchainProcessor{
		Chaintype:    ct,
		Generator:    blocksmith,
		BlockService: blockService,
	}
}

// InitGenerator initiate generator
func NewBlocksmith(nodeSecretPhrase, accountAddress string) *Blocksmith {
	blocksmith := &Blocksmith{
		AccountAddress: accountAddress,
		Score:          big.NewInt(constant.DefaultParticipationScore),
		SecretPhrase:   nodeSecretPhrase,
		NodePublicKey:  util.GetPublicKeyFromSeed(nodeSecretPhrase),
	}
	return blocksmith
}

// CalculateSmith calculate seed, smithTime, and deadline
func (bp *BlockchainProcessor) CalculateSmith(lastBlock *model.Block, generator *Blocksmith) *Blocksmith {
	//FIXME: implement logic based on participation score:
	// 1. get the node's participation score: from participation_score table, by generator.NodePublicKey or generator.AccountAddress
	// 2. use it for smithing: multiply by 1000 the value of ps, if we want to start with the same value below,
	//    since the default ps is 100000
	if generator.AccountAddress == "" {
		generator.Score = big.NewInt(0)
	} else {
		ps, err := bp.BlockService.GetParticipationScore(generator.AccountAddress)
		if err != nil {
			log.Errorf("Participation score calculation: %s", err)
			generator.Score = big.NewInt(0)
		} else {
			//TODO: default participation score is 1000 smaller than the old balance we used to use to smith
			//      so unless we tweak the GetSmithTime, mainchain could be slower now
			generator.Score = big.NewInt(int64(math.Max(0, float64(ps))))
		}
	}

	if generator.Score.Sign() == 0 {
		generator.SmithTime = 0
		generator.BlockSeed = big.NewInt(0)
	}
	generatorPublicKey, _ := util.GetPublicKeyFromAddress(generator.AccountAddress)
	generator.BlockSeed, _ = coreUtil.GetBlockSeed(generatorPublicKey, lastBlock)
	generator.SmithTime = coreUtil.GetSmithTime(generator.Score, generator.BlockSeed, lastBlock)
	generator.deadline = uint32(math.Max(0, float64(generator.SmithTime-lastBlock.GetTimestamp())))
	return generator
}

// StartSmithing start smithing loop
func (bp *BlockchainProcessor) StartSmithing() error {
	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return errors.New("Genesis:notAddedYet")
	}
	smithMax := time.Now().Unix() - bp.Chaintype.GetChainSmithingDelayTime()
	bp.Generator = bp.CalculateSmith(lastBlock, bp.Generator)
	if lastBlock.GetID() != bp.LastBlockID || bp.Generator.AccountAddress != "" {
		if bp.Generator.SmithTime > smithMax {
			log.Printf("skip forge\n")
			return errors.New("SmithSkip")
		}

		timestamp := bp.Generator.GetTimestamp(smithMax)
		if !bp.BlockService.VerifySeed(bp.Generator.BlockSeed, bp.Generator.Score, lastBlock, timestamp) {
			return errors.New("VerifySeed:false")
		}
		stop := false
		for {
			if stop {
				return nil
			}
			previousBlock, err := bp.BlockService.GetLastBlock()
			if err != nil {
				return err
			}

			block, err := bp.BlockService.GenerateBlock(
				previousBlock,
				bp.Generator.SecretPhrase,
				timestamp,
				bp.Generator.AccountAddress,
			)
			if err != nil {
				return err
			}
			// validate
			err = coreUtil.ValidateBlock(block, previousBlock, timestamp) // err / !err
			if err != nil {
				return err
			}
			// if validated push
			err = bp.BlockService.PushBlock(previousBlock, block, true)
			if err != nil {
				return err
			}
			stop = true
		}
	}
	return errors.New("GeneratorNotSet")
}

// GetTimestamp max timestamp allowed block to be smithed
func (blocksmith *Blocksmith) GetTimestamp(smithMax int64) int64 {
	elapsed := smithMax - blocksmith.SmithTime
	if elapsed > 3600 {
		return smithMax

	}
	return blocksmith.SmithTime + 1
}

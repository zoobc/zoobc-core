package smith

import (
	"errors"
	"math"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	// Blocksmith is wrapper for the account in smithing process
	Blocksmith struct {
		NodePublicKey []byte
		Score         *big.Int
		SmithTime     int64
		BlockSeed     *big.Int
		SecretPhrase  string
		deadline      uint32
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
		Score:         big.NewInt(constant.DefaultParticipationScore),
		SecretPhrase:  nodeSecretPhrase,
		NodePublicKey: util.GetPublicKeyFromSeed(nodeSecretPhrase),
	}
	return blocksmith
}

// CalculateSmith calculate seed, smithTime, and deadline
func (bp *BlockchainProcessor) CalculateSmith(lastBlock *model.Block, generator *Blocksmith) *Blocksmith {
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
		//TODO: default participation score is 10000 smaller than the old balance we used to use to smith
		//      so we multiply it by 10000 for now, to keep the same ratio
		generator.Score = big.NewInt(int64(math.Max(0, float64(ps*10000))))
	}
	if generator.Score.Sign() == 0 {
		generator.SmithTime = 0
		generator.BlockSeed = big.NewInt(0)
	}

	generator.BlockSeed, _ = coreUtil.GetBlockSeed(generator.NodePublicKey, lastBlock)
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
	if lastBlock.GetID() != bp.LastBlockID {
		if bp.Generator.SmithTime > smithMax {
			return errors.New("skip forge")
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
			err = bp.BlockService.PushBlock(previousBlock, block, true, true)
			if err != nil {
				log.Warn("pushBlock err ", block.Height, " ", err)
				return err
			}
			log.Printf("block forged: fee %d\n", block.TotalFee)
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

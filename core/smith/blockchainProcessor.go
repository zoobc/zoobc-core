package smith

import (
	"errors"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	// Blocksmith is wrapper for the account in smithing process
	Blocksmith struct {
		NodePublicKey  []byte
		AccountAddress string
		Balance        *big.Int
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
func NewBlocksmith(nodeSecretPhrase string) *Blocksmith {
	// todo: get node[private + public key] + look up account [public key, ID]
	blocksmith := &Blocksmith{
		AccountAddress: util.GetAddressFromSeed(nodeSecretPhrase),
		Balance:        big.NewInt(1000000000),
		SecretPhrase:   nodeSecretPhrase,
		NodePublicKey:  util.GetPublicKeyFromSeed(nodeSecretPhrase),
	}
	return blocksmith
}

// CalculateSmith calculate seed, smithTime, and deadline
func (*BlockchainProcessor) CalculateSmith(lastBlock *model.Block, generator *Blocksmith) *Blocksmith {
	account := model.AccountBalance{
		AccountAddress:   "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		Balance:          1000000000,
		SpendableBalance: 1000000000,
	}
	if account.AccountAddress == "" {
		generator.Balance = big.NewInt(0)
	} else {
		// FIXME: till we use POS to compute the smithing power, we should add to the account balance, the locked funds (in node_registry)
		accountEffectiveBalance := account.GetBalance()
		generator.Balance = big.NewInt(int64(math.Max(0, float64(accountEffectiveBalance))))
	}

	if generator.Balance.Sign() == 0 {
		generator.SmithTime = 0
		generator.BlockSeed = big.NewInt(0)
	}
	generatorPublicKey, _ := util.GetPublicKeyFromAddress(generator.AccountAddress)
	generator.BlockSeed, _ = coreUtil.GetBlockSeed(generatorPublicKey, lastBlock)
	generator.SmithTime = coreUtil.GetSmithTime(generator.Balance, generator.BlockSeed, lastBlock)
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
		if !bp.BlockService.VerifySeed(bp.Generator.BlockSeed, bp.Generator.Balance, lastBlock, timestamp) {
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

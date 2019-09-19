package smith

import (
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/observer"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		Chaintype               chaintype.ChainType
		Generator               *model.Blocksmith
		BlockService            service.BlockServiceInterface
		NodeRegistrationService service.NodeRegistrationServiceInterface
		LastBlockID             int64
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
		//TODO: default participation score is 10000 smaller than the old balance we used to use to smith
		//      so we multiply it by 10000 for now, to keep the same ratio
		generator.Score = big.NewInt(ps)
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

// StartSmithing start smithing loop
func (bp *BlockchainProcessor) StartSmithing() error {
	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}
	smithMax := time.Now().Unix() - bp.Chaintype.GetChainSmithingDelayTime()
	if lastBlock.GetID() != bp.LastBlockID {
		bp.LastBlockID = lastBlock.GetID()
		// if lastBlock.Timestamp > time.Now().Unix()-bp.Chaintype.GetChainSmithingDelayTime()*10 {
		// TODO: andy-shi88
		// pop off last block if has been absent for 10*delay
		// put processed transaction to process later
		// }
		// caching: only calculate smith time once per new block
		bp.Generator = bp.CalculateSmith(lastBlock, bp.Generator)
	}
	if bp.Generator.SmithTime > smithMax {
		return blocker.NewBlocker(
			blocker.SmithingErr, "skipping block creation")
	}
	timestamp := bp.Generator.GetTimestamp(smithMax)
	if !bp.BlockService.VerifySeed(bp.Generator.BlockSeed, bp.Generator.Score, lastBlock, timestamp) {
		return blocker.NewBlocker(
			blocker.SmithingErr, "verify seed return false")
	}
	stop := false
	for { // start creating block
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
		err = coreUtil.ValidateBlock(block, previousBlock, timestamp) // err / !err
		if err != nil {
			return err
		}
		// if validated push
		err = bp.BlockService.PushBlock(previousBlock, block, true, true)
		if err != nil {
			return err
		}
		stop = true
	}
	return blocker.NewBlocker(
		blocker.SmithingErr, "generator is not set")
}

func (bp *BlockchainProcessor) SortBlocksmith(sortedBlocksmiths *[]model.Blocksmith) observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			// fetch valid blocksmiths
			lastBlock := block.(*model.Block)
			var blocksmiths []model.Blocksmith
			activeBlocksmiths, err := bp.NodeRegistrationService.GetActiveNodes()
			if err != nil {
				return
			}
			for _, blocksmith := range activeBlocksmiths {
				if blocksmith.Score.Cmp(big.NewInt(0)) > 0 {
					blocksmiths = append(blocksmiths, *blocksmith)
				}
			}
			// sort blocksmiths
			sort.SliceStable(blocksmiths, func(i, j int) bool {
				bi, bj := blocksmiths[i], blocksmiths[j]
				nodePKI := new(big.Int).SetBytes(bi.NodePublicKey)
				nodePKJ := new(big.Int).SetBytes(bj.NodePublicKey)
				resI := new(big.Int).Mul(bi.Score, new(big.Int).SetInt64(
					nodePKI.Int64()^new(big.Int).SetBytes(lastBlock.BlockSeed).Int64()))
				resJ := new(big.Int).Mul(bj.Score, new(big.Int).SetInt64(
					nodePKJ.Int64()^new(big.Int).SetBytes(lastBlock.BlockSeed).Int64()))
				res := resI.Cmp(resJ)
				if res == 0 {
					// compare node public key
					res = nodePKI.Cmp(nodePKJ)
				}
				// ascending sort
				return res < 0
			})
			*sortedBlocksmiths = blocksmiths
		},
	}
}

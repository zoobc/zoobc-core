package smith

import (
	"errors"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/zoobc/zoobc-core/common/contract"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"
)

type (

	// Blocksmith is wrapper for the account in smithing process
	Blocksmith struct {
		NodePublicKey []byte
		AccountID     []byte
		Balance       *big.Int
		SmithTime     int64
		BlockSeed     *big.Int
		SecretPhrase  string
		deadline      uint32
	}

	// BlockchainProcessor handle smithing process, can be switch to process different chain by supplying different chain type
	BlockchainProcessor struct {
		Chaintype      contract.ChainType
		Generator      *Blocksmith
		BlockService   service.BlockServiceInterface
		MempoolService service.MempoolServiceInterface
		LastBlockID    int64
	}
)

// NewBlockchainProcessor create new instance of BlockchainProcessor
func NewBlockchainProcessor(ct contract.ChainType, blocksmith *Blocksmith,
	blockService service.BlockServiceInterface, mempoolService service.MempoolServiceInterface) *BlockchainProcessor {
	return &BlockchainProcessor{
		Chaintype:      ct,
		Generator:      blocksmith,
		BlockService:   blockService,
		MempoolService: mempoolService,
	}
}

// InitGenerator initiate generator
func NewBlocksmith(nodeSecretPhrase string) *Blocksmith {
	// todo: get node[private + public key] + look up account [public key, ID]
	blocksmith := &Blocksmith{
		AccountID: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139,
			255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		Balance:      big.NewInt(1000000000),
		SecretPhrase: nodeSecretPhrase,
		NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219,
			80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	}
	return blocksmith
}

// CalculateSmith calculate seed, smithTime, and deadline
func (*BlockchainProcessor) CalculateSmith(lastBlock *model.Block, generator *Blocksmith) *Blocksmith {
	account := model.AccountBalance{
		AccountID: []byte{4, 38, 113, 185, 80, 213, 37, 71, 68, 177, 176, 126, 241, 58, 3, 32, 129, 1, 156, 65, 199, 111,
			241, 130, 176, 116, 63, 35, 232, 241, 210, 172},
		Balance:          1000000000,
		SpendableBalance: 1000000000,
	}
	if len(account.AccountID) == 0 {
		generator.Balance = big.NewInt(0)
	} else {
		accountEffectiveBalance := account.GetBalance()
		generator.Balance = big.NewInt(int64(math.Max(0, float64(accountEffectiveBalance))))
	}

	if generator.Balance.Sign() == 0 {
		generator.SmithTime = 0
		generator.BlockSeed = big.NewInt(0)
	}
	generator.BlockSeed, _ = coreUtil.GetBlockSeed(generator.AccountID, lastBlock)
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
	if lastBlock.GetID() != bp.LastBlockID || bp.Generator.AccountID != nil {
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

			block, err := bp.GenerateBlock(previousBlock, bp.Generator.SecretPhrase, timestamp)
			if err != nil {
				return err
			}
			// validate
			err = coreUtil.ValidateBlock(block, previousBlock, timestamp) // err / !err
			if err != nil {
				return err
			}
			// if validated push
			err = bp.BlockService.PushBlock(previousBlock, block)
			if err != nil {
				return err
			}
			allBlocks, _ := bp.BlockService.GetBlocks()
			log.Printf("block pushed: %d\n", len(allBlocks))
			stop = true
		}
	}
	return errors.New("GeneratorNotSet")
}

// GenerateBlock generate block from transactions in mempool
func (bp *BlockchainProcessor) GenerateBlock(previousBlock *model.Block, secretPhrase string, timestamp int64) (*model.Block, error) {
	newBlockHeight := previousBlock.Height + 1
	digest := sha3.New512()
	var totalAmount int64
	var totalFee int64
	//TODO: missing coinbase calculation
	var totalCoinbase int64
	var payloadLength uint32

	// only for mainchain
	var sortedTx []*model.Transaction
	var payloadHash []byte

	if _, ok := bp.Chaintype.(*chaintype.MainChain); ok {
		mempoolTransactions, err := bp.MempoolService.SelectTransactionsFromMempool(timestamp)
		if err != nil {
			log.Println(err)
			return nil, errors.New("MempoolReadError")
		}
		digest := sha3.New512()
		for _, mpTx := range mempoolTransactions {
			tx, err := coreUtil.ParseTransactionBytes(mpTx.TransactionBytes, true)
			if err != nil {
				return nil, err
			}

			sortedTx = append(sortedTx, tx)
			_, _ = digest.Write(mpTx.TransactionBytes)
			txType := transaction.GetTransactionType(tx, nil)
			totalAmount += txType.GetAmount()
			totalFee += tx.Fee
			payloadLength += txType.GetSize()
		}
		payloadHash = digest.Sum([]byte{})
	}

	// loop through transaction to build block hash
	hash := digest.Sum([]byte{})
	digest.Reset() // reset the digest
	_, _ = digest.Write(previousBlock.GetBlockSeed())
	blocksmith := bp.Generator.AccountID
	_, _ = digest.Write(blocksmith)
	blockSeed := digest.Sum([]byte{})
	digest.Reset() // reset the digest
	previousBlockByte, _ := coreUtil.GetBlockByte(previousBlock, true)
	previousBlockHash := sha3.Sum512(previousBlockByte)
	block := bp.BlockService.NewBlock(1, previousBlockHash[:], blockSeed, blocksmith, string(hash), newBlockHeight, timestamp,
		totalAmount, totalFee, totalCoinbase, sortedTx, payloadHash, payloadLength, secretPhrase)
	log.Printf("block forged: fee %d\n", totalFee)
	return block, nil
}

// AddGenesis add genesis block of chain to the chain
func (bp *BlockchainProcessor) AddGenesis() error {
	digest := sha3.New512()
	var totalAmount int64
	var totalFee int64
	var totalCoinBase int64
	var payloadLength uint32
	var blockTransactions []*model.Transaction
	for _, tx := range service.GetGenesisTransactions(bp.Chaintype) {
		txBytes, _ := coreUtil.GetTransactionBytes(tx, true)
		_, _ = digest.Write(txBytes)
		if tx.TransactionType == util.ConvertBytesToUint32([]byte{1, 0, 0, 0}) { // if type = send money
			totalAmount += tx.GetSendMoneyTransactionBody().Amount
		}
		txType := transaction.GetTransactionType(tx, nil)
		totalAmount += txType.GetAmount()
		totalFee += tx.Fee
		payloadLength += txType.GetSize()
		blockTransactions = append(blockTransactions, tx)
	}
	payloadHash := digest.Sum([]byte{})
	block := bp.BlockService.NewGenesisBlock(1, nil, make([]byte, 64), []byte{}, "",
		0, constant.GenesisBlockTimestamp, totalAmount, totalFee, totalCoinBase, blockTransactions, payloadHash, payloadLength,
		constant.InitialSmithScale, big.NewInt(0), constant.GenesisBlockSignature)
	// assign genesis block id
	block.ID = coreUtil.GetBlockID(block)
	err := bp.BlockService.PushBlock(&model.Block{ID: -1, Height: 0}, block)
	if err != nil {
		panic("PushGenesisBlock:fail")
	}
	return nil
}

// CheckGenesis check if genesis has been added
func (bp *BlockchainProcessor) CheckGenesis() bool {
	genesisBlock, err := bp.BlockService.GetGenesisBlock()
	if err != nil { // Genesis is not in the blockchain yet
		return false
	}
	if genesisBlock.ID != bp.Chaintype.GetGenesisBlockID() {
		log.Fatalf("Genesis ID does not match, expect: %d, get: %d", bp.Chaintype.GetGenesisBlockID(), genesisBlock.ID)
	}
	return true
}

// GetTimestamp max timestamp allowed block to be smithed
func (blocksmith *Blocksmith) GetTimestamp(smithMax int64) int64 {
	elapsed := smithMax - blocksmith.SmithTime
	if elapsed > 3600 {
		return smithMax

	}
	return blocksmith.SmithTime + 1
}

package smith

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/core/util"
)

var mockBlocksmith = Blocksmith{
	AccountID: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
		81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
	Balance:      big.NewInt(1000000000),
	SecretPhrase: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
	NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80,
		242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
}

// Mockchain
type mockChain struct {
	chaintype.MainChain
}

func (*mockChain) GetChainSmithingDelayTime() int64 { return 10 }
func (*mockChain) GetGenesisBlockID() int64         { return 1 }

// BlockService mock success
type mockMempoolServiceSuccess struct {
	service.MempoolService
}

func (*mockMempoolServiceSuccess) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error) {
	mempoolTransactions := []*model.MempoolTransaction{
		{
			ID:               1,
			FeePerByte:       1,
			ArrivalTimestamp: 1562893305,
			TransactionBytes: getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
		},
		{
			ID:               2,
			FeePerByte:       10,
			ArrivalTimestamp: 1562893304,
			TransactionBytes: getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
		},
		{
			ID:               3,
			FeePerByte:       1,
			ArrivalTimestamp: 1562893302,
			TransactionBytes: getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
		},
	}
	return mempoolTransactions, nil
}

func getTestSignedMempoolTransaction(id, timestamp int64) *model.MempoolTransaction {
	tx := buildTransaction(id, timestamp, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN")
	txBytes, _ := util.GetTransactionBytes(tx, true)
	return &model.MempoolTransaction{
		ID:               id,
		FeePerByte:       1,
		ArrivalTimestamp: timestamp,
		TransactionBytes: txBytes,
	}
}

func buildTransaction(id, timestamp int64, sender, recipient string) *model.Transaction {
	return &model.Transaction{
		Version:                 1,
		ID:                      id,
		BlockID:                 1,
		Height:                  1,
		SenderAccountType:       0,
		SenderAccountAddress:    sender,
		RecipientAccountType:    0,
		RecipientAccountAddress: recipient,
		TransactionType:         0,
		Fee:                     1,
		Timestamp:               timestamp,
		TransactionHash:         make([]byte, 32),
		TransactionBodyLength:   0,
		TransactionBodyBytes:    make([]byte, 0),
		TransactionBody:         nil,
		Signature:               make([]byte, 64),
	}
}

type mockBlockServiceSuccess struct {
	service.BlockService
}

func (*mockBlockServiceSuccess) VerifySeed(seed, balance *big.Int, previousBlock *model.Block, timestamp int64) bool {
	return true
}
func (*mockBlockServiceSuccess) NewBlock(version uint32, previousBlockHash, blockSeed, blocksmithID []byte, hash string,
	previousBlockHeight uint32, timestamp, totalAmount, totalFee, totalCoinBase int64, transactions []*model.Transaction,
	payloadHash []byte, secretPhrase string) *model.Block {
	return &model.Block{
		Version:           1,
		PreviousBlockHash: []byte{},
		BlockSeed:         []byte{},
		BlocksmithID:      []byte{},
		Timestamp:         15875392,
		TotalAmount:       0,
		TotalFee:          0,
		TotalCoinBase:     0,
		Transactions:      []*model.Transaction{},
		PayloadHash:       []byte{},
	}
}
func (*mockBlockServiceSuccess) NewGenesisBlock(version uint32, previousBlockHash, blockSeed, blocksmithID []byte,
	hash string, previousBlockHeight uint32, timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction, payloadHash []byte, smithScale int64, cumulativeDifficulty *big.Int,
	genesisSignature []byte) *model.Block {
	return &model.Block{
		Version:              1,
		PreviousBlockHash:    []byte{},
		BlockSeed:            []byte{},
		BlocksmithID:         []byte{},
		Timestamp:            15875392,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Transactions:         []*model.Transaction{},
		PayloadHash:          []byte{},
		SmithScale:           0,
		CumulativeDifficulty: "1",
		BlockSignature:       []byte{},
	}
}

func (*mockBlockServiceSuccess) PushBlock(previousBlock, block *model.Block) error { return nil }

func (*mockBlockServiceSuccess) GetGenesisBlock() (*model.Block, error) {
	return &model.Block{
		ID:                   1,
		Version:              1,
		PreviousBlockHash:    []byte{},
		BlockSeed:            []byte{},
		BlocksmithID:         []byte{},
		Timestamp:            15875392,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Transactions:         []*model.Transaction{},
		PayloadHash:          []byte{},
		SmithScale:           0,
		CumulativeDifficulty: "1",
		BlockSignature:       []byte{},
	}, nil
}

// BlockService mock fail
type mockBlockServiceFail struct {
	mockBlockServiceSuccess
}

func (*mockBlockServiceFail) GetGenesisBlock() (*model.Block, error) {
	return &model.Block{}, errors.New("mockError")
}

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		chaintype      contract.ChainType
		blocksmith     *Blocksmith
		blockService   service.BlockServiceInterface
		mempoolService service.MempoolServiceInterface
	}
	test := struct {
		name string
		args args
		want *BlockchainProcessor
	}{
		name: "NewBlockchainProcessor:success",
		args: args{
			chaintype:      &chaintype.MainChain{},
			blocksmith:     &Blocksmith{},
			blockService:   nil,
			mempoolService: nil,
		},
		want: &BlockchainProcessor{
			Chaintype:      &chaintype.MainChain{},
			Generator:      &Blocksmith{},
			BlockService:   nil,
			MempoolService: nil,
			LastBlockID:    0,
		},
	}
	if got := NewBlockchainProcessor(test.args.chaintype, test.args.blocksmith, test.args.blockService,
		test.args.mempoolService); !reflect.DeepEqual(got, test.want) {
		t.Errorf("NewBlockchainProcessor() = %v, want %v", got, test.want)
	}
}

func TestNewBlocksmith(t *testing.T) {
	type args struct {
		secretPhrase string
	}
	test := struct {
		name string
		args args
		want *Blocksmith
	}{
		name: "NewBlocksmith:success",
		args: args{
			secretPhrase: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
		},
		want: &mockBlocksmith,
	}
	if got := NewBlocksmith(test.args.secretPhrase); !reflect.DeepEqual(got, test.want) {
		t.Errorf("NewBlocksmith() = %v, want %v", got, test.want)
	}
}

func TestBlockchainProcessor_CalculateSmith(t *testing.T) { // todo: test can be written once account_balance is integrated.
	type fields struct {
		Chaintype    contract.ChainType
		Generator    *Blocksmith
		BlockService service.BlockServiceInterface
		LastBlockID  int64
	}
	type args struct {
		lastBlock *model.Block
		generator *Blocksmith
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Blocksmith
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockchainProcessor{
				Chaintype:    tt.fields.Chaintype,
				Generator:    tt.fields.Generator,
				BlockService: tt.fields.BlockService,
				LastBlockID:  tt.fields.LastBlockID,
			}
			if got := b.CalculateSmith(tt.args.lastBlock, tt.args.generator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainProcessor.CalculateSmith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainProcessor_GenerateBlock(t *testing.T) { //todo: update test when transaction and pop implemented.
	type fields struct {
		Chaintype      contract.ChainType
		Generator      *Blocksmith
		BlockService   service.BlockServiceInterface
		MempoolService service.MempoolServiceInterface
		LastBlockID    int64
	}
	type args struct {
		previousBlock       *model.Block
		secretPhrase        string
		timestamp           int64
		mempoolTransactions []*model.MempoolTransaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GenerateBlock:success-{}",
			fields: fields{
				Chaintype:      &chaintype.MainChain{},
				Generator:      &Blocksmith{},
				BlockService:   &mockBlockServiceSuccess{},
				MempoolService: &mockMempoolServiceSuccess{},
				LastBlockID:    0,
			},
			args: args{
				previousBlock:       &model.Block{},
				secretPhrase:        "",
				timestamp:           1562585975339,
				mempoolTransactions: []*model.MempoolTransaction{},
			},
			want: &model.Block{
				Version:           1,
				PreviousBlockHash: []byte{},
				BlockSeed:         []byte{},
				BlocksmithID:      []byte{},
				Timestamp:         15875392,
				TotalAmount:       0,
				TotalFee:          0,
				TotalCoinBase:     0,
				Transactions:      []*model.Transaction{},
				PayloadHash:       []byte{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockchainProcessor{
				Chaintype:      tt.fields.Chaintype,
				Generator:      tt.fields.Generator,
				BlockService:   tt.fields.BlockService,
				MempoolService: tt.fields.MempoolService,
				LastBlockID:    tt.fields.LastBlockID,
			}
			got, err := bp.GenerateBlock(tt.args.previousBlock, tt.args.secretPhrase, tt.args.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockchainProcessor.GenerateBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainProcessor.GenerateBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainProcessor_AddGenesis(t *testing.T) { //todo: update test when genesis transaction added
	type fields struct {
		Chaintype    contract.ChainType
		Generator    *Blocksmith
		BlockService service.BlockServiceInterface
		LastBlockID  int64
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "AddGenesis:success-{BlockService.PushBlock:success}",
			fields: fields{
				Chaintype:    &mockChain{},
				Generator:    &Blocksmith{},
				BlockService: &mockBlockServiceSuccess{},
				LastBlockID:  0,
			},
			wantErr: false,
		},
		{
			name: "AddGenesis:success-{BlockService.PushBlock:fail}",
			fields: fields{
				Chaintype:    &mockChain{},
				Generator:    &Blocksmith{},
				BlockService: &mockBlockServiceFail{},
				LastBlockID:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockchainProcessor{
				Chaintype:    tt.fields.Chaintype,
				Generator:    tt.fields.Generator,
				BlockService: tt.fields.BlockService,
				LastBlockID:  tt.fields.LastBlockID,
			}
			if err := bp.AddGenesis(); (err != nil) != tt.wantErr {
				t.Errorf("BlockchainProcessor.AddGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchainProcessor_CheckGenesis(t *testing.T) {
	type fields struct {
		Chaintype    contract.ChainType
		Generator    *Blocksmith
		BlockService service.BlockServiceInterface
		LastBlockID  int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "CheckGenesis:success-{}",
			fields: fields{
				Chaintype:    &mockChain{},
				Generator:    &mockBlocksmith,
				BlockService: &mockBlockServiceSuccess{},
				LastBlockID:  0,
			},
			want: true,
		},
		{
			name: "CheckGenesis:fail-{}",
			fields: fields{
				Chaintype:    &mockChain{},
				Generator:    &mockBlocksmith,
				BlockService: &mockBlockServiceFail{},
				LastBlockID:  0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockchainProcessor{
				Chaintype:    tt.fields.Chaintype,
				Generator:    tt.fields.Generator,
				BlockService: tt.fields.BlockService,
				LastBlockID:  tt.fields.LastBlockID,
			}
			if got := bp.CheckGenesis(); got != tt.want {
				t.Errorf("BlockchainProcessor.CheckGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmith_GetTimestamp(t *testing.T) {
	type fields struct {
		NodePublicKey    []byte
		AccountPublicKey []byte
		Balance          *big.Int
		SmithTime        int64
		BlockSeed        *big.Int
		SecretPhrase     string
		deadline         uint32
	}
	type args struct {
		smithMax int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "GetTimestamp:success-{elapsed<=3600}",
			fields: fields{
				AccountPublicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
					81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				Balance:      big.NewInt(1000000000),
				SecretPhrase: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
				NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80,
					242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				SmithTime: 10000,
			},
			args: args{
				smithMax: 13300,
			},
			want: 10001, // return blocksmith smithTime+1 if elapsed < 3600
		},
		{
			name: "GetTimestamp:success-{elapsed=3600}",
			fields: fields{
				AccountPublicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81,
					229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				Balance:      big.NewInt(1000000000),
				SecretPhrase: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
				NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80,
					242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				SmithTime: 10000,
			},
			args: args{
				smithMax: 13600,
			},
			want: 10001, // return blocksmith smithTime+1 if elapsed < 3600
		},
		{
			name: "GetTimestamp:success-{elapsed>3600}",
			fields: fields{
				AccountPublicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81,
					229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				Balance:      big.NewInt(1000000000),
				SecretPhrase: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
				NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242,
					244, 100, 134, 144, 246, 37, 144, 213, 135},
				SmithTime: 10000,
			},
			args: args{
				smithMax: 14000,
			},
			want: 14000, // return smithMax if elapsed > 3600
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocksmith := &Blocksmith{
				NodePublicKey: tt.fields.NodePublicKey,
				AccountID:     tt.fields.AccountPublicKey,
				Balance:       tt.fields.Balance,
				SmithTime:     tt.fields.SmithTime,
				BlockSeed:     tt.fields.BlockSeed,
				SecretPhrase:  tt.fields.SecretPhrase,
				deadline:      tt.fields.deadline,
			}
			if got := blocksmith.GetTimestamp(tt.args.smithMax); got != tt.want {
				t.Errorf("Blocksmith.GetTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

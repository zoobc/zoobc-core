package blockchainsync

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"
)

// ALL SERVICE SUCCESS
type mockServiceBlockSuccess struct {
	service.BlockServiceInterface
}

type mockServiceChainType struct {
	chaintype.ChainType
}

type mockServiceQueryExecutor struct {
	query.ExecutorInterface
}

// BLOCK SERVICE FAILS
type mockServiceBlockFailGetLastBlock struct {
	service.BlockServiceInterface
}

type mockServiceBlockFailGetBlockByHeight struct {
	service.BlockServiceInterface
}

type mockServiceBlockFailGetBlockByID struct {
	service.BlockServiceInterface
}

type mockServiceQueryExecutorCommitTXFail struct {
	query.ExecutorInterface
}

// Mock function for Block interface
func (*mockServiceBlockSuccess) GetLastBlock() (*model.Block, error) {
	return &model.Block{ID: 58, Height: 66}, nil
}

func (*mockServiceBlockSuccess) GetBlockByHeight(height uint32) (*model.Block, error) {
	if height == 0 {
		return &model.Block{ID: 1, Height: height}, nil // genesis
	}
	return &model.Block{ID: 58, Height: height}, nil
}

func (*mockServiceBlockSuccess) GetBlockByID(int64) (*model.Block, error) {
	return &model.Block{ID: 40, Height: 69}, nil
}

func (*mockServiceBlockSuccess) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	transaction := []*model.Transaction{
		{Version: 1,
			ID:      789,
			BlockID: 40,
			Height:  69,
		},
	}
	return transaction, nil
}

// Mock Function for Chaintype Interface
func (*mockServiceChainType) GetGenesisBlockID() int64 {
	return 1
}

func (*mockServiceChainType) GetName() string {
	return "Mainchain"
}

func (*mockServiceChainType) GetChainSmithingDelayTime() int64 {
	return 60
}

func (*mockServiceChainType) GetTablePrefix() string {
	return "main"
}

func (*mockServiceChainType) GetTypeInt() int32 {
	return 0
}

// Mock Function for Query Executor Interface
func (*mockServiceQueryExecutor) BeginTx() error {
	return nil
}

func (*mockServiceQueryExecutor) ExecuteTransactions(qStr [][]interface{}) error {
	return nil
}

func (*mockServiceQueryExecutor) CommitTx() error {
	return nil
}

func (*mockServiceQueryExecutor) RollbackTx() error {
	return nil
}

// MOCK BLOCK SERVICES FUNCTION FAILS
// Mock Function Block Services GetLastBlock Fail
func (*mockServiceBlockFailGetLastBlock) GetLastBlock() (*model.Block, error) {
	return nil, blocker.NewBlocker(
		blocker.AuthErr,
		"error in getting LAST BLOCK",
	)
}

func (*mockServiceBlockFailGetLastBlock) GetBlockByHeight(height uint32) (*model.Block, error) {
	return &model.Block{ID: 57, Height: height}, nil
}

func (*mockServiceBlockFailGetLastBlock) GetBlockByID(int64) (*model.Block, error) {
	return &model.Block{ID: 40, Height: 69}, nil
}

func (*mockServiceBlockFailGetLastBlock) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	transaction := []*model.Transaction{
		{
			Version: 1,
			ID:      789,
			BlockID: 40,
			Height:  69,
		},
	}
	return transaction, nil
}

// Mock Function Block Service Fail GetBLockByHeight Fail
func (*mockServiceBlockFailGetBlockByHeight) GetLastBlock() (*model.Block, error) {
	return &model.Block{ID: 58, Height: 66}, nil
}

func (*mockServiceBlockFailGetBlockByHeight) GetBlockByHeight(height uint32) (*model.Block, error) {
	return nil, blocker.NewBlocker(
		blocker.BlockNotFoundErr,
		"ERROR WHEN GETTING BLOCK USING HEIGHT",
	)
}

func (*mockServiceBlockFailGetBlockByHeight) GetBlockByID(int64) (*model.Block, error) {
	return &model.Block{ID: 40, Height: 69}, nil
}

func (*mockServiceBlockFailGetBlockByHeight) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	transaction := []*model.Transaction{
		{Version: 1,
			ID:      789,
			BlockID: 40,
			Height:  69,
		},
	}
	return transaction, nil
}

// Mock Function Block Service GetBlockByID fail
func (*mockServiceBlockFailGetBlockByID) GetLastBlock() (*model.Block, error) {
	return &model.Block{ID: 58, Height: 800}, nil
}

func (*mockServiceBlockFailGetBlockByID) GetBlockByHeight(height uint32) (*model.Block, error) {
	return &model.Block{ID: 58, Height: height}, nil
}

func (*mockServiceBlockFailGetBlockByID) GetBlockByID(int64) (*model.Block, error) {
	return nil, blocker.NewBlocker(
		blocker.AuthErr,
		"ERROR WHEN GETTING BLOCK USING ID",
	)
}

func (*mockServiceBlockFailGetBlockByID) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	transaction := []*model.Transaction{
		{Version: 1,
			ID:      789,
			BlockID: 40,
			Height:  69,
		},
	}
	return transaction, nil
}

// COMMITX FAIL
func (*mockServiceQueryExecutorCommitTXFail) BeginTx() error {
	return nil
}

func (*mockServiceQueryExecutorCommitTXFail) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockServiceQueryExecutorCommitTXFail) CommitTx() error {
	return blocker.NewBlocker(
		blocker.AuthErr,
		"failed to commit TX",
	)
}

func (*mockServiceQueryExecutorCommitTXFail) RollbackTx() error {
	return nil
}

func TestService_ProcessFork(t *testing.T) {
	type fields struct {
		NeedGetMoreBlocks          bool
		IsDownloading              bool
		LastBlockchainFeeder       *model.Peer
		LastBlockchainFeederHeight uint32
		PeerHasMore                bool
		ChainType                  chaintype.ChainType
		BlockService               service.BlockServiceInterface
		LastBlock                  model.Block
		TransactionQuery           query.TransactionQueryInterface
		ForkingProcess             ForkingProcessorInterface
		QueryExecutor              query.ExecutorInterface
		BlockQuery                 query.BlockQueryInterface
	}
	type args struct {
		forkBlocks  []*model.Block
		commonBlock *model.Block
		feederPeer  *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &ForkingProcessor{
				ChainType:    tt.fields.ChainType,
				BlockService: tt.fields.BlockService,
			}
			if err := fp.ProcessFork(tt.args.forkBlocks, tt.args.commonBlock, tt.args.feederPeer); (err != nil) != tt.wantErr {
				t.Errorf("Service.ProcessFork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package service

import (
	"errors"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
	}

	// NodeAdminService contains all transactions in mempool plus a mux to manage locks in concurrency
	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Sign          crypto.SignatureInterface
	}
)

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership() error {
	nas.LookupLastBlock()
	nas.LookupOwnerAccount()
	nas.SignData()
}

// GetLastBlock return the last pushed block
func (nas *NodeAdminService) LookupLastBlock() (*model.Block, error) {
	rows, err := nas.QueryExecutor.ExecuteSelect(nas.BlockQuery.GetLastBlock())
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return &model.Block{
			ID: -1,
		}, err
	}
	var blocks []*model.Block
	blocks = nas.BlockQuery.BuildModel(blocks, rows)
	if len(blocks) > 0 {
		return blocks[0], nil
	}
	return &model.Block{
		ID: -1,
	}, errors.New("BlockNotFound")

}

func (nas *NodeAdminService) LookupOwnerAccount(accountID []byte) (*model.Account, error){
	rows, err := nas.QueryExecutor.ExecuteSelect(nas.AccountQuery.GetAccountByID(accountID []byte))
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return &model.Account{
			ID: -1,
		}, err
	}
	var account []*model.Account
	account = nas.AccountQuery.BuildModel(account, rows)
	if len(account) > 0 {
		return account[0], nil
	}
	return &model.Account{
		ID: -1,
	}, errors.New("Account Not Found")
}

func (nas *NodeAdminService) SignData() error {
	
}

// validate proof of ownership
func (nas *NodeAdminService) ValidateProofOfOwnership(mpTx *model.MempoolTransaction) error {

}
func (nas *NodeAdminService) ValidateSignature() error {

}
func (nas *NodeAdminService) ValidateHeight() error {

}
func (nas *NodeAdminService) LookupBlock() error {

}

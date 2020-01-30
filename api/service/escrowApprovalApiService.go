package service

import (
	"database/sql"
	"errors"

	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	EscrowApprovalService struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Observer           *observer.Observer
	}
)

var escrowApprovalServiceInstance *EscrowApprovalService

func NewEscrowApprovalService(
	queryExecutor query.ExecutorInterface,
	signature crypto.SignatureInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	mempoolService service.MempoolServiceInterface,
	observer *observer.Observer,
) *EscrowApprovalService {
	if escrowApprovalServiceInstance == nil {
		escrowApprovalServiceInstance = &EscrowApprovalService{
			Query:              queryExecutor,
			Signature:          signature,
			ActionTypeSwitcher: txTypeSwitcher,
			MempoolService:     mempoolService,
			Observer:           observer,
		}
	}
	return escrowApprovalServiceInstance
}

/*
PostApprovalEscrowTransaction represents POST request method approval escrow transaction
*/
func (eas *EscrowApprovalService) PostApprovalEscrowTransaction(
	request *model.PostEscrowApprovalRequest,
) (*model.Transaction, error) {
	var (
		approval           model.EscrowApproval
		txType             transaction.TypeAction
		escrow, nextEscrow model.Escrow
		err                error
		id                 []byte
		tx                 model.Transaction
		caseQuery          = query.NewCaseQuery()
		escrowQuery        = query.NewEscrowTransactionQuery()
		row                *sql.Row
	)

	approval, id, err = transaction.ParseEscrowApprovalBytes(request.GetApprovalBytes())
	if err != nil {
		return nil, err
	}

	escrowQ, escrowArgs := escrowQuery.GetLatestEscrowTransactionByID(id)
	row, err = eas.Query.ExecuteSelectRow(escrowQ, false, escrowArgs...)
	if err != nil {
		return nil, err
	}
	err = escrowQuery.Scan(&escrow, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.New("transaction not found")
	}
	nextEscrow = escrow

	if escrow.GetID() != int64(util.ConvertBytesToUint64(id)) {
		return nil, errors.New("transaction id not match")
	}
	switch approval {
	case model.EscrowApproval_Approve:
		nextEscrow.Status = model.EscrowStatus_Approved
	case model.EscrowApproval_Reject:
		nextEscrow.Status = model.EscrowStatus_Rejected
	}
	escrow.Latest = false
	nextEscrow.Latest = true

	return nil, nil
}

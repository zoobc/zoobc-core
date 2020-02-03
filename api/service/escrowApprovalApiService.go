package service

import (
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// EscrowApprovalService fields that needed for EscrowApproval
	EscrowApprovalService struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Observer           *observer.Observer
	}
	EscrowApprovalServiceInterface interface {
		PostApprovalEscrowTransaction(
			chainType chaintype.ChainType,
			request *model.PostEscrowApprovalRequest,
		) (*model.Transaction, error)
	}
)

var escrowApprovalServiceInstance *EscrowApprovalService

// NewEscrowApprovalService build and return an EscrowApprovalService instance
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
	chainType chaintype.ChainType,
	request *model.PostEscrowApprovalRequest,
) (*model.Transaction, error) {
	var (
		txBytes = request.GetApprovalBytes()
		txType  transaction.TypeAction
		err     error
		tx      *model.Transaction
	)

	tx, err = transaction.ParseTransactionBytes(txBytes, true)
	if err != nil {
		return nil, err
	}
	// Validate Tx
	txType, err = eas.ActionTypeSwitcher.GetTransactionType(tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Save to mempool
	mpTx := &model.MempoolTransaction{
		FeePerByte:              util.FeePerByteTransaction(tx.GetFee(), txBytes),
		ID:                      tx.GetID(),
		TransactionBytes:        txBytes,
		ArrivalTimestamp:        time.Now().Unix(),
		SenderAccountAddress:    tx.GetSenderAccountAddress(),
		RecipientAccountAddress: tx.GetRecipientAccountAddress(),
	}

	if errValidate := eas.MempoolService.ValidateMempoolTransaction(mpTx); errValidate != nil {
		return nil, status.Error(codes.Internal, errValidate.Error())
	}

	err = eas.Query.BeginTx()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: repetitive way
	escrowable, ok := txType.Escrowable()
	switch ok {
	case true:
		err = escrowable.EscrowApplyUnconfirmed()
	default:
		err = txType.ApplyUnconfirmed()
	}

	if err != nil {
		errRollback := eas.Query.RollbackTx()
		if errRollback != nil {
			return nil, status.Error(codes.Internal, errRollback.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = eas.MempoolService.AddMempoolTransaction(mpTx)
	if err != nil {
		errRollback := eas.Query.RollbackTx()
		if errRollback != nil {
			return nil, status.Error(codes.Internal, errRollback.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = eas.Query.CommitTx()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	eas.Observer.Notify(observer.TransactionAdded, mpTx.GetTransactionBytes(), chainType)
	return tx, nil
}

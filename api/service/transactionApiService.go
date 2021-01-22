// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package service

import (
	"database/sql"
	"math"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/feedbacksystem"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	TransactionServiceInterface interface {
		GetTransaction(chaintype.ChainType, *model.GetTransactionRequest) (*model.Transaction, error)
		GetTransactions(chaintype.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error)
		PostTransaction(chaintype.ChainType, *model.PostTransactionRequest) (*model.Transaction, error)
		GetTransactionMinimumFee(request *model.GetTransactionMinimumFeeRequest) (
			*model.GetTransactionMinimumFeeResponse, error,
		)
	}

	// TransactionService represents struct of TransactionService
	TransactionService struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Observer           *observer.Observer
		TransactionUtil    transaction.UtilInterface
		FeedbackStrategy   feedbacksystem.FeedbackStrategyInterface
		Logger             *log.Logger
	}
)

var transactionServiceInstance *TransactionService

// NewTransactionService creates a singleton instance of TransactionService
func NewTransactionService(
	queryExecutor query.ExecutorInterface,
	signature crypto.SignatureInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	mempoolService service.MempoolServiceInterface,
	observer *observer.Observer,
	transactionUtil transaction.UtilInterface,
	feedbackStrategy feedbacksystem.FeedbackStrategyInterface,
	logger *log.Logger,
) *TransactionService {
	if transactionServiceInstance == nil {
		transactionServiceInstance = &TransactionService{
			Query:              queryExecutor,
			Signature:          signature,
			ActionTypeSwitcher: txTypeSwitcher,
			MempoolService:     mempoolService,
			Observer:           observer,
			TransactionUtil:    transactionUtil,
			FeedbackStrategy:   feedbackStrategy,
			Logger:             logger,
		}
	}
	return transactionServiceInstance
}

// GetTransaction fetches a single transaction from DB
func (ts *TransactionService) GetTransaction(
	chainType chaintype.ChainType,
	params *model.GetTransactionRequest,
) (*model.Transaction, error) {
	var (
		err error
		row *sql.Row
		tx  model.Transaction
	)

	txQuery := query.NewTransactionQuery(chainType)
	row, _ = ts.Query.ExecuteSelectRow(txQuery.GetTransaction(params.GetID()), false)
	err = txQuery.Scan(&tx, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.NotFound, err.Error())
	}
	txType, err := ts.ActionTypeSwitcher.GetTransactionType(&tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	txType.GetTransactionBody(&tx)
	return &tx, nil
}

// GetTransactions fetches a single transaction from DB
// included filters
func (ts *TransactionService) GetTransactions(
	chainType chaintype.ChainType,
	params *model.GetTransactionsRequest,
) (*model.GetTransactionsResponse, error) {
	var (
		err          error
		rowCount     *sql.Row
		rows2        *sql.Rows
		txs          []*model.Transaction
		selectQuery  string
		args         []interface{}
		totalRecords uint64
		txQuery      = query.NewTransactionQuery(chainType)
		caseQuery    = query.NewCaseQuery()
		// Represent transaction fields
		txFields = map[string]string{
			"Height":  "block_height",
			"BlockID": "block_id",
		}
	)
	caseQuery.Select(txQuery.TableName, txQuery.Fields...)

	page := params.GetPagination()
	height := params.GetHeight()
	if height != 0 {
		caseQuery.Where(caseQuery.Equal("block_height", height))
		if page != nil && page.GetLimit() == 0 {
			page.Limit = math.MaxUint32
		}
	}

	timestampStart := params.GetTimestampStart()
	timestampEnd := params.GetTimestampEnd()
	if timestampStart > 0 {
		caseQuery.And(caseQuery.Between("timestamp", timestampStart, timestampEnd))
	}

	transactionType := params.GetTransactionType()
	if transactionType > 0 {
		caseQuery.And(caseQuery.Equal("transaction_type", transactionType))
	}

	accountAddress := params.GetAccountAddress()
	if accountAddress != nil {
		caseQuery.AndOr(
			caseQuery.Equal("sender_account_address", accountAddress),
			caseQuery.Equal("recipient_account_address", accountAddress),
		)
	}
	selectQuery, args = caseQuery.Build()

	fromBlock := params.GetFromBlock()
	toBlock := params.GetToBlock()

	if fromBlock > toBlock {
		return nil, status.Error(codes.Internal, "FromBlock height cannot exceed toBlock height")
	}

	if toBlock != 0 && fromBlock <= toBlock {
		caseQuery.And(caseQuery.Between("height", fromBlock, toBlock))
	}

	// count first
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	rowCount, err = ts.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = rowCount.Scan(
		&totalRecords,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get Transactions with Pagination
	if page.GetOrderField() == "" || txFields[page.GetOrderField()] == "" {
		caseQuery.OrderBy("timestamp", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(txFields[page.GetOrderField()], page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())
	selectQuery, args = caseQuery.Build()

	rows2, err = ts.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows2.Close()

	for rows2.Next() {
		var tx model.Transaction
		err = rows2.Scan(
			&tx.ID,
			&tx.BlockID,
			&tx.Height,
			&tx.SenderAccountAddress,
			&tx.RecipientAccountAddress,
			&tx.TransactionType,
			&tx.Fee,
			&tx.Timestamp,
			&tx.TransactionHash,
			&tx.TransactionBodyLength,
			&tx.TransactionBodyBytes,
			&tx.Signature,
			&tx.Version,
			&tx.TransactionIndex,
			&tx.MultisigChild,
			&tx.Message,
		)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, status.Error(codes.Internal, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		txType, err := ts.ActionTypeSwitcher.GetTransactionType(&tx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		txType.GetTransactionBody(&tx)
		txs = append(txs, &tx)
	}

	return &model.GetTransactionsResponse{
		Total:        totalRecords,
		Transactions: txs,
	}, nil
}

// PostTransaction represents POST transaction method
func (ts *TransactionService) PostTransaction(
	chaintype chaintype.ChainType,
	req *model.PostTransactionRequest,
) (*model.Transaction, error) {
	var (
		txBytes = req.GetTransactionBytes()
		txType  transaction.TypeAction
		tx      *model.Transaction
		err     error
		tpsProcessed,
		tpsReceived int
		isDbTransactionHighPriority = false
	)

	// Set txReceived (transactions to be processed received by clients since last node run)
	ts.FeedbackStrategy.IncrementVarCount("txReceived")
	monitoring.IncreaseTxReceived()

	// TODO: this is an example to prove that, by limiting number of tx per second
	//  when the node is too busy due to high number of goroutines,
	//  the network can regulate itself without leading to blockchain splits or hard forks
	tpsReceived = ts.FeedbackStrategy.IncrementVarCount("tpsReceivedTmp").(int)
	if limitReached, limitLevel := ts.FeedbackStrategy.IsCPULimitReached(constant.FeedbackCPUMinSamples); limitReached {
		if limitLevel == constant.FeedbackLimitHigh {
			ts.Logger.Error("Tx dropped due to high cpu usage")
			monitoring.IncreaseTxFiltered()
			return nil, status.Error(codes.Unavailable, "Service is currently not available")
		}
	}
	// STEF removing goroutine limit (only considering CPU usage)
	// if limitReached, limitLevel := ts.FeedbackStrategy.IsGoroutineLimitReached(constant.FeedbackMinSamples); limitReached {
	// 	switch limitLevel {
	// 	case constant.FeedbackLimitHigh:
	// 		ts.Logger.Error("Tx dropped due to network being spammed with too many transactions")
	// 		monitoring.IncreaseTxFiltered()
	// 		return nil, status.Error(codes.Internal, "TooManyTps")
	// 	case constant.FeedbackLimitMedium:
	// 		if tpsReceived > 1 {
	// 			ts.Logger.Error("Tx dropped due to network being spammed with too many transactions")
	// 			monitoring.IncreaseTxFiltered()
	// 			return nil, status.Error(codes.Internal, "TooManyTps")
	// 		}
	// 	}
	// }
	if limitReached, limitLevel := ts.FeedbackStrategy.IsP2PRequestLimitReached(constant.FeedbackMinSamples); limitReached {
		switch limitLevel {
		case constant.FeedbackLimitHigh:
			ts.Logger.Error("Tx dropped due to node being too busy resolving P2P requests")
			monitoring.IncreaseTxFiltered()
			return nil, status.Error(codes.Internal, "TooManyP2PRequests")
		case constant.FeedbackLimitMedium:
			if tpsReceived > 2 {
				ts.Logger.Error("Tx dropped due to node being too busy resolving P2P requests")
				monitoring.IncreaseTxFiltered()
				return nil, status.Error(codes.Internal, "TooManyP2PRequests")
			}
		}
	}

	// get unsigned bytes
	tx, err = ts.TransactionUtil.ParseTransactionBytes(txBytes, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Validate Tx
	txType, err = ts.ActionTypeSwitcher.GetTransactionType(tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = ts.MempoolService.ValidateMempoolTransaction(tx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Apply Unconfirmed
	err = ts.Query.BeginTx(isDbTransactionHighPriority, monitoring.PostTransactionServiceOwnerProcess)
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
		errRollback := ts.Query.RollbackTx(isDbTransactionHighPriority)
		if errRollback != nil {
			return nil, status.Error(codes.Internal, errRollback.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Save to mempool
	err = ts.MempoolService.AddMempoolTransaction(tx, txBytes)
	if err != nil {
		errRollback := ts.Query.RollbackTx(isDbTransactionHighPriority)
		if errRollback != nil {
			return nil, status.Error(codes.Internal, errRollback.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = ts.Query.CommitTx(isDbTransactionHighPriority)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Set tpsProcessed (transactions per seconds already processed received by clients).
	// Note: these are the ones that produce network traffic because they must be broadcast to peers
	tpsProcessed = ts.FeedbackStrategy.IncrementVarCount("tpsProcessedTmp").(int)
	monitoring.SetTpsProcessed(tpsProcessed)

	// Set txProcessed (transactions already processed received by clients since last node run).
	ts.FeedbackStrategy.IncrementVarCount("txProcessed")
	monitoring.IncreaseTxProcessed()

	ts.Observer.Notify(observer.TransactionAdded, txBytes, chaintype)
	// return parsed transaction
	return tx, nil
}

func (ts *TransactionService) GetTransactionMinimumFee(req *model.GetTransactionMinimumFeeRequest) (
	*model.GetTransactionMinimumFeeResponse, error,
) {
	var (
		txBytes = req.TransactionBytes
		err     error
	)
	tx, err := ts.TransactionUtil.ParseTransactionBytes(txBytes, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// get the TypeAction object
	txType, err := ts.ActionTypeSwitcher.GetTransactionType(tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	minFee, err := txType.GetMinimumFee()
	if err != nil {
		return nil, err
	}
	return &model.GetTransactionMinimumFeeResponse{
		Fee: minFee,
	}, nil
}

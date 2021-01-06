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

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	PendingTransactionServiceInterface interface {
		ExpiringPendingTransactions(blockHeight uint32, useTX bool) error
	}

	PendingTransactionService struct {
		Log                     *logrus.Logger
		QueryExecutor           query.ExecutorInterface
		TypeActionSwitcher      transaction.TypeActionSwitcher
		TransactionUtil         transaction.UtilInterface
		TransactionQuery        query.TransactionQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
	}
)

func NewPendingTransactionService(
	log *logrus.Logger,
	queryExecutor query.ExecutorInterface,
	typeActionSwitcher transaction.TypeActionSwitcher,
	transactionUtil transaction.UtilInterface,
	transactionQuery query.TransactionQueryInterface,
	pendingTransactionQuery query.PendingTransactionQueryInterface,
) PendingTransactionServiceInterface {
	return &PendingTransactionService{
		Log:                     log,
		QueryExecutor:           queryExecutor,
		TypeActionSwitcher:      typeActionSwitcher,
		TransactionUtil:         transactionUtil,
		TransactionQuery:        transactionQuery,
		PendingTransactionQuery: pendingTransactionQuery,
	}
}

// ExpiringPendingTransactions will set status to be expired caused by current block height
func (tg *PendingTransactionService) ExpiringPendingTransactions(blockHeight uint32, useTX bool) error {
	var (
		pendingTransactions []*model.PendingTransaction
		innerTransaction    *model.Transaction
		typeAction          transaction.TypeAction
		rows                *sql.Rows
		err                 error
	)

	err = func() error {
		qy, qArgs := tg.PendingTransactionQuery.GetPendingTransactionsExpireByHeight(blockHeight)
		rows, err = tg.QueryExecutor.ExecuteSelect(qy, useTX, qArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		pendingTransactions, err = tg.PendingTransactionQuery.BuildModel(pendingTransactions, rows)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	if len(pendingTransactions) > 0 {
		if !useTX {
			err = tg.QueryExecutor.BeginTx()
			if err != nil {
				return err
			}
		}
		for _, pendingTransaction := range pendingTransactions {

			/**
			SET PendingTransaction
			1. block height = current block height
			2. status = expired
			*/
			nPendingTransaction := pendingTransaction
			nPendingTransaction.BlockHeight = blockHeight
			nPendingTransaction.Status = model.PendingTransactionStatus_PendingTransactionExpired
			q := tg.PendingTransactionQuery.InsertPendingTransaction(nPendingTransaction)
			err = tg.QueryExecutor.ExecuteTransactions(q)
			if err != nil {
				break
			}
			// Do UndoApplyConfirmed
			innerTransaction, err = tg.TransactionUtil.ParseTransactionBytes(nPendingTransaction.GetTransactionBytes(), false)
			if err != nil {
				break
			}
			typeAction, err = tg.TypeActionSwitcher.GetTransactionType(innerTransaction)
			if err != nil {
				break
			}
			err = typeAction.UndoApplyUnconfirmed()
			if err != nil {
				break
			}
		}

		if !useTX {
			/*
				Check the latest error is not nil, otherwise need to aborting the whole query transactions safety with rollBack.
				And automatically unlock mutex
			*/
			if err != nil {
				if rollbackErr := tg.QueryExecutor.RollbackTx(); rollbackErr != nil {
					tg.Log.Errorf("Rollback fail: %s", rollbackErr.Error())
				}
				return err
			}
			err = tg.QueryExecutor.CommitTx()
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

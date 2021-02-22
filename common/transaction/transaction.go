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
package transaction

import (
	"fmt"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// TypeAction is transaction methods collection
	TypeAction interface {
		// ApplyConfirmed perhaps this method called with QueryExecutor.BeginTX() because inside this process has separated QueryExecutor.Execute
		ApplyConfirmed(blockTimestamp int64) error
		ApplyUnconfirmed() error
		UndoApplyUnconfirmed() error
		// Validate dbTx specify whether validation should read from transaction state or db state
		Validate(dbTx bool) error
		GetMinimumFee() (int64, error)
		GetAmount() int64
		GetSize() (uint32, error)
		ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error)
		GetBodyBytes() ([]byte, error)
		GetTransactionBody(transaction *model.Transaction)
		SkipMempoolTransaction(
			selectedTransactions []*model.Transaction,
			blockTimestamp int64,
			blockHeight uint32,
		) (bool, error)
		// Escrowable check if transaction type has escrow part and it will refill escrow part
		Escrowable() (EscrowTypeAction, bool)
	}
	// TypeActionSwitcher assert transaction to TypeAction / EscrowTypeAction
	TypeActionSwitcher interface {
		GetTransactionType(tx *model.Transaction) (TypeAction, error)
	}
	// TypeSwitcher is TypeActionSwitcher shell
	TypeSwitcher struct {
		Executor                   query.ExecutorInterface
		NodeAuthValidation         auth.NodeAuthValidationInterface
		MempoolCacheStorage        storage.CacheStorageInterface
		NodeAddressInfoStorage     storage.TransactionalCache
		PendingNodeRegistryStorage storage.TransactionalCache
		ActiveNodeRegistryStorage  storage.TransactionalCache
		FeeScaleService            fee.FeeScaleServiceInterface
	}
)

// GetTransactionType assert transaction to TypeAction
func (ts *TypeSwitcher) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	var (
		buf                  = util.ConvertUint32ToBytes(tx.GetTransactionType())
		accountBalanceHelper = NewAccountBalanceHelper(ts.Executor, query.NewAccountBalanceQuery(), query.NewAccountLedgerQuery())
		transactionHelper    = NewTransactionHelper(query.NewTransactionQuery(&chaintype.MainChain{}), ts.Executor)
		transactionBody      model.TransactionBodyInterface
		transactionUtil      = &Util{
			MempoolCacheStorage: ts.MempoolCacheStorage,
			FeeScaleService:     ts.FeeScaleService,
		}
		err error
	)

	switch buf[0] {
	case 0: // Empty Transaction
		switch buf[1] {
		case 0:
			return &TXEmpty{}, nil
		default:
			return nil, nil
		}
	// Send zbc
	case 1:
		switch buf[1] {
		case 0:
			transactionBody, err = new(SendZBC).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SendZBC{
				TransactionObject:    tx,
				Body:                 transactionBody.(*model.SendZBCTransactionBody),
				QueryExecutor:        ts.Executor,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				BlockQuery:           query.NewBlockQuery(&chaintype.MainChain{}),
				FeeScaleService:      ts.FeeScaleService,
				AccountBalanceHelper: accountBalanceHelper,
			}, nil
		default:
			return nil, nil
		}
	// Node Registry
	case 2:
		switch buf[1] {
		case 0:
			transactionBody, err = (&NodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &NodeRegistration{
				TransactionObject:        tx,
				Body:                     transactionBody.(*model.NodeRegistrationTransactionBody),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery:  query.NewParticipationScoreQuery(),
				AuthPoown:                ts.NodeAuthValidation,
				QueryExecutor:            ts.Executor,
				EscrowQuery:              query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:     accountBalanceHelper,
				FeeScaleService:          ts.FeeScaleService,
				PendingNodeRegistryCache: ts.PendingNodeRegistryStorage,
			}, nil
		case 1:
			transactionBody, err = (&UpdateNodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &UpdateNodeRegistration{
				TransactionObject:            tx,
				Body:                         transactionBody.(*model.UpdateNodeRegistrationTransactionBody),
				NodeRegistrationQuery:        query.NewNodeRegistrationQuery(),
				BlockQuery:                   query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:                    ts.NodeAuthValidation,
				QueryExecutor:                ts.Executor,
				EscrowQuery:                  query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:         accountBalanceHelper,
				FeeScaleService:              ts.FeeScaleService,
				PendingNodeRegistrationCache: ts.PendingNodeRegistryStorage,
				ActiveNodeRegistrationCache:  ts.ActiveNodeRegistryStorage,
			}, nil
		case 2:
			transactionBody, err = new(RemoveNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveNodeRegistration{
				TransactionObject:        tx,
				Body:                     transactionBody.(*model.RemoveNodeRegistrationTransactionBody),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				NodeAddressInfoQuery:     query.NewNodeAddressInfoQuery(),
				QueryExecutor:            ts.Executor,
				AccountBalanceHelper:     accountBalanceHelper,
				EscrowQuery:              query.NewEscrowTransactionQuery(),
				FeeScaleService:          ts.FeeScaleService,
				NodeAddressInfoStorage:   ts.NodeAddressInfoStorage,
				PendingNodeRegistryCache: ts.PendingNodeRegistryStorage,
				ActiveNodeRegistryCache:  ts.ActiveNodeRegistryStorage,
			}, nil
		case 3:
			transactionBody, err = new(ClaimNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &ClaimNodeRegistration{
				TransactionObject:       tx,
				Body:                    transactionBody.(*model.ClaimNodeRegistrationTransactionBody),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:               ts.NodeAuthValidation,
				QueryExecutor:           ts.Executor,
				AccountBalanceHelper:    accountBalanceHelper,
				EscrowQuery:             query.NewEscrowTransactionQuery(),
				FeeScaleService:         ts.FeeScaleService,
				NodeAddressInfoStorage:  ts.NodeAddressInfoStorage,
				ActiveNodeRegistryCache: ts.ActiveNodeRegistryStorage,
				NodeAddressInfoQuery:    query.NewNodeAddressInfoQuery(),
			}, nil
		default:
			return nil, nil
		}
	// Account Dataset
	case 3:
		switch buf[1] {
		case 0:
			transactionBody, err = new(SetupAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SetupAccountDataset{
				TransactionObject:    tx,
				Body:                 transactionBody.(*model.SetupAccountDatasetTransactionBody),
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        ts.Executor,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				FeeScaleService:      ts.FeeScaleService,
			}, nil
		case 1:
			transactionBody, err = new(RemoveAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveAccountDataset{
				TransactionObject:    tx,
				Body:                 transactionBody.(*model.RemoveAccountDatasetTransactionBody),
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        ts.Executor,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				FeeScaleService:      ts.FeeScaleService,
			}, nil
		default:
			return nil, nil
		}
	// Escrow
	case 4:
		switch buf[1] {
		case 0:
			transactionBody, err = new(ApprovalEscrowTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &ApprovalEscrowTransaction{
				TransactionObject:    tx,
				Body:                 transactionBody.(*model.ApprovalEscrowTransactionBody),
				QueryExecutor:        ts.Executor,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				TypeActionSwitcher:   ts,
				TransactionQuery:     query.NewTransactionQuery(&chaintype.MainChain{}),
				BlockQuery:           query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper: accountBalanceHelper,
				FeeScaleService:      ts.FeeScaleService,
			}, nil
		default:
			return nil, nil
		}
	// Multi Signature
	case 5:
		switch buf[1] {
		// MultiSignatureTransaction
		case 0:
			// initialize service for pending_tx, pending_sig and multisig_info
			typeSwitcher := &TypeSwitcher{
				Executor: ts.Executor,
			}

			pendingTransactionHelper := &PendingTransactionHelper{
				MultisignatureInfoQuery: query.NewMultisignatureInfoQuery(),
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				TransactionUtil:         transactionUtil,
				TypeSwitcher:            typeSwitcher,
				QueryExecutor:           ts.Executor,
			}
			multisignatureInfoHelper := &MultisignatureInfoHelper{
				MultiSignatureParticipantQuery: query.NewMultiSignatureParticipantQuery(),
				MultisignatureInfoQuery:        query.NewMultisignatureInfoQuery(),
				QueryExecutor:                  ts.Executor,
			}
			signatureInfoHelper := &SignatureInfoHelper{
				PendingSignatureQuery:   query.NewPendingSignatureQuery(),
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           ts.Executor,
				Signature:               &crypto.Signature{},
			}
			multisigUtil := NewMultisigTransactionUtil()
			transactionBody, err = new(MultiSignatureTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &MultiSignatureTransaction{
				TransactionObject:        tx,
				Body:                     transactionBody.(*model.MultiSignatureTransactionBody),
				TransactionUtil:          transactionUtil,
				TypeSwitcher:             typeSwitcher,
				Signature:                &crypto.Signature{},
				TransactionHelper:        transactionHelper,
				AccountBalanceHelper:     accountBalanceHelper,
				MultisigUtil:             multisigUtil,
				SignatureInfoHelper:      signatureInfoHelper,
				PendingTransactionHelper: pendingTransactionHelper,
				MultisignatureInfoHelper: multisignatureInfoHelper,
				FeeScaleService:          ts.FeeScaleService,
				EscrowQuery:              query.NewEscrowTransactionQuery(),
				QueryExecutor:            ts.Executor,
			}, nil
		default:
			return nil, nil
		}
	case 6:
		switch buf[1] {
		case 0: // LiquidPayment Transaction
			transactionBody, err = new(LiquidPaymentTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &LiquidPaymentTransaction{
				TransactionObject:             tx,
				Body:                          transactionBody.(*model.LiquidPaymentTransactionBody),
				QueryExecutor:                 ts.Executor,
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				EscrowQuery:                   query.NewEscrowTransactionQuery(),
				FeeScaleService:               ts.FeeScaleService,
			}, nil
		case 1: // LiquidPaymentStop Transaction
			transactionBody, err = new(LiquidPaymentStopTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &LiquidPaymentStopTransaction{
				TransactionObject:             tx,
				Body:                          transactionBody.(*model.LiquidPaymentStopTransactionBody),
				FeeScaleService:               ts.FeeScaleService,
				QueryExecutor:                 ts.Executor,
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher:            ts,
				EscrowQuery:                   query.NewEscrowTransactionQuery(),
			}, nil
		default:
			return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf("transaction type is not valid: %v", buf[1]))
		}
	// Fee Voting
	case 7:
		switch buf[1] {
		case 0:
			transactionBody, err = new(FeeVoteCommitTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &FeeVoteCommitTransaction{
				TransactionObject:          tx,
				Body:                       transactionBody.(*model.FeeVoteCommitTransactionBody),
				QueryExecutor:              ts.Executor,
				AccountBalanceHelper:       accountBalanceHelper,
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				NodeRegistrationQuery:      query.NewNodeRegistrationQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeScaleService:            ts.FeeScaleService,
				EscrowQuery:                query.NewEscrowTransactionQuery(),
			}, nil
		case 1:
			transactionBody, err = new(FeeVoteRevealTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &FeeVoteRevealTransaction{
				TransactionObject:      tx,
				Body:                   transactionBody.(*model.FeeVoteRevealTransactionBody),
				QueryExecutor:          ts.Executor,
				AccountBalanceHelper:   accountBalanceHelper,
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				FeeVoteCommitVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeVoteRevealVoteQuery: query.NewFeeVoteRevealVoteQuery(),
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				SignatureInterface:     crypto.NewSignature(),
				FeeScaleService:        ts.FeeScaleService,
				EscrowQuery:            query.NewEscrowTransactionQuery(),
			}, nil
		default:
			return nil, nil
		}
	default:
		return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf("transaction type is not valid: %v", buf[0]))
	}
}

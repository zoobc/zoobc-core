package transaction

import (
	"fmt"

	"github.com/zoobc/zoobc-core/common/storage"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// TypeAction is transaction methods collection
	TypeAction interface {
		ApplyConfirmed(blockTimestamp int64) error
		ApplyUnconfirmed() error
		UndoApplyUnconfirmed() error
		// Validate dbTx specify whether validation should read from transaction state or db state
		Validate(dbTx bool) error
		GetMinimumFee() (int64, error)
		GetAmount() int64
		GetSize() uint32
		ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error)
		GetBodyBytes() []byte
		GetTransactionBody(transaction *model.Transaction)
		SkipMempoolTransaction(
			selectedTransactions []*model.Transaction,
			blockTimestamp int64,
			blockHeight uint32,
		) (bool, error)
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
		NodeAddressInfoStorage     storage.NodeAddressInfoStorageInterface
		MempoolCacheStorage        storage.CacheStorageInterface
		PendingNodeRegistryStorage storage.CacheStorageInterface
		ActiveNodeRegistryStorage  storage.CacheStorageInterface
	}
)

// GetTransactionType assert transaction to TypeAction
func (ts *TypeSwitcher) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	var (
		buf                  = util.ConvertUint32ToBytes(tx.GetTransactionType())
		accountBalanceHelper = NewAccountBalanceHelper(query.NewAccountBalanceQuery(), ts.Executor)
		accountLedgerHelper  = NewAccountLedgerHelper(query.NewAccountLedgerQuery(), ts.Executor)
		transactionHelper    = NewTransactionHelper(query.NewTransactionQuery(&chaintype.MainChain{}), ts.Executor)
		transactionBody      model.TransactionBodyInterface
		feeScaleService      = fee.NewFeeScaleService(
			query.NewFeeScaleQuery(),
			query.NewBlockQuery(&chaintype.MainChain{}),
			ts.Executor,
		)
		transactionUtil = &Util{
			MempoolCacheStorage: ts.MempoolCacheStorage,
			FeeScaleService:     feeScaleService,
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
	// Send Money
	case 1:
		switch buf[1] {
		case 0:
			transactionBody, err = new(SendMoney).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SendMoney{
				ID:                  tx.GetID(),
				Body:                transactionBody.(*model.SendMoneyTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				RecipientAddress:    tx.GetRecipientAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				Escrow:              tx.GetEscrow(),
				EscrowQuery:         query.NewEscrowTransactionQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:           fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
			}, nil
		default:
			return nil, nil
		}
	// Node Registry
	case 2:
		txPendingNodeRegistryCache := ts.PendingNodeRegistryStorage.(storage.TransactionalCache)
		txActiveNodeRegistryCache := ts.PendingNodeRegistryStorage.(storage.TransactionalCache)
		switch buf[1] {
		case 0:
			transactionBody, err = (&NodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &NodeRegistration{
				ID:                       tx.ID, // assign tx ID to nodeID
				Body:                     transactionBody.(*model.NodeRegistrationTransactionBody),
				Fee:                      tx.Fee,
				SenderAddress:            tx.GetSenderAccountAddress(),
				Height:                   tx.GetHeight(),
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery:  query.NewParticipationScoreQuery(),
				AuthPoown:                ts.NodeAuthValidation,
				QueryExecutor:            ts.Executor,
				AccountLedgerQuery:       query.NewAccountLedgerQuery(),
				Escrow:                   tx.GetEscrow(),
				PendingNodeRegistryCache: txPendingNodeRegistryCache,
			}, nil
		case 1:
			transactionBody, err = (&UpdateNodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &UpdateNodeRegistration{
				ID:                           tx.GetID(),
				Body:                         transactionBody.(*model.UpdateNodeRegistrationTransactionBody),
				Fee:                          tx.Fee,
				SenderAddress:                tx.GetSenderAccountAddress(),
				Height:                       tx.GetHeight(),
				AccountBalanceQuery:          query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:        query.NewNodeRegistrationQuery(),
				BlockQuery:                   query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:                    ts.NodeAuthValidation,
				QueryExecutor:                ts.Executor,
				AccountLedgerQuery:           query.NewAccountLedgerQuery(),
				Escrow:                       tx.GetEscrow(),
				PendingNodeRegistrationCache: txPendingNodeRegistryCache,
				ActiveNodeRegistrationCache:  txActiveNodeRegistryCache,
			}, nil
		case 2:
			transactionBody, err = new(RemoveNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveNodeRegistration{
				ID:                       tx.GetID(),
				Body:                     transactionBody.(*model.RemoveNodeRegistrationTransactionBody),
				Fee:                      tx.Fee,
				SenderAddress:            tx.GetSenderAccountAddress(),
				Height:                   tx.GetHeight(),
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				NodeAddressInfoQuery:     query.NewNodeAddressInfoQuery(),
				QueryExecutor:            ts.Executor,
				AccountLedgerQuery:       query.NewAccountLedgerQuery(),
				AccountBalanceHelper:     accountBalanceHelper,
				NodeAddressInfoStorage:   ts.NodeAddressInfoStorage,
				PendingNodeRegistryCache: txPendingNodeRegistryCache,
				ActiveNodeRegistryCache:  txActiveNodeRegistryCache,
			}, nil
		case 3:
			transactionBody, err = new(ClaimNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &ClaimNodeRegistration{
				ID:                      tx.GetID(),
				Body:                    transactionBody.(*model.ClaimNodeRegistrationTransactionBody),
				Fee:                     tx.Fee,
				SenderAddress:           tx.GetSenderAccountAddress(),
				Height:                  tx.GetHeight(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:               ts.NodeAuthValidation,
				QueryExecutor:           ts.Executor,
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				AccountBalanceHelper:    accountBalanceHelper,
				NodeAddressInfoStorage:  ts.NodeAddressInfoStorage,
				ActiveNodeRegistryCache: txActiveNodeRegistryCache,
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
				ID:                  tx.GetID(),
				Body:                transactionBody.(*model.SetupAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				RecipientAddress:    tx.GetRecipientAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				Escrow:              tx.GetEscrow(),
			}, nil
		case 1:
			transactionBody, err = new(RemoveAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveAccountDataset{
				ID:                  tx.GetID(),
				Body:                transactionBody.(*model.RemoveAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				RecipientAddress:    tx.GetRecipientAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				Escrow:              tx.GetEscrow(),
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
				ID:                  tx.GetID(),
				Body:                transactionBody.(*model.ApprovalEscrowTransactionBody),
				Fee:                 tx.GetFee(),
				SenderAddress:       tx.GetSenderAccountAddress(),
				Height:              tx.GetHeight(),
				Escrow:              tx.GetEscrow(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				EscrowQuery:         query.NewEscrowTransactionQuery(),
				TypeActionSwitcher:  ts,
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
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
				ID:                       tx.ID,
				Body:                     transactionBody.(*model.MultiSignatureTransactionBody),
				Fee:                      tx.GetFee(),
				SenderAddress:            tx.GetSenderAccountAddress(),
				NormalFee:                fee.NewConstantFeeModel(constant.OneZBC / 100),
				TransactionUtil:          transactionUtil,
				TypeSwitcher:             typeSwitcher,
				Signature:                &crypto.Signature{},
				Height:                   tx.Height,
				BlockID:                  tx.BlockID,
				TransactionHelper:        transactionHelper,
				AccountBalanceHelper:     accountBalanceHelper,
				AccountLedgerHelper:      accountLedgerHelper,
				MultisigUtil:             multisigUtil,
				SignatureInfoHelper:      signatureInfoHelper,
				PendingTransactionHelper: pendingTransactionHelper,
				MultisignatureInfoHelper: multisignatureInfoHelper,
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
				ID:                            tx.GetID(),
				Body:                          transactionBody.(*model.LiquidPaymentTransactionBody),
				Fee:                           tx.GetFee(),
				SenderAddress:                 tx.GetSenderAccountAddress(),
				RecipientAddress:              tx.GetRecipientAccountAddress(),
				Height:                        tx.GetHeight(),
				NormalFee:                     fee.NewConstantFeeModel(constant.OneZBC / 100),
				QueryExecutor:                 ts.Executor,
				AccountBalanceHelper:          accountBalanceHelper,
				AccountLedgerHelper:           accountLedgerHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
			}, nil
		case 1: // LiquidPaymentStop Transaction
			transactionBody, err = new(LiquidPaymentStopTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &LiquidPaymentStopTransaction{
				ID:                            tx.GetID(),
				Body:                          transactionBody.(*model.LiquidPaymentStopTransactionBody),
				Fee:                           tx.GetFee(),
				SenderAddress:                 tx.GetSenderAccountAddress(),
				RecipientAddress:              tx.GetRecipientAccountAddress(),
				Height:                        tx.GetHeight(),
				NormalFee:                     fee.NewConstantFeeModel(constant.OneZBC / 100),
				QueryExecutor:                 ts.Executor,
				AccountBalanceHelper:          accountBalanceHelper,
				AccountLedgerHelper:           accountLedgerHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher:            ts,
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
				ID:                         tx.ID,
				Fee:                        tx.Fee,
				SenderAddress:              tx.SenderAccountAddress,
				Height:                     tx.Height,
				Body:                       transactionBody.(*model.FeeVoteCommitTransactionBody),
				QueryExecutor:              ts.Executor,
				AccountBalanceHelper:       accountBalanceHelper,
				AccountLedgerHelper:        accountLedgerHelper,
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				NodeRegistrationQuery:      query.NewNodeRegistrationQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeScaleService:            feeScaleService,
			}, nil
		case 1:
			transactionBody, err = new(FeeVoteRevealTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &FeeVoteRevealTransaction{
				ID:                     tx.GetID(),
				Fee:                    tx.GetFee(),
				SenderAddress:          tx.GetSenderAccountAddress(),
				Height:                 tx.GetHeight(),
				Timestamp:              tx.GetTimestamp(),
				Body:                   transactionBody.(*model.FeeVoteRevealTransactionBody),
				QueryExecutor:          ts.Executor,
				AccountBalanceHelper:   accountBalanceHelper,
				AccountLedgerHelper:    accountLedgerHelper,
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				FeeVoteCommitVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeVoteRevealVoteQuery: query.NewFeeVoteRevealVoteQuery(),
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				SignatureInterface:     crypto.NewSignature(),
				FeeScaleService:        feeScaleService,
			}, nil
		default:
			return nil, nil
		}
	default:
		return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf("transaction type is not valid: %v", buf[0]))
	}
}

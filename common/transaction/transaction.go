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
	// Send Money
	case 1:
		switch buf[1] {
		case 0:
			transactionBody, err = new(SendMoney).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SendMoney{
				ID:               tx.GetID(),
				Body:             transactionBody.(*model.SendMoneyTransactionBody),
				Fee:              tx.Fee,
				SenderAddress:    tx.GetSenderAccountAddress(),
				RecipientAddress: tx.GetRecipientAccountAddress(),
				Height:           tx.GetHeight(),
				QueryExecutor:    ts.Executor,
				Escrow:           tx.GetEscrow(),
				EscrowQuery:      query.NewEscrowTransactionQuery(),
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:            fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				ID:                      tx.ID, // assign tx ID to nodeID
				Body:                    transactionBody.(*model.NodeRegistrationTransactionBody),
				Fee:                     tx.Fee,
				SenderAddress:           tx.GetSenderAccountAddress(),
				Height:                  tx.GetHeight(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				AuthPoown:               &auth.NodeAuthValidation{},
				QueryExecutor:           ts.Executor,
				Escrow:                  tx.GetEscrow(),
				EscrowQuery:             query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:    accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:                fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				ID:                    tx.GetID(),
				Body:                  transactionBody.(*model.UpdateNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.NodeAuthValidation{},
				QueryExecutor:         ts.Executor,
				Escrow:                tx.GetEscrow(),
				EscrowQuery:           query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:  accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:                    fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				PendingNodeRegistrationCache: ts.PendingNodeRegistryStorage,
				ActiveNodeRegistrationCache:  ts.ActiveNodeRegistryStorage,
			}, nil
		case 2:
			transactionBody, err = new(RemoveNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveNodeRegistration{
				ID:                    tx.GetID(),
				Body:                  transactionBody.(*model.RemoveNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				NodeAddressInfoQuery:  query.NewNodeAddressInfoQuery(),
				QueryExecutor:         ts.Executor,
				AccountBalanceHelper:  accountBalanceHelper,
				Escrow:                tx.GetEscrow(),
				EscrowQuery:           query.NewEscrowTransactionQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:                fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				ID:                    tx.GetID(),
				Body:                  transactionBody.(*model.ClaimNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.NodeAuthValidation{},
				QueryExecutor:         ts.Executor,
				AccountBalanceHelper:  accountBalanceHelper,
				Escrow:                tx.GetEscrow(),
				EscrowQuery:           query.NewEscrowTransactionQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:               fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				NodeAddressInfoStorage:  ts.NodeAddressInfoStorage,
				ActiveNodeRegistryCache: ts.ActiveNodeRegistryStorage,
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
				ID:                   tx.GetID(),
				Body:                 transactionBody.(*model.SetupAccountDatasetTransactionBody),
				Fee:                  tx.Fee,
				SenderAddress:        tx.GetSenderAccountAddress(),
				RecipientAddress:     tx.GetRecipientAccountAddress(),
				Height:               tx.GetHeight(),
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        ts.Executor,
				Escrow:               tx.GetEscrow(),
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
			}, nil
		case 1:
			transactionBody, err = new(RemoveAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveAccountDataset{
				ID:                   tx.GetID(),
				Body:                 transactionBody.(*model.RemoveAccountDatasetTransactionBody),
				Fee:                  tx.Fee,
				SenderAddress:        tx.GetSenderAccountAddress(),
				RecipientAddress:     tx.GetRecipientAccountAddress(),
				Height:               tx.GetHeight(),
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        ts.Executor,
				Escrow:               tx.GetEscrow(),
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				ID:                   tx.GetID(),
				Body:                 transactionBody.(*model.ApprovalEscrowTransactionBody),
				Fee:                  tx.GetFee(),
				SenderAddress:        tx.GetSenderAccountAddress(),
				Height:               tx.GetHeight(),
				Escrow:               tx.GetEscrow(),
				QueryExecutor:        ts.Executor,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				TypeActionSwitcher:   ts,
				TransactionQuery:     query.NewTransactionQuery(&chaintype.MainChain{}),
				BlockQuery:           query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				TransactionUtil:          transactionUtil,
				TypeSwitcher:             typeSwitcher,
				Signature:                &crypto.Signature{},
				Height:                   tx.Height,
				BlockID:                  tx.BlockID,
				TransactionHelper:        transactionHelper,
				AccountBalanceHelper:     accountBalanceHelper,
				MultisigUtil:             multisigUtil,
				SignatureInfoHelper:      signatureInfoHelper,
				PendingTransactionHelper: pendingTransactionHelper,
				MultisignatureInfoHelper: multisignatureInfoHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:     fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				EscrowQuery:   query.NewEscrowTransactionQuery(),
				QueryExecutor: ts.Executor,
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
				QueryExecutor:                 ts.Executor,
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				Escrow:                        tx.GetEscrow(),
				EscrowQuery:                   query.NewEscrowTransactionQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
			}, nil
		case 1: // LiquidPaymentStop Transaction
			transactionBody, err = new(LiquidPaymentStopTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &LiquidPaymentStopTransaction{
				ID:               tx.GetID(),
				Body:             transactionBody.(*model.LiquidPaymentStopTransactionBody),
				Fee:              tx.GetFee(),
				SenderAddress:    tx.GetSenderAccountAddress(),
				RecipientAddress: tx.GetRecipientAccountAddress(),
				Height:           tx.GetHeight(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),

				QueryExecutor:                 ts.Executor,
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher:            ts,
				Escrow:                        tx.GetEscrow(),
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
				ID:                         tx.ID,
				Fee:                        tx.Fee,
				SenderAddress:              tx.SenderAccountAddress,
				Height:                     tx.Height,
				Body:                       transactionBody.(*model.FeeVoteCommitTransactionBody),
				QueryExecutor:              ts.Executor,
				AccountBalanceHelper:       accountBalanceHelper,
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				NodeRegistrationQuery:      query.NewNodeRegistrationQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeScaleService:            ts.FeeScaleService,
				Escrow:                     tx.GetEscrow(),
				EscrowQuery:                query.NewEscrowTransactionQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				FeeVoteCommitVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeVoteRevealVoteQuery: query.NewFeeVoteRevealVoteQuery(),
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				SignatureInterface:     crypto.NewSignature(),
				FeeScaleService:        ts.FeeScaleService,
				Escrow:                 tx.GetEscrow(),
				EscrowQuery:            query.NewEscrowTransactionQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
			}, nil
		default:
			return nil, nil
		}
	case 8:
		transactionBody, err = (&AtomicTransaction{
			TransactionUtil: transactionUtil,
		}).ParseBodyBytes(tx.GetTransactionBodyBytes())
		if err != nil {
			return nil, err
		}
		return &AtomicTransaction{
			ID:                     tx.GetID(),
			Fee:                    tx.GetFee(),
			SenderAddress:          tx.GetSenderAccountAddress(),
			Height:                 tx.GetHeight(),
			Body:                   transactionBody.(*model.AtomicTransactionBody),
			AtomicTransactionQuery: query.NewAtomicTransactionQuery(),
			Escrow:                 tx.GetEscrow(),
			EscrowQuery:            query.NewEscrowTransactionQuery(),
			QueryExecutor:          ts.Executor,
			TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
			TypeActionSwitcher:     ts,
			AccountBalanceHelper:   accountBalanceHelper,
			EscrowFee: fee.NewBlockLifeTimeFeeModel(
				10, fee.SendMoneyFeeConstant,
			),
			NormalFee:       fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
			TransactionUtil: transactionUtil,
			Signature:       crypto.NewSignature(),
		}, nil
	default:
		return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf("transaction type is not valid: %v", buf[0]))
	}
}

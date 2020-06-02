package transaction

import (
	"fmt"

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
		SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error)
		Escrowable() (EscrowTypeAction, bool)
	}
	// TypeActionSwitcher assert transaction to TypeAction / EscrowTypeAction
	TypeActionSwitcher interface {
		GetTransactionType(tx *model.Transaction) (TypeAction, error)
	}
	// TypeSwitcher is TypeActionSwitcher shell
	TypeSwitcher struct {
		Executor query.ExecutorInterface
	}
)

// GetTransactionType assert transaction to TypeAction
func (ts *TypeSwitcher) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	buf := util.ConvertUint32ToBytes(tx.GetTransactionType())
	// TODO OPTIMIZATION: initiate accountBalanceHelper and accountLedgerHelper once only in the instance lifecycle
	accountBalanceHelper := NewAccountBalanceHelper(query.NewAccountBalanceQuery(), ts.Executor)
	accountLedgerHelper := NewAccountLedgerHelper(query.NewAccountLedgerQuery(), ts.Executor)
	transactionHelper := NewTransactionHelper(query.NewTransactionQuery(&chaintype.MainChain{}), ts.Executor)
	switch buf[0] {
	case 0: // Empty Transaction
		switch buf[1] {
		case 0:
			return &TXEmpty{}, nil
		default:
			return nil, nil
		}
	case 1: // SendMoney Transaction
		switch buf[1] {
		case 0:
			sendMoneyBody, err := new(SendMoney).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SendMoney{
				ID:                  tx.GetID(),
				Body:                sendMoneyBody.(*model.SendMoneyTransactionBody),
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
					10, constant.OneZBC/100,
				),
				NormalFee:           fee.NewConstantFeeModel(constant.OneZBC / 100),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
			}, nil
		default:
			return nil, nil
		}
	case 2: // NodeRegistration Transaction
		switch buf[1] {
		case 0: // NodeRegistration Transaction
			nodeRegistrationBody, err := (&NodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &NodeRegistration{
				ID:                      tx.ID, // assign tx ID to nodeID
				Body:                    nodeRegistrationBody.(*model.NodeRegistrationTransactionBody),
				Fee:                     tx.Fee,
				SenderAddress:           tx.GetSenderAccountAddress(),
				Height:                  tx.GetHeight(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				AuthPoown:               &auth.ProofOfOwnershipValidation{},
				QueryExecutor:           ts.Executor,
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				Escrow:                  tx.GetEscrow(),
			}, nil
		case 1: // UpdateNodeRegistration Transaction
			nodeRegistrationBody, err := (&UpdateNodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &UpdateNodeRegistration{
				ID:                    tx.GetID(),
				Body:                  nodeRegistrationBody.(*model.UpdateNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.ProofOfOwnershipValidation{},
				QueryExecutor:         ts.Executor,
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
				Escrow:                tx.GetEscrow(),
			}, nil
		case 2: // RemoveNodeRegistration Transaction
			removeNodeRegistrationBody, err := new(RemoveNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveNodeRegistration{
				ID:                    tx.GetID(),
				Body:                  removeNodeRegistrationBody.(*model.RemoveNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         ts.Executor,
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
				Escrow:                tx.GetEscrow(),
			}, nil
		case 3: // ClaimNodeRegistration Transaction
			claimNodeRegistrationBody, err := new(ClaimNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &ClaimNodeRegistration{
				ID:                    tx.GetID(),
				Body:                  claimNodeRegistrationBody.(*model.ClaimNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.ProofOfOwnershipValidation{},
				QueryExecutor:         ts.Executor,
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
				Escrow:                tx.GetEscrow(),
			}, nil
		default:
			return nil, nil
		}
	case 3: // AccountDataset Transaction
		switch buf[1] {
		case 0: // SetupAccountDataset Transaction
			setupAccountDatasetTransactionBody, err := new(SetupAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SetupAccountDataset{
				ID:                  tx.GetID(),
				Body:                setupAccountDatasetTransactionBody.(*model.SetupAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				Escrow:              tx.GetEscrow(),
			}, nil
		case 1: // RemoveAccountDataset Transaction
			removeAccountDatasetTransactionBody, err := new(RemoveAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveAccountDataset{
				ID:                  tx.GetID(),
				Body:                removeAccountDatasetTransactionBody.(*model.RemoveAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
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
	case 4: // Escrow Transaction
		switch buf[1] {
		case 0: // ApprovalEscrowTransaction Transaction
			approvalEscrowTransactionBody, err := new(ApprovalEscrowTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &ApprovalEscrowTransaction{
				ID:                  tx.GetID(),
				Body:                approvalEscrowTransactionBody.(*model.ApprovalEscrowTransactionBody),
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
	case 5: // NewMultisig Transaction
		switch buf[1] {
		case 0: // MultiSignatureTransaction
			// initialize service for pending_tx, pending_sig and multisig_info
			typeSwitcher := &TypeSwitcher{
				Executor: ts.Executor,
			}

			pendingTransactionHelper := &PendingTransactionHelper{
				MultisignatureInfoQuery: query.NewMultisignatureInfoQuery(),
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				TransactionUtil:         &Util{},
				TypeSwitcher:            typeSwitcher,
				QueryExecutor:           ts.Executor,
			}
			multisignatureInfoHelper := &MultisignatureInfoHelper{
				MultisignatureInfoQuery: query.NewMultisignatureInfoQuery(),
				QueryExecutor:           ts.Executor,
			}
			signatureInfoHelper := &SignatureInfoHelper{
				PendingSignatureQuery:   query.NewPendingSignatureQuery(),
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           ts.Executor,
				Signature:               &crypto.Signature{},
			}
			multisigUtil := NewMultisigTransactionUtil()
			multiSigTransactionBody, err := new(MultiSignatureTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &MultiSignatureTransaction{
				ID:                       tx.ID,
				Body:                     multiSigTransactionBody.(*model.MultiSignatureTransactionBody),
				Fee:                      tx.GetFee(),
				SenderAddress:            tx.GetSenderAccountAddress(),
				NormalFee:                fee.NewConstantFeeModel(constant.OneZBC / 100),
				TransactionUtil:          &Util{},
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
	case 6: // LiquidPayment Transaction
		switch buf[1] {
		case 0: // LiquidPayment Transaction
			liquidPaymentTransactionBody, err := new(LiquidPaymentTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &LiquidPaymentTransaction{
				ID:                            tx.GetID(),
				Body:                          liquidPaymentTransactionBody.(*model.LiquidPaymentTransactionBody),
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
			liquidPaymentStopTransactionBody, err := new(LiquidPaymentStopTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &LiquidPaymentStopTransaction{
				ID:                            tx.GetID(),
				Body:                          liquidPaymentStopTransactionBody.(*model.LiquidPaymentStopTransactionBody),
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
	default:
		return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf("transaction type is not valid: %v", buf[0]))
	}
}

package transaction

import (
	"github.com/zoobc/zoobc-core/common/auth"
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
	// TypeActionSwitcher assert transaction to TypeAction / EscrowTyepAction
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
	switch buf[0] {
	case 0:
		switch buf[1] {
		case 0:
			return &TXEmpty{}, nil
		default:
			return nil, nil
		}
	case 1:
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
				NormalFee: fee.NewConstantFeeModel(constant.OneZBC / 100),
			}, nil
		default:
			return nil, nil
		}
	case 2:
		switch buf[1] {
		case 0:
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
		case 1:
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
		case 2:
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
		case 3:
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
	case 3:
		switch buf[1] {
		case 0:
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
		case 1:
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
	case 4:
		switch buf[1] {
		case 0:
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
	case 5:
		switch buf[1] {
		case 0:
			// initialize service for pending_tx, pending_sig and multisig_info
			multisigUtil := NewMultisigTransactionUtil(
				ts.Executor,
				query.NewPendingTransactionQuery(),
				query.NewPendingSignatureQuery(),
				query.NewMultisignatureInfoQuery(),
				&Util{},
			)
			multiSigTransactionBody, err := new(MultiSignatureTransaction).ParseBodyBytes(tx.GetTransactionBodyBytes())
			if err != nil {
				return nil, err
			}
			return &MultiSignatureTransaction{
				ID:              tx.ID,
				Body:            multiSigTransactionBody.(*model.MultiSignatureTransactionBody),
				Fee:             tx.GetFee(),
				SenderAddress:   tx.GetSenderAccountAddress(),
				NormalFee:       fee.NewConstantFeeModel(constant.OneZBC / 100),
				TransactionUtil: &Util{},
				TypeSwitcher: &TypeSwitcher{
					Executor: ts.Executor,
				},
				Signature:               &crypto.Signature{},
				Height:                  tx.Height,
				BlockID:                 tx.BlockID,
				MultisigUtil:            multisigUtil,
				QueryExecutor:           ts.Executor,
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				MultisignatureInfoQuery: query.NewMultisignatureInfoQuery(),
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				PendingSignatureQuery:   query.NewPendingSignatureQuery(),
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
			}, nil
		default:
			return nil, nil
		}
	default:
		return nil, nil
	}
}

package service

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type MockQueryExecutor struct {
	query.ExecutorInterface
	BeginTxError  error
	CommitTxError error
}

func (m *MockQueryExecutor) BeginTx() error {
	return m.BeginTxError
}

func (m *MockQueryExecutor) CommitTx() error {
	return m.CommitTxError
}

type MockActionTypeSwitcher struct {
	transaction.TypeActionSwitcher
	GetTransactionTypeResult transaction.TypeAction
	GetTransactionTypeError  error
}

func (m *MockActionTypeSwitcher) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return m.GetTransactionTypeResult, m.GetTransactionTypeError
}

type MockTypeAction struct {
	transaction.TypeAction
	ApplyUnconfirmedError error
}

func (m *MockTypeAction) ApplyUnconfirmed() error {
	return m.ApplyUnconfirmedError
}

type MockMempoolServiceUtil struct {
	MempoolServiceUtilInterface
	ValidatidateMempoolError   error
	AddMempoolTransactionError error
}

func (m *MockMempoolServiceUtil) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return m.ValidatidateMempoolError
}

func (m *MockMempoolServiceUtil) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return m.AddMempoolTransactionError
}

type MockTransactionUtil struct {
	transaction.UtilInterface
	ParseTransactionResult *model.Transaction
	ParseTransactionError  error
}

func (m *MockTransactionUtil) ParseTransactionBytes(transactionBytes []byte, sign bool) (*model.Transaction, error) {
	return m.ParseTransactionResult, m.ParseTransactionError
}

type MockKVExecutor struct {
	kvdb.KVExecutorInterface
	KvdbGetResult []byte
	KvdbGetError  error
}

func (m *MockKVExecutor) Get(key string) ([]byte, error) {
	return m.KvdbGetResult, m.KvdbGetError
}

type MockReceiptUtil struct {
	coreUtil.ReceiptUtilInterface
	GetReceiptKeyResult []byte
	GetReceiptKeyError  error
}

func (m *MockReceiptUtil) GetReceiptKey(
	dataHash, senderPublicKey []byte,
) ([]byte, error) {
	return m.GetReceiptKeyResult, m.GetReceiptKeyError
}

type MockReceiptService struct {
	ReceiptServiceInterface
	GenerateBatchReceiptWithReminderResult *model.BatchReceipt
	GenerateBatchReceiptWithReminderError  error
}

func (m *MockReceiptService) GenerateBatchReceiptWithReminder(
	ct chaintype.ChainType,
	receivedDatumHash []byte,
	lastBlock *model.Block,
	senderPublicKey []byte,
	nodeSecretPhrase, receiptKey string,
	datumType uint32,
) (*model.BatchReceipt, error) {
	return m.GenerateBatchReceiptWithReminderResult, m.GenerateBatchReceiptWithReminderError
}

type (
	MockMempoolGetter struct {
		MempoolGetterInterface
		GetMempoolTransactionResult *model.MempoolTransaction
		GetMempoolTransactionError  error
	}
)

func (m *MockMempoolGetter) GetMempoolTransaction(id int64) (*model.MempoolTransaction, error) {
	return m.GetMempoolTransactionResult, m.GetMempoolTransactionError
}

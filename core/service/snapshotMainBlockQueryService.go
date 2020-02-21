package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	SnapshotMainBlockQueryServiceInterface interface {
		GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error)
		GetNodeRegistrations(fromHeight, toHeight uint32) ([]*model.NodeRegistration, error)
		GetAccountDatasets(fromHeight, toHeight uint32) ([]*model.AccountDataset, error)
		GetParticipationScores(fromHeight, toHeight uint32) ([]*model.ParticipationScore, error)
		GetPublishedReceipts(fromHeight, toHeight, limit uint32) ([]*model.PublishedReceipt, error)
		GetEscrowTransactions(fromHeight, toHeight uint32) ([]*model.Escrow, error)
		InsertSnapshotPayloadToDb(payload SnapshotPayload) error
	}

	SnapshotMainBlockQueryService struct {
		QueryExecutor           query.ExecutorInterface
		Logger                  *log.Logger
		MainBlockQuery          query.BlockQueryInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		AccountDatasetQuery     query.AccountDatasetsQueryInterface
		EscrowTransactionQuery  query.EscrowTransactionQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		SnapshotQueries         map[string]query.SnapshotQuery
	}
)

// GetAccountBalances get account balances for snapshot (wrapper function around account balance query)
func (smbq *SnapshotMainBlockQueryService) GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error) {
	qry := smbq.SnapshotQueries["accountBalance"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.AccountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetNodeRegistrations get node registrations for snapshot (wrapper function around node registration query)
func (smbq *SnapshotMainBlockQueryService) GetNodeRegistrations(fromHeight, toHeight uint32) ([]*model.NodeRegistration, error) {
	qry := smbq.SnapshotQueries["nodeRegistration"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetAccountDatasets get account datasets  for snapshot (wrapper function around account dataset query)
func (smbq *SnapshotMainBlockQueryService) GetAccountDatasets(fromHeight, toHeight uint32) ([]*model.AccountDataset, error) {
	qry := smbq.SnapshotQueries["accountDataset"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.AccountDatasetQuery.BuildModel([]*model.AccountDataset{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ParticipationScores get participation scores  for snapshot (wrapper function around participationscore query)
func (smbq *SnapshotMainBlockQueryService) GetParticipationScores(fromHeight, toHeight uint32) ([]*model.ParticipationScore, error) {
	qry := smbq.SnapshotQueries["participationScore"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.ParticipationScoreQuery.BuildModel([]*model.ParticipationScore{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetPublishedReceipts get published Receipts for snapshot (wrapper function around published receipts query)
func (smbq *SnapshotMainBlockQueryService) GetPublishedReceipts(fromHeight, toHeight, limit uint32) ([]*model.PublishedReceipt, error) {
	// limit number of blocks to scan for receipts
	if toHeight-fromHeight > limit {
		fromHeight = toHeight - limit
	}
	qry := smbq.SnapshotQueries["publishedReceipt"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.PublishedReceiptQuery.BuildModel([]*model.PublishedReceipt{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetEscrowTransactions get escrowtransactions for snapshot (wrapper function around escrow transaction query)
func (smbq *SnapshotMainBlockQueryService) GetEscrowTransactions(fromHeight, toHeight uint32) ([]*model.Escrow, error) {
	qry := smbq.SnapshotQueries["escrowTransaction"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.EscrowTransactionQuery.BuildModels(rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// InsertSnapshotPayloadToDb insert snapshot data to db
func (smbq *SnapshotMainBlockQueryService) InsertSnapshotPayloadToDb(payload SnapshotPayload) error {
	var (
		queries [][]interface{}
	)

	err := smbq.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}

	for _, rec := range payload.AccountBalances {
		qry, args := smbq.AccountBalanceQuery.InsertAccountBalance(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)

	}

	for _, rec := range payload.NodeRegistrations {
		qry, args := smbq.NodeRegistrationQuery.InsertNodeRegistration(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.PublishedReceipts {
		qry, args := smbq.PublishedReceiptQuery.InsertPublishedReceipt(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.ParticipationScores {
		qry, args := smbq.ParticipationScoreQuery.InsertParticipationScore(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.EscrowTransactions {
		qryArgs := smbq.EscrowTransactionQuery.InsertEscrowTransaction(rec)
		queries = append(queries, qryArgs...)
	}

	for _, rec := range payload.AccountDatasets {
		qryArgs := smbq.AccountDatasetQuery.AddDataset(rec)
		queries = append(queries, qryArgs...)
	}

	err = smbq.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := smbq.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			smbq.Logger.Error(rollbackErr.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("fail to insert snapshot into db: %v", err))
	}
	err = smbq.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}

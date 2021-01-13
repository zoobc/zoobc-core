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
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	mockChainType struct {
		chaintype.MainChain
		SnapshotInterval          uint32
		SnapshotGenerationTimeout time.Duration
	}
)

func (mct *mockChainType) GetSnapshotInterval() uint32 {
	return mct.SnapshotInterval
}

func TestSnapshotMainBlockService_IsSnapshotHeight(t *testing.T) {
	type fields struct {
		chainType chaintype.ChainType
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_1}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: 1,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_2}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_3}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 9,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_4}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_5}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 20,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_1}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_2}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_3}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_4}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: 2 * (constant.MinRollbackBlocks + 10),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				chainType: tt.fields.chainType,
			}
			if got := ss.IsSnapshotHeight(tt.args.height); got != tt.want {
				t.Errorf("SnapshotMainBlockService.IsSnapshotHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSnapshotBasicChunkStrategy struct {
		SnapshotBasicChunkStrategy
		success bool
	}
	mockSnapshotQueryExecutor struct {
		query.Executor
		success bool
	}
	mockSnapshotAccountBalanceQuery struct {
		query.AccountBalanceQueryInterface
		success bool
	}
	mockSnapshotNodeRegistrationQuery struct {
		query.NodeRegistrationQueryInterface
		success bool
	}
	mockSnapshotAccountDatasetQuery struct {
		query.AccountDatasetQueryInterface
		success bool
	}
	mockSnapshotParticipationScoreQuery struct {
		query.ParticipationScoreQueryInterface
		success bool
	}
	mockSnapshotPublishedReceiptQuery struct {
		query.PublishedReceiptQueryInterface
		success bool
	}
	mockSnapshotEscrowTransactionQuery struct {
		query.EscrowTransactionQueryInterface
		success bool
	}
	mockSnapshotPendingTransactionQuery struct {
		query.PendingTransactionQueryInterface
		success bool
	}
	mockSnapshotPendingSignatureQuery struct {
		query.PendingSignatureQueryInterface
		success bool
	}
	mockSnapshotMultisignatureInfoQuery struct {
		query.MultisignatureInfoQueryInterface
		success bool
	}
	mockSkippedBlocksmithQuery struct {
		query.SkippedBlocksmithQueryInterface
		success bool
	}
	mockSnapshotBlockQuery struct {
		query.BlockQueryInterface
		success bool
	}
	mockSnapshotFeeScaleQuery struct {
		query.FeeScaleQueryInterface
		success bool
	}
	mockSnapshotFeeVoteCommitmentQuery struct {
		query.FeeVoteCommitmentVoteQueryInterface
		success bool
	}
	mockSnapshotFeeVoteRevealQuery struct {
		query.FeeVoteRevealVoteQueryInterface
		success bool
	}
	mockSnapshotLiquidPaymentTransactionQuery struct {
		query.LiquidPaymentTransactionQueryInterface
		success bool
	}
	mockSnapshotNodeAdmissionTimestampQuery struct {
		query.NodeAdmissionTimestampQueryInterface
		success bool
	}
	mockBlockMainServiceSuccess struct {
		BlockServiceInterface
	}
)

func (msfsq *mockSnapshotFeeScaleQuery) BuildModel([]*model.FeeScale, *sql.Rows) ([]*model.FeeScale, error) {
	if msfsq.success {
		return []*model.FeeScale{}, nil
	}
	return nil, errors.New("mockedError")
}

func (msfr *mockSnapshotFeeVoteRevealQuery) BuildModel([]*model.FeeVoteRevealVote, *sql.Rows) ([]*model.FeeVoteRevealVote, error) {
	if msfr.success {
		return []*model.FeeVoteRevealVote{}, nil
	}
	return nil, errors.New("mockedError")
}

func (msfvc *mockSnapshotFeeVoteCommitmentQuery) BuildModel(
	[]*model.FeeVoteCommitmentVote, *sql.Rows,
) ([]*model.FeeVoteCommitmentVote, error) {
	if msfvc.success {
		return []*model.FeeVoteCommitmentVote{}, nil
	}
	return nil, errors.New("mockError")
}

func (mslpt *mockSnapshotLiquidPaymentTransactionQuery) BuildModels(*sql.Rows) ([]*model.LiquidPayment, error) {
	if mslpt.success {
		return []*model.LiquidPayment{}, nil
	}
	return nil, errors.New("mockedError")
}

func (msnat *mockSnapshotNodeAdmissionTimestampQuery) BuildModel(
	[]*model.NodeAdmissionTimestamp, *sql.Rows,
) ([]*model.NodeAdmissionTimestamp, error) {
	if msnat.success {
		return []*model.NodeAdmissionTimestamp{}, nil
	}
	return nil, errors.New("mockError")
}

var (
	accBal1 = &model.AccountBalance{
		AccountAddress:   bcsAddress1,
		Balance:          10000000000,
		BlockHeight:      1,
		Latest:           true,
		PopRevenue:       100000000,
		SpendableBalance: 10000000000,
	}
	accBal2 = &model.AccountBalance{
		AccountAddress:   bcsAddress2,
		Balance:          100000000000,
		BlockHeight:      1,
		Latest:           true,
		PopRevenue:       100000000,
		SpendableBalance: 100000000000,
	}
	nr1 = &model.NodeRegistration{
		AccountAddress:     bcsAddress1,
		Latest:             true,
		Height:             0,
		LockedBalance:      10000000000,
		NodeID:             11111,
		NodePublicKey:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		RegistrationHeight: 0,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
	}
	nr2 = &model.NodeRegistration{
		AccountAddress:     bcsAddress2,
		Latest:             true,
		Height:             0,
		LockedBalance:      10000000000,
		NodeID:             22222,
		NodePublicKey:      []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		RegistrationHeight: 0,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
	}
	ps1 = &model.ParticipationScore{
		Latest: true,
		Height: 0,
		NodeID: 11111,
		Score:  1000000,
	}
	pr1 = &model.PublishedReceipt{
		BlockHeight:        1,
		IntermediateHashes: make([]byte, 32),
		PublishedIndex:     100,
		ReceiptIndex:       10,
		Receipt:            &model.Receipt{},
	}
	escrowTx1 = &model.Escrow{
		BlockHeight:      1,
		Latest:           true,
		ID:               999999,
		Amount:           1000000000,
		ApproverAddress:  bcsAddress1,
		Commission:       100000000,
		Instruction:      "test test",
		RecipientAddress: bcsAddress2,
		SenderAddress:    bcsAddress3,
		Status:           model.EscrowStatus_Pending,
		Timeout:          15875392,
	}
	accDataSet1 = &model.AccountDataset{
		Height:                  1,
		Latest:                  true,
		Property:                "testProp",
		RecipientAccountAddress: bcsAddress1,
		SetterAccountAddress:    bcsAddress2,
		Value:                   "testVal",
	}
	blockForSnapshot1 = &model.Block{
		Height:    1440,
		Timestamp: 15875392,
	}
	snapshotFullHash = []byte{24, 221, 153, 30, 107, 6, 128, 163, 98, 204, 96, 191, 126, 13, 184, 12, 200, 188, 39, 15,
		65, 111, 26, 6, 181, 130, 90, 175, 180, 57, 152, 169}
	snapshotChunk1Hash = []byte{
		1, 1, 1, 249, 145, 71, 241, 88, 208, 4, 80, 132, 88, 43, 189, 93, 19, 104, 255, 61, 177, 177, 223,
		188, 144, 9, 73, 75, 6, 1, 1, 1,
	}
	snapshotChunk2Hash = []byte{
		2, 2, 2, 249, 145, 71, 241, 88, 208, 4, 80, 132, 88, 43, 189, 93, 19, 104, 255, 61, 177, 177, 223,
		188, 144, 9, 73, 75, 6, 2, 2, 2,
	}
)

func (mkQry *mockSnapshotAccountBalanceQuery) BuildModel(accountBalances []*model.AccountBalance, rows *sql.Rows) ([]*model.AccountBalance,
	error) {
	if !mkQry.success {
		return nil, errors.New("AccountBalanceQueryFailed")
	}
	return []*model.AccountBalance{
		accBal1,
		accBal2,
	}, nil
}

func (*mockSnapshotNodeRegistrationQuery) BuildModel(noderegistrations []*model.NodeRegistration,
	rows *sql.Rows) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		nr1,
		nr2,
	}, nil
}

func (*mockSnapshotAccountDatasetQuery) BuildModel(accountDatasets []*model.AccountDataset, rows *sql.Rows) ([]*model.AccountDataset,
	error) {
	return []*model.AccountDataset{
		accDataSet1,
	}, nil
}

func (*mockSnapshotParticipationScoreQuery) BuildModel(participationScores []*model.ParticipationScore,
	rows *sql.Rows) ([]*model.ParticipationScore,
	error) {
	return []*model.ParticipationScore{
		ps1,
	}, nil
}

func (*mockSnapshotPublishedReceiptQuery) BuildModel(publishedReceipts []*model.PublishedReceipt,
	rows *sql.Rows) ([]*model.PublishedReceipt,
	error) {
	return []*model.PublishedReceipt{
		pr1,
	}, nil
}

func (*mockSnapshotPendingTransactionQuery) BuildModel(pendingTransactions []*model.PendingTransaction,
	rows *sql.Rows) ([]*model.PendingTransaction,
	error) {
	return []*model.PendingTransaction{}, nil
}

func (*mockSnapshotPendingSignatureQuery) BuildModel(pendingSignatures []*model.PendingSignature,
	rows *sql.Rows) ([]*model.PendingSignature,
	error) {
	return []*model.PendingSignature{}, nil
}

func (*mockSnapshotMultisignatureInfoQuery) BuildModel(multisignatureInfo []*model.MultiSignatureInfo,
	rows *sql.Rows) ([]*model.MultiSignatureInfo,
	error) {
	return []*model.MultiSignatureInfo{}, nil
}

func (*mockSkippedBlocksmithQuery) BuildModel(skippedBlocksmith []*model.SkippedBlocksmith,
	rows *sql.Rows) ([]*model.SkippedBlocksmith,
	error) {
	return []*model.SkippedBlocksmith{}, nil
}

func (*mockSnapshotBlockQuery) BuildModel(blocks []*model.Block,
	rows *sql.Rows) ([]*model.Block,
	error) {
	return []*model.Block{}, nil
}

func (*mockSnapshotEscrowTransactionQuery) BuildModels(*sql.Rows) ([]*model.Escrow, error) {
	return []*model.Escrow{
		escrowTx1,
	}, nil
}

func (mct *mockChainType) GetSnapshotGenerationTimeout() time.Duration {
	return mct.SnapshotGenerationTimeout
}

func (*mockSnapshotQueryExecutor) ExecuteSelect(qry string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	switch qry {
	// special case for multisig info because it has custom logic to be tested
	case "SELECT multisig_address, minimum_signatures, nonce, block_height, latest FROM multisignature_info " +
		"WHERE (multisig_address, block_height) IN (SELECT t2.multisig_address, MAX(t2.block_height) " +
		"FROM multisignature_info t2 WHERE t2.block_height >= 0 AND t2.block_height <= 720 AND t2.block_height != 0 " +
		"GROUP BY t2.multisig_address) ORDER BY block_height":
		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(query.NewMultiSignatureParticipantQuery().Fields).
				AddRow(
					address1,
					2,
					1,
					10,
					true,
				).
				AddRow(
					address2,
					3,
					2,
					11,
					true,
				))
	case "SELECT multisig_address,account_address,account_address_index,latest,block_height " +
		"FROM multisignature_participant WHERE multisig_address = ? AND block_height >= ? AND block_height <= ? " +
		"ORDER BY account_address_index":
		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(query.NewMultiSignatureParticipantQuery().Fields).
				AddRow(
					address1,
					address2,
					0,
					true,
					10,
				).
				AddRow(
					address1,
					address3,
					0,
					true,
					10,
				))
	default:
		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}))
	}
	return db.Query("")
}

func (mocksbcs *mockSnapshotBasicChunkStrategy) GenerateSnapshotChunks(
	*model.SnapshotPayload,
) (fullHash []byte, fileChunkHashes [][]byte, err error) {
	if !mocksbcs.success {
		return nil, nil, errors.New("GenerateSnapshotChunksFailed")
	}
	fileChunkHashes = [][]byte{
		snapshotChunk1Hash,
		snapshotChunk2Hash,
	}
	return snapshotFullHash, fileChunkHashes, nil
}

func (mocksbcs *mockSnapshotBasicChunkStrategy) BuildSnapshotFromChunks([]byte, [][]byte) (*model.SnapshotPayload, error) {
	if !mocksbcs.success {
		return nil, errors.New("BuildSnapshotFromChunksFailed")
	}
	return &model.SnapshotPayload{
		AccountBalances: []*model.AccountBalance{
			accBal1,
		},
		EscrowTransactions: []*model.Escrow{
			escrowTx1,
		},
		PublishedReceipts: []*model.PublishedReceipt{
			pr1,
		},
		ParticipationScores: []*model.ParticipationScore{
			ps1,
		},
		AccountDatasets: []*model.AccountDataset{
			accDataSet1,
		},
		NodeRegistrations: []*model.NodeRegistration{
			nr1,
		},
	}, nil
}

func (*mockBlockMainServiceSuccess) UpdateLastBlockCache(block *model.Block) error {
	return nil
}

func (*mockBlockMainServiceSuccess) InitializeBlocksCache() error {
	return nil
}

func (*mockBlockMainServiceSuccess) GetLastBlock() (*model.Block, error) {
	mockedBlock := transaction.GetFixturesForBlock(100, 123456789)
	return mockedBlock, nil
}

func TestSnapshotMainBlockService_NewSnapshotFile(t *testing.T) {
	type fields struct {
		SnapshotPath                   string
		chainType                      chaintype.ChainType
		Logger                         *log.Logger
		SnapshotBasicChunkStrategy     SnapshotChunkStrategyInterface
		QueryExecutor                  query.ExecutorInterface
		AccountBalanceQuery            query.AccountBalanceQueryInterface
		NodeRegistrationQuery          query.NodeRegistrationQueryInterface
		ParticipationScoreQuery        query.ParticipationScoreQueryInterface
		AccountDatasetQuery            query.AccountDatasetQueryInterface
		EscrowTransactionQuery         query.EscrowTransactionQueryInterface
		PublishedReceiptQuery          query.PublishedReceiptQueryInterface
		PendingTransactionQuery        query.PendingTransactionQueryInterface
		PendingSignatureQuery          query.PendingSignatureQueryInterface
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		SkippedBlocksmithQuery         query.SkippedBlocksmithQueryInterface
		FeeScaleQuery                  query.FeeScaleQueryInterface
		FeeVoteCommitmentVoteQuery     query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery         query.FeeVoteRevealVoteQueryInterface
		LiquidPaymentTransactionQuery  query.LiquidPaymentTransactionQueryInterface
		NodeAdmissionTimestampQuery    query.NodeAdmissionTimestampQueryInterface
		BlockQuery                     query.BlockQueryInterface
		SnapshotQueries                map[string]query.SnapshotQuery
		BlocksmithSafeQuery            map[string]bool
		DerivedQueries                 []query.DerivedQuery
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SnapshotFileInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "NewSnapshotFile:success",
			fields: fields{
				SnapshotBasicChunkStrategy: &mockSnapshotBasicChunkStrategy{
					success: true,
				},
				Logger:       log.New(),
				SnapshotPath: "testdata/snapshots",
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1 * time.Second,
				},
				QueryExecutor:                  &mockSnapshotQueryExecutor{success: true},
				AccountBalanceQuery:            &mockSnapshotAccountBalanceQuery{success: true},
				NodeRegistrationQuery:          &mockSnapshotNodeRegistrationQuery{success: true},
				ParticipationScoreQuery:        &mockSnapshotParticipationScoreQuery{success: true},
				AccountDatasetQuery:            &mockSnapshotAccountDatasetQuery{success: true},
				EscrowTransactionQuery:         &mockSnapshotEscrowTransactionQuery{success: true},
				PublishedReceiptQuery:          &mockSnapshotPublishedReceiptQuery{success: true},
				PendingTransactionQuery:        &mockSnapshotPendingTransactionQuery{success: true},
				PendingSignatureQuery:          &mockSnapshotPendingSignatureQuery{success: true},
				MultisignatureInfoQuery:        query.NewMultisignatureInfoQuery(),
				MultiSignatureParticipantQuery: query.NewMultiSignatureParticipantQuery(),
				SkippedBlocksmithQuery:         &mockSkippedBlocksmithQuery{success: true},
				BlockQuery:                     &mockSnapshotBlockQuery{success: true},
				FeeScaleQuery:                  &mockSnapshotFeeScaleQuery{success: true},
				FeeVoteCommitmentVoteQuery:     &mockSnapshotFeeVoteCommitmentQuery{success: true},
				FeeVoteRevealVoteQuery:         &mockSnapshotFeeVoteRevealQuery{success: true},
				LiquidPaymentTransactionQuery:  &mockSnapshotLiquidPaymentTransactionQuery{success: true},
				NodeAdmissionTimestampQuery:    &mockSnapshotNodeAdmissionTimestampQuery{success: true},
				SnapshotQueries:                query.GetSnapshotQuery(chaintype.GetChainType(0)),
				BlocksmithSafeQuery:            query.GetBlocksmithSafeQuery(chaintype.GetChainType(0)),
				DerivedQueries:                 query.GetDerivedQuery(chaintype.GetChainType(0)),
			},
			args: args{
				block: blockForSnapshot1,
			},
			want: &model.SnapshotFileInfo{
				SnapshotFileHash: snapshotFullHash,
				FileChunksHashes: [][]byte{
					snapshotChunk1Hash,
					snapshotChunk2Hash,
				},
				ChainType:                  0,
				Height:                     blockForSnapshot1.Height - constant.MinRollbackBlocks,
				ProcessExpirationTimestamp: blockForSnapshot1.Timestamp + 1,
				SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				SnapshotPath:                   tt.fields.SnapshotPath,
				chainType:                      tt.fields.chainType,
				Logger:                         tt.fields.Logger,
				SnapshotBasicChunkStrategy:     tt.fields.SnapshotBasicChunkStrategy,
				QueryExecutor:                  tt.fields.QueryExecutor,
				AccountBalanceQuery:            tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:          tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:        tt.fields.ParticipationScoreQuery,
				AccountDatasetQuery:            tt.fields.AccountDatasetQuery,
				EscrowTransactionQuery:         tt.fields.EscrowTransactionQuery,
				PublishedReceiptQuery:          tt.fields.PublishedReceiptQuery,
				PendingTransactionQuery:        tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:          tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery:        tt.fields.MultisignatureInfoQuery,
				MultisignatureParticipantQuery: tt.fields.MultiSignatureParticipantQuery,
				SkippedBlocksmithQuery:         tt.fields.SkippedBlocksmithQuery,
				BlockQuery:                     tt.fields.BlockQuery,
				SnapshotQueries:                tt.fields.SnapshotQueries,
				BlocksmithSafeQuery:            tt.fields.BlocksmithSafeQuery,
				FeeScaleQuery:                  tt.fields.FeeScaleQuery,
				FeeVoteCommitmentVoteQuery:     tt.fields.FeeVoteCommitmentVoteQuery,
				FeeVoteRevealVoteQuery:         tt.fields.FeeVoteRevealVoteQuery,
				LiquidPaymentTransactionQuery:  tt.fields.LiquidPaymentTransactionQuery,
				NodeAdmissionTimestampQuery:    tt.fields.NodeAdmissionTimestampQuery,
				DerivedQueries:                 tt.fields.DerivedQueries,
			}
			got, err := ss.NewSnapshotFile(tt.args.block)
			if err != nil {
				if tt.wantErr {
					if tt.errMsg != err.Error() {
						t.Errorf("error differs from what expected. wrong test exit line. gotErr %s, wantErr %s",
							err.Error(),
							tt.errMsg)
					}
					return
				}
				t.Errorf("SnapshotMainBlockService.NewSnapshotFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotMainBlockService.NewSnapshotFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSnapshotMainBlockService_Integration_NewSnapshotFile this test will generate a snapshot based on mocked data and write the file to
// disk. Then will check the file hash against the generated file and delete it.
func TestSnapshotMainBlockService_Integration_NewSnapshotFile(t *testing.T) {
	type fields struct {
		SnapshotPath                  string
		chainType                     chaintype.ChainType
		Logger                        *log.Logger
		SnapshotBasicChunkStrategy    SnapshotChunkStrategyInterface
		QueryExecutor                 query.ExecutorInterface
		AccountBalanceQuery           query.AccountBalanceQueryInterface
		NodeRegistrationQuery         query.NodeRegistrationQueryInterface
		ParticipationScoreQuery       query.ParticipationScoreQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowTransactionQuery        query.EscrowTransactionQueryInterface
		PublishedReceiptQuery         query.PublishedReceiptQueryInterface
		PendingTransactionQuery       query.PendingTransactionQueryInterface
		PendingSignatureQuery         query.PendingSignatureQueryInterface
		MultisignatureInfoQuery       query.MultisignatureInfoQueryInterface
		SkippedBlocksmithQuery        query.SkippedBlocksmithQueryInterface
		FeeScaleQuery                 query.FeeScaleQueryInterface
		FeeVoteCommitmentVoteQuery    query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery        query.FeeVoteRevealVoteQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		BlockQuery                    query.BlockQueryInterface
		NodeAdmissionTimestampQuery   query.NodeAdmissionTimestampQueryInterface
		SnapshotQueries               map[string]query.SnapshotQuery
		BlocksmithSafeQuery           map[string]bool
		DerivedQueries                []query.DerivedQuery
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte // the snapshot file hash
	}{
		{
			name: "NewSnapshotFile-IntegrationTest:success-{oneChunkFile}",
			fields: fields{
				SnapshotBasicChunkStrategy: NewSnapshotBasicChunkStrategy(
					10000000, // 10MB chunks
					NewFileService(
						log.New(),
						new(codec.CborHandle),
						"testdata/snapshots",
					),
				),
				Logger:       log.New(),
				SnapshotPath: "testdata/snapshots",
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 10,
				},
				QueryExecutor:                 &mockSnapshotQueryExecutor{success: true},
				AccountBalanceQuery:           &mockSnapshotAccountBalanceQuery{success: true},
				NodeRegistrationQuery:         &mockSnapshotNodeRegistrationQuery{success: true},
				ParticipationScoreQuery:       &mockSnapshotParticipationScoreQuery{success: true},
				AccountDatasetQuery:           &mockSnapshotAccountDatasetQuery{success: true},
				EscrowTransactionQuery:        &mockSnapshotEscrowTransactionQuery{success: true},
				PublishedReceiptQuery:         &mockSnapshotPublishedReceiptQuery{success: true},
				PendingTransactionQuery:       &mockSnapshotPendingTransactionQuery{success: true},
				PendingSignatureQuery:         &mockSnapshotPendingSignatureQuery{success: true},
				MultisignatureInfoQuery:       &mockSnapshotMultisignatureInfoQuery{success: true},
				SkippedBlocksmithQuery:        &mockSkippedBlocksmithQuery{success: true},
				BlockQuery:                    &mockSnapshotBlockQuery{success: true},
				FeeScaleQuery:                 &mockSnapshotFeeScaleQuery{success: true},
				FeeVoteCommitmentVoteQuery:    &mockSnapshotFeeVoteCommitmentQuery{success: true},
				FeeVoteRevealVoteQuery:        &mockSnapshotFeeVoteRevealQuery{success: true},
				LiquidPaymentTransactionQuery: &mockSnapshotLiquidPaymentTransactionQuery{success: true},
				NodeAdmissionTimestampQuery:   &mockSnapshotNodeAdmissionTimestampQuery{success: true},
				SnapshotQueries:               query.GetSnapshotQuery(chaintype.GetChainType(0)),
				DerivedQueries:                query.GetDerivedQuery(chaintype.GetChainType(0)),
				BlocksmithSafeQuery:           query.GetBlocksmithSafeQuery(chaintype.GetChainType(0)),
			},
			args: args{
				block: blockForSnapshot1,
			},
			want: snapshotFullHash,
		},
		{
			name: "NewSnapshotFile-IntegrationTest:success-{multiChunksFile}",
			fields: fields{
				SnapshotBasicChunkStrategy: NewSnapshotBasicChunkStrategy(
					1000, // 1000 Bytes chunks
					NewFileService(
						log.New(),
						new(codec.CborHandle),
						"testdata/snapshots",
					),
				),
				Logger:       log.New(),
				SnapshotPath: "testdata/snapshots",
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 10,
				},
				QueryExecutor:                 &mockSnapshotQueryExecutor{success: true},
				AccountBalanceQuery:           &mockSnapshotAccountBalanceQuery{success: true},
				NodeRegistrationQuery:         &mockSnapshotNodeRegistrationQuery{success: true},
				ParticipationScoreQuery:       &mockSnapshotParticipationScoreQuery{success: true},
				AccountDatasetQuery:           &mockSnapshotAccountDatasetQuery{success: true},
				EscrowTransactionQuery:        &mockSnapshotEscrowTransactionQuery{success: true},
				PublishedReceiptQuery:         &mockSnapshotPublishedReceiptQuery{success: true},
				PendingTransactionQuery:       &mockSnapshotPendingTransactionQuery{success: true},
				PendingSignatureQuery:         &mockSnapshotPendingSignatureQuery{success: true},
				MultisignatureInfoQuery:       &mockSnapshotMultisignatureInfoQuery{success: true},
				SkippedBlocksmithQuery:        &mockSkippedBlocksmithQuery{success: true},
				BlockQuery:                    &mockSnapshotBlockQuery{success: true},
				FeeScaleQuery:                 &mockSnapshotFeeScaleQuery{success: true},
				FeeVoteCommitmentVoteQuery:    &mockSnapshotFeeVoteCommitmentQuery{success: true},
				FeeVoteRevealVoteQuery:        &mockSnapshotFeeVoteRevealQuery{success: true},
				LiquidPaymentTransactionQuery: &mockSnapshotLiquidPaymentTransactionQuery{success: true},
				NodeAdmissionTimestampQuery:   &mockSnapshotNodeAdmissionTimestampQuery{success: true},
				SnapshotQueries:               query.GetSnapshotQuery(chaintype.GetChainType(0)),
				DerivedQueries:                query.GetDerivedQuery(chaintype.GetChainType(0)),
				BlocksmithSafeQuery:           query.GetBlocksmithSafeQuery(chaintype.GetChainType(0)),
			},
			args: args{
				block: blockForSnapshot1,
			},
			want: snapshotFullHash,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				SnapshotPath:                  tt.fields.SnapshotPath,
				chainType:                     tt.fields.chainType,
				Logger:                        tt.fields.Logger,
				SnapshotBasicChunkStrategy:    tt.fields.SnapshotBasicChunkStrategy,
				QueryExecutor:                 tt.fields.QueryExecutor,
				AccountBalanceQuery:           tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:         tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:       tt.fields.ParticipationScoreQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowTransactionQuery:        tt.fields.EscrowTransactionQuery,
				PublishedReceiptQuery:         tt.fields.PublishedReceiptQuery,
				PendingTransactionQuery:       tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:         tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery:       tt.fields.MultisignatureInfoQuery,
				SkippedBlocksmithQuery:        tt.fields.SkippedBlocksmithQuery,
				BlockQuery:                    tt.fields.BlockQuery,
				SnapshotQueries:               tt.fields.SnapshotQueries,
				DerivedQueries:                tt.fields.DerivedQueries,
				BlocksmithSafeQuery:           tt.fields.BlocksmithSafeQuery,
				FeeScaleQuery:                 tt.fields.FeeScaleQuery,
				FeeVoteCommitmentVoteQuery:    tt.fields.FeeVoteCommitmentVoteQuery,
				FeeVoteRevealVoteQuery:        tt.fields.FeeVoteRevealVoteQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				NodeAdmissionTimestampQuery:   tt.fields.NodeAdmissionTimestampQuery,
			}
			got, err := ss.NewSnapshotFile(tt.args.block)
			if err != nil {
				t.Errorf("SnapshotMainBlockService.NewSnapshotFile() error = %v", err)
				return
			}
			// this is the hash of encoded bynary data
			if !reflect.DeepEqual(got.SnapshotFileHash, tt.want) {
				t.Errorf("SnapshotMainBlockService.NewSnapshotFile() = \n%v, want \n%v", got.SnapshotFileHash, tt.want)
			}
			// remove generated files
			s1 := "3puTLlMoE9A3u5ykop5G-TWDt5lDWS-9zybgH3N896E="
			_ = os.Remove(filepath.Join(tt.fields.SnapshotPath, s1))
			s2 := "jica4f9TBxknRQC_gDcd83OMRno9SkmIPBJQbyjK2F8="
			_ = os.Remove(filepath.Join(tt.fields.SnapshotPath, s2))
			s3 := "JWx5HOAgG11sFIAHVF-G1dtveG4iIm5K7VoZsxrBlOw="
			_ = os.Remove(filepath.Join(tt.fields.SnapshotPath, s3))
		})
	}
}

func (*mockSnapshotQueryExecutor) BeginTx(bool) error {
	return nil
}

func (*mockSnapshotQueryExecutor) CommitTx(bool) error {
	return nil
}

func (*mockSnapshotQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockSnapshotQueryExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

type (
	mockImportSnapshotFileNodeRegistrationServiceSuccess struct {
		NodeRegistrationServiceInterface
	}
	mockScrambleNodeServiceInitSuccess struct {
		ScrambleNodeService
	}
)

func (*mockScrambleNodeServiceInitSuccess) InitializeScrambleCache(lastBlockHeight uint32) error {
	return nil
}

func (*mockImportSnapshotFileNodeRegistrationServiceSuccess) UpdateNextNodeAdmissionCache(
	newNextNodeAdmission *model.NodeAdmissionTimestamp) error {
	return nil
}

func (*mockImportSnapshotFileNodeRegistrationServiceSuccess) InitializeCache() error {
	return nil
}

func TestSnapshotMainBlockService_ImportSnapshotFile(t *testing.T) {
	type fields struct {
		SnapshotPath                   string
		chainType                      chaintype.ChainType
		Logger                         *log.Logger
		SnapshotBasicChunkStrategy     SnapshotChunkStrategyInterface
		QueryExecutor                  query.ExecutorInterface
		AccountBalanceQuery            query.AccountBalanceQueryInterface
		NodeRegistrationQuery          query.NodeRegistrationQueryInterface
		ParticipationScoreQuery        query.ParticipationScoreQueryInterface
		AccountDatasetQuery            query.AccountDatasetQueryInterface
		EscrowTransactionQuery         query.EscrowTransactionQueryInterface
		PublishedReceiptQuery          query.PublishedReceiptQueryInterface
		PendingTransactionQuery        query.PendingTransactionQueryInterface
		PendingSignatureQuery          query.PendingSignatureQueryInterface
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		SkippedBlocksmithQuery         query.SkippedBlocksmithQueryInterface
		FeeScaleQuery                  query.FeeScaleQueryInterface
		FeeVoteCommitmentVoteQuery     query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery         query.FeeVoteRevealVoteQueryInterface
		LiquidPaymentTransactionQuery  query.LiquidPaymentTransactionQueryInterface
		NodeAdmissionTimestampQuery    query.NodeAdmissionTimestampQueryInterface
		BlockQuery                     query.BlockQueryInterface
		SnapshotQueries                map[string]query.SnapshotQuery
		BlocksmithSafeQuery            map[string]bool
		DerivedQueries                 []query.DerivedQuery
		TransactionUtil                transaction.UtilInterface
		TypeActionSwitcher             transaction.TypeActionSwitcher
		BlockMainService               BlockServiceInterface
		NodeRegistrationService        NodeRegistrationServiceInterface
		ScrambleNodeService            ScrambleNodeServiceInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errMsg  string
	}{
		{
			name: "ParseSnapshotFile_IntegrationTest:success",
			fields: fields{
				SnapshotBasicChunkStrategy: &mockSnapshotBasicChunkStrategy{
					success: true,
				},
				Logger:       log.New(),
				SnapshotPath: "testdata/snapshots",
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 10,
				},
				QueryExecutor:                  &mockSnapshotQueryExecutor{success: true},
				AccountBalanceQuery:            query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:          query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery:        query.NewParticipationScoreQuery(),
				AccountDatasetQuery:            query.NewAccountDatasetsQuery(),
				EscrowTransactionQuery:         query.NewEscrowTransactionQuery(),
				PublishedReceiptQuery:          query.NewPublishedReceiptQuery(),
				PendingTransactionQuery:        query.NewPendingTransactionQuery(),
				PendingSignatureQuery:          query.NewPendingSignatureQuery(),
				MultisignatureInfoQuery:        query.NewMultisignatureInfoQuery(),
				MultiSignatureParticipantQuery: query.NewMultiSignatureParticipantQuery(),
				SkippedBlocksmithQuery:         query.NewSkippedBlocksmithQuery(&chaintype.MainChain{}),
				FeeScaleQuery:                  query.NewFeeScaleQuery(),
				FeeVoteCommitmentVoteQuery:     query.NewFeeVoteCommitmentVoteQuery(),
				FeeVoteRevealVoteQuery:         query.NewFeeVoteRevealVoteQuery(),
				LiquidPaymentTransactionQuery:  query.NewLiquidPaymentTransactionQuery(),
				BlockQuery:                     query.NewBlockQuery(&chaintype.MainChain{}),
				NodeAdmissionTimestampQuery:    query.NewNodeAdmissionTimestampQuery(),
				SnapshotQueries:                query.GetSnapshotQuery(chaintype.GetChainType(0)),
				BlocksmithSafeQuery:            query.GetBlocksmithSafeQuery(chaintype.GetChainType(0)),
				DerivedQueries:                 query.GetDerivedQuery(chaintype.GetChainType(0)),
				TransactionUtil:                &transaction.Util{},
				TypeActionSwitcher: &transaction.TypeSwitcher{
					Executor: &mockSnapshotQueryExecutor{success: true},
				},
				BlockMainService:        &mockBlockMainServiceSuccess{},
				NodeRegistrationService: &mockImportSnapshotFileNodeRegistrationServiceSuccess{},
				ScrambleNodeService:     &mockScrambleNodeServiceInitSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				SnapshotPath:                   tt.fields.SnapshotPath,
				chainType:                      tt.fields.chainType,
				Logger:                         tt.fields.Logger,
				SnapshotBasicChunkStrategy:     tt.fields.SnapshotBasicChunkStrategy,
				QueryExecutor:                  tt.fields.QueryExecutor,
				AccountBalanceQuery:            tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:          tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:        tt.fields.ParticipationScoreQuery,
				AccountDatasetQuery:            tt.fields.AccountDatasetQuery,
				EscrowTransactionQuery:         tt.fields.EscrowTransactionQuery,
				PublishedReceiptQuery:          tt.fields.PublishedReceiptQuery,
				PendingTransactionQuery:        tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:          tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery:        tt.fields.MultisignatureInfoQuery,
				MultisignatureParticipantQuery: tt.fields.MultiSignatureParticipantQuery,
				SkippedBlocksmithQuery:         tt.fields.SkippedBlocksmithQuery,
				FeeScaleQuery:                  tt.fields.FeeScaleQuery,
				FeeVoteCommitmentVoteQuery:     tt.fields.FeeVoteCommitmentVoteQuery,
				FeeVoteRevealVoteQuery:         tt.fields.FeeVoteRevealVoteQuery,
				LiquidPaymentTransactionQuery:  tt.fields.LiquidPaymentTransactionQuery,
				BlockQuery:                     tt.fields.BlockQuery,
				NodeAdmissionTimestampQuery:    tt.fields.NodeAdmissionTimestampQuery,
				SnapshotQueries:                tt.fields.SnapshotQueries,
				BlocksmithSafeQuery:            tt.fields.BlocksmithSafeQuery,
				DerivedQueries:                 tt.fields.DerivedQueries,
				TransactionUtil:                tt.fields.TransactionUtil,
				TypeActionSwitcher:             tt.fields.TypeActionSwitcher,
				BlockMainService:               tt.fields.BlockMainService,
				NodeRegistrationService:        tt.fields.NodeRegistrationService,
				ScrambleNodeService:            tt.fields.ScrambleNodeService,
			}
			snapshotFileInfo, err := ss.NewSnapshotFile(blockForSnapshot1)
			if err != nil {
				t.Errorf("SnapshotMainBlockService.ImportSnapshotFile() error creating snapshots: %v", err)
				return
			}
			if err := ss.ImportSnapshotFile(snapshotFileInfo); err != nil {
				if tt.wantErr {
					if tt.errMsg != err.Error() {
						t.Errorf("error differs from what expected. wrong test exit line. gotErr %s, wantErr %s",
							err.Error(),
							tt.errMsg)
					}
					return
				}
				t.Errorf("SnapshotMainBlockService.ImportSnapshotFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

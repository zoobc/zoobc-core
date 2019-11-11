package auth

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorValidateSuccess struct {
		query.Executor
	}
)

func (*mockExecutorValidateSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_address = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountAddress",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			"BCZ",
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty,"+
		" smith_scale, payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version"+
		" FROM main_block ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"block_hash",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"smith_scale",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		}).AddRow(
			0,
			[]byte{204, 49, 74, 114, 213, 94, 136, 109, 130, 252, 182, 25, 31, 49, 83, 238,
				65, 207, 202, 144, 95, 176, 46, 69, 165, 6, 135, 93, 217, 0, 101, 18},
			[]byte{},
			1,
			1562806389280,
			[]byte{},
			[]byte{},
			100000000,
			1,
			0,
			[]byte{},
			nodePubKey1,
			100000000,
			10000000,
			1,
			0,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty,"+
		" smith_scale, payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version"+
		" FROM main_block WHERE height = 0" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"block_hash",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"smith_scale",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		}).AddRow(
			0,
			[]byte{204, 49, 74, 114, 213, 94, 136, 109, 130, 252, 182, 25, 31, 49, 83, 238,
				65, 207, 202, 144, 95, 176, 46, 69, 165, 6, 135, 93, 217, 0, 101, 18},
			[]byte{},
			1,
			1562806389280,
			[]byte{},
			[]byte{},
			100000000,
			1,
			0,
			[]byte{},
			nodePubKey1,
			100000000,
			10000000,
			1,
			0,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("A")
	}
	return nil, nil
}
func TestProofOfOwnershipValidation_ValidateProofOfOwnership(t *testing.T) {
	poown := GetFixturesProofOfOwnershipValidation(0, nil, nil)
	type args struct {
		poown         *model.ProofOfOwnership
		nodePublicKey []byte
		queryExecutor query.ExecutorInterface
		blockQuery    query.BlockQueryInterface
	}
	poownInvalidSignature := GetFixturesProofOfOwnershipValidation(0, nil, nil)
	poownInvalidSignature.Signature = []byte{41, 7, 108, 68, 19, 119, 1, 128, 65, 227, 181, 177,
		137, 219, 248, 111, 54, 166, 110, 77, 164, 196, 19, 178, 152, 106, 199, 184,
		220, 8, 90, 171, 165, 229, 238, 235, 181, 89, 60, 28, 124, 22, 201, 237, 143,
		63, 59, 156, 133, 194, 189, 97, 150, 245, 96, 45, 192, 236, 109, 80, 14, 31, 243, 10}
	poownInvalidMessage := &model.ProofOfOwnership{
		Signature: []byte{69, 237, 231, 113, 208, 107, 56, 109, 104, 211, 67, 117, 63, 55, 237,
			243, 249, 78, 34, 90, 183, 37, 212, 42, 219, 45, 45, 247, 151, 129, 222, 244, 210,
			185, 54, 184, 17, 214, 72, 231, 195, 159, 171, 184, 73, 193, 84, 224, 51, 37, 139,
			70, 237, 153, 122, 67, 247, 182, 141, 51, 168, 53, 125, 0},
		MessageBytes: []byte{41, 7, 108, 68, 19, 119, 1, 128, 65, 227, 181, 177,
			137, 219, 248, 111, 54, 166, 110, 77, 164, 196, 19, 178, 152, 106, 199, 184,
			220, 8, 90, 171, 165, 229, 238, 235, 181, 89, 60, 28, 124, 22, 201, 237, 143,
			63, 59, 156, 133, 194, 189, 97, 150, 245, 96, 45, 192, 236, 109, 80, 14, 31, 243, 10,
		},
	}
	poownBlockHeightExpired := GetFixturesProofOfOwnershipValidation(101, nil, nil)
	poownBlockInvalidBlockHash := GetFixturesProofOfOwnershipValidation(0, nil, &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		SmithScale:           1,
		PayloadLength:        0,
		PayloadHash:          []byte{0, 0, 0, 1},
		BlocksmithPublicKey:  nodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	})

	tests := []struct {
		name    string
		p       *ProofOfOwnershipValidation
		args    args
		wantErr bool
	}{
		{
			name: "Validate:success",
			args: args{
				poown:         poown,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{InvalidSignature}",
			args: args{
				poown:         poownInvalidSignature,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{InvalidMessageBytes}",
			args: args{
				poown:         poownInvalidMessage,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{BlockHeightExpired}",
			args: args{
				poown:         poownBlockHeightExpired,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{InvalidBlockHash}",
			args: args{
				poown:         poownBlockInvalidBlockHash,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProofOfOwnershipValidation{}
			err := p.ValidateProofOfOwnership(tt.args.poown, tt.args.nodePublicKey,
				tt.args.queryExecutor, tt.args.blockQuery)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("ProofOfOwnershipValidation.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

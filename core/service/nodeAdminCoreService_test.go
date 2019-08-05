package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
)

type (
	spyNodeAdminCoreServiceHelper struct {
		NodeAdminService
	}
	nodeAdminMockQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*nodeAdminMockQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version FROM main_block ORDER BY " +
		"height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	default:
		return nil, errors.New("QueryNotMocked")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func TestNodeAdminService_GenerateProofOfOwnership(t *testing.T) {
	if err := commonUtil.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	}
	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
		Helpers       NodeAdminServiceHelpersInterface
	}
	type args struct {
		accountType    uint32
		accountAddress string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ProofOfOwnership
		wantErr bool
	}{
		{
			name: "GenerateProofOfOwnership:Success",
			fields: fields{
				QueryExecutor: &nodeAdminMockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				AccountQuery:  nil,
				Signature:     nil,
				Helpers:       &spyNodeAdminCoreServiceHelper{},
			},
			args: args{
				accountType:    1,
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			want: &model.ProofOfOwnership{
				MessageBytes: []byte{8, 1, 18, 44, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57, 90, 83,
					52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108, 69, 26, 64, 28,
					67, 36, 177, 18, 86, 20, 83, 73, 100, 118, 236, 245, 79, 57, 156, 69, 220, 166, 222, 128, 172, 119, 169, 85, 168,
					111, 124, 143, 109, 18, 226, 91, 149, 235, 82, 49, 204, 97, 180, 91, 82, 40, 9, 12, 15, 94, 49, 245, 175, 150, 243,
					217, 140, 133, 89, 117, 200, 193, 235, 101, 145, 8, 195, 32, 1},
				Signature: []byte{108, 140, 82, 187, 173, 96, 10, 233, 75, 1, 197, 35, 205, 247, 142, 133, 132, 30, 225, 39, 90, 155,
					248, 131, 54, 19, 80, 223, 60, 11, 31, 154, 84, 95, 86, 54, 228, 76, 222, 144, 3, 225, 226, 219, 29, 117, 17, 137,
					47, 249, 49, 166, 12, 172, 98, 227, 167, 69, 188, 18, 177, 223, 159, 3},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				AccountQuery:  tt.fields.AccountQuery,
				Signature:     tt.fields.Signature,
				Helpers:       tt.fields.Helpers,
			}
			got, err := nas.GenerateProofOfOwnership(tt.args.accountType, tt.args.accountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.GenerateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.GenerateProofOfOwnership() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdminService_ValidateProofOfOwnership(t *testing.T) {
	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
		Helpers       NodeAdminServiceHelpersInterface
	}
	type args struct {
		poown         *model.ProofOfOwnership
		nodePublicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidateProofOfOwnership:Success",
			fields: fields{
				QueryExecutor: &nodeAdminMockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				AccountQuery:  nil,
				Signature:     nil,
				Helpers:       &spyNodeAdminCoreServiceHelper{},
			},
			args: args{
				poown: &model.ProofOfOwnership{
					MessageBytes: []byte{8, 1, 18, 44, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57, 90, 83,
						52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108, 69, 26, 64, 28,
						67, 36, 177, 18, 86, 20, 83, 73, 100, 118, 236, 245, 79, 57, 156, 69, 220, 166, 222, 128, 172, 119, 169, 85, 168,
						111, 124, 143, 109, 18, 226, 91, 149, 235, 82, 49, 204, 97, 180, 91, 82, 40, 9, 12, 15, 94, 49, 245, 175, 150, 243,
						217, 140, 133, 89, 117, 200, 193, 235, 101, 145, 8, 195, 32, 1},
					Signature: []byte{108, 140, 82, 187, 173, 96, 10, 233, 75, 1, 197, 35, 205, 247, 142, 133, 132, 30, 225, 39, 90, 155,
						248, 131, 54, 19, 80, 223, 60, 11, 31, 154, 84, 95, 86, 54, 228, 76, 222, 144, 3, 225, 226, 219, 29, 117, 17, 137,
						47, 249, 49, 166, 12, 172, 98, 227, 167, 69, 188, 18, 177, 223, 159, 3},
				},
				nodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242,
					244, 100, 134, 144, 246, 37, 144, 213, 135},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				AccountQuery:  tt.fields.AccountQuery,
				Signature:     tt.fields.Signature,
				Helpers:       tt.fields.Helpers,
			}
			if err := nas.ValidateProofOfOwnership(tt.args.poown, tt.args.nodePublicKey); (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

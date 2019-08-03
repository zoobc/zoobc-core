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
)

type (
	spyNodeAdminCoreServiceHelper     struct{}
	nodeAdminMockQueryExecutorSuccess struct {
		query.Executor
	}
)

// var (
// 	lastBlockHash []byte = []byte{28, 67, 36, 177, 18, 86, 20, 83, 73, 100, 118, 236, 245, 79, 57, 156, 69, 220, 166,
// 		222, 128, 172, 119, 169, 85, 168, 111, 124, 143, 109, 18, 226, 91, 149, 235, 82, 49, 204, 97, 180, 91, 82, 40,
// 		9, 12, 15, 94, 49, 245, 175, 150, 243, 217, 140, 133, 89, 117, 200, 193, 235, 101, 145, 8, 195}
// )

func (*spyNodeAdminCoreServiceHelper) LoadOwnerAccountFromConfig() (ownerAccountType uint32, ownerAccountAddress string, err error) {
	return 0, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", nil
}

func (*spyNodeAdminCoreServiceHelper) LoadNodeSeedFromConfig() (nodeSeed string, err error) {
	return "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness", nil
}

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
	default:
		return nil, errors.New("QueryNotMocked")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

// func TestValidateProofOfOwnership(t *testing.T) {

// 	type paramsStruct struct {
// 		nodeMessages []byte
// 		signature    []byte
// 		publicKey    []byte
// 	}

// 	type wantStruct struct {
// 		err error
// 	}

// 	type fields struct {
// 		QueryExecutor query.ExecutorInterface
// 		BlockQuery    query.BlockQueryInterface
// 		AccountQuery  query.AccountQueryInterface
// 		Signature     crypto.SignatureInterface
// 	}

// 	tests := []struct {
// 		name   string
// 		fields fields
// 		params *paramsStruct
// 		want   *wantStruct
// 	}{
// 		{
// 			name: "Validate Proof Of Ownership",
// 			fields: fields{
// 				QueryExecutor: &mockQueryExecutorSuccess{},
// 				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
// 				AccountQuery:  nil,
// 				Signature:     nil,
// 			},
// 			params: &paramsStruct{
// 				nodeMessages: []byte{1, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86,
// 					102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75,
// 					108, 69, 28, 67, 36, 177, 18, 86, 20, 83, 73, 100, 118, 236, 245, 79, 57, 156, 69, 220, 166, 222, 128, 172,
// 					119, 169, 85, 168, 111, 124, 143, 109, 18, 226, 91, 149, 235, 82, 49, 204, 97, 180, 91, 82, 40, 9, 12, 15, 94,
// 					49, 245, 175, 150, 243, 217, 140, 133, 89, 117, 200, 193, 235, 101, 145, 8, 195, 1, 0, 0, 0},
// 				signature: []byte{115, 74, 30, 212, 221, 118, 106, 246, 87, 93, 149, 146, 141, 111, 100, 45, 29, 48, 16, 212, 236,
// 					60, 30, 50, 73, 134, 217, 91, 220, 41, 69, 7, 44, 181, 253, 159, 156, 174, 68, 206, 19, 51, 47, 211, 90, 100,
// 					38, 32, 178, 155, 204, 215, 194, 5, 109, 251, 106, 118, 238, 8, 24, 127, 170, 4},
// 				publicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
// 					239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
// 			},
// 			want: &wantStruct{
// 				err: nil,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			nodeAdminService := &NodeAdminService{
// 				QueryExecutor: tt.fields.QueryExecutor,
// 				BlockQuery:    tt.fields.BlockQuery,
// 				AccountQuery:  tt.fields.AccountQuery,
// 				Signature:     tt.fields.Signature,
// 			}
// 			res := nodeAdminService.ValidateProofOfOwnership(tt.params.nodeMessages, tt.params.signature, tt.params.publicKey)

// 			if res != tt.want.err {
// 				t.Errorf("Validate proof of ownership \ngot = %v, \nwant = %v", res, tt.want.err)
// 				return
// 			}

// 		})
// 	}
// }

func TestNodeAdminService_GenerateProofOfOwnership(t *testing.T) {
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

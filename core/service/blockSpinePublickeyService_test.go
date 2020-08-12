package service

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"regexp"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockNodeRegistrationQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockNodeRegistrationQueryExecutorSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE height >= (SELECT MIN(height) " +
		"FROM main_block AS mb1 WHERE mb1.timestamp >= 1) AND height <= (SELECT MAX(height) " +
		"FROM main_block AS mb2 WHERE mb2.timestamp < 2) AND registration_status != 1 AND latest=1 ORDER BY height":
		mockNodeRegistrationRows := mockSpine.NewRows(query.NewNodeRegistrationQuery().Fields)
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockNodeRegistrationRows)
	default:
		return nil, fmt.Errorf("unmocked query for mockNodeRegistrationQueryExecutorSuccess: %s", qStr)
	}
	rows, _ := db.Query(qStr)
	return rows, nil
}

func TestBlockSpinePublicKeyService_BuildSpinePublicKeysFromNodeRegistry(t *testing.T) {
	type fields struct {
		Signature             crypto.SignatureInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
	}
	type args struct {
		fromTimestamp int64
		toTimestamp   int64
		spineHeight   uint32
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantSpinePublicKeys []*model.SpinePublicKey
		wantErr             bool
	}{
		{
			name: "BuildSpinePublicKeysFromNodeRegistry:success",
			fields: fields{
				QueryExecutor:         &mockNodeRegistrationQueryExecutorSuccess{},
				SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				Signature:             nil,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args: args{
				fromTimestamp: 1,
				toTimestamp:   2,
				spineHeight:   1,
			},
			wantSpinePublicKeys: []*model.SpinePublicKey{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bsf := &BlockSpinePublicKeyService{
				Signature:             tt.fields.Signature,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
			}
			gotSpinePublicKeys, err := bsf.BuildSpinePublicKeysFromNodeRegistry(tt.args.fromTimestamp, tt.args.toTimestamp, tt.args.spineHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpinePublicKeyService.BuildSpinePublicKeysFromNodeRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSpinePublicKeys, tt.wantSpinePublicKeys) {
				t.Errorf("BlockSpinePublicKeyService.BuildSpinePublicKeysFromNodeRegistry() = %v, want %v", gotSpinePublicKeys, tt.wantSpinePublicKeys)
			}
		})
	}
}

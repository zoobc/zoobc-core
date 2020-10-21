package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockBlockchainObjectQuery = NewBlockchainObjectQuery()
	mockBlockchainObject      = &model.BlockchainObject{
		ID:                  []byte{1, 2},
		OwnerAccountAddress: []byte{1, 2},
		BlockHeight:         12,
	}
)

func TestNewBlockchainObjectQuery(t *testing.T) {
	tests := []struct {
		name string
		want *BlockchainObjectQuery
	}{
		{
			name: "wantSuccess",
			want: mockBlockchainObjectQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchainObjectQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainObjectQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectQuery_InsertBlockcahinObject(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockchainObject *model.BlockchainObject
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectQuery),
			args: args{
				blockchainObject: mockBlockchainObject,
			},
			wantStr: "INSERT INTO blockchain_object (id, owner, block_height) VALUES(? , ?, ?)",
			wantArgs: []interface{}{
				mockBlockchainObject.ID,
				mockBlockchainObject.OwnerAccountAddress,
				mockBlockchainObject.BlockHeight,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := boq.InsertBlockcahinObject(tt.args.blockchainObject)
			if gotStr != tt.wantStr {
				t.Errorf("BlockchainObjectQuery.InsertBlockcahinObject() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BlockchainObjectQuery.InsertBlockcahinObject() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

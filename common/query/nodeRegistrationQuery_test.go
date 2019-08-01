package query

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
)

var mockNodeRegistrationQuery = &NodeRegistrationQuery{
	Fields: []string{"node_public_key", "account_id", "registration_height", "node_address", "locked_balance", "queued",
		"latest", "height"},
	TableName: "node_registry",
}

var mockNodeRegistry = &model.NodeRegistration{
	NodePublicKey:      []byte{1},
	AccountId:          []byte{2},
	RegistrationHeight: 1,
	NodeAddress:        "127.0.0.1",
	LockedBalance:      10000,
	Queued:             true,
	Latest:             true,
	Height:             0,
}

func TestNewNodeRegistrationQuery(t *testing.T) {
	tests := []struct {
		name string
		want *NodeRegistrationQuery
	}{
		{
			name: "NewNodeRegistrationQuery:success",
			want: mockNodeRegistrationQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeRegistrationQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeRegistrationQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistrationQuery_getTableName(t *testing.T) {
	t.Run("NodeRegistrationQuery", func(t *testing.T) {
		if mockNodeRegistrationQuery.getTableName() != mockNodeRegistrationQuery.TableName {
			t.Error("gettableName not returning tablename")
		}
	})
}

func TestNodeRegistrationQuery_InsertNodeRegistration(t *testing.T) {
	t.Run("InsertNodeRegistration:success", func(t *testing.T) {

		q, args := mockNodeRegistrationQuery.InsertNodeRegistration(mockNodeRegistry)
		wantQ := "INSERT INTO node_registry (node_public_key,account_id,registration_height,node_address," +
			"locked_balance,queued,latest,height) VALUES(? , ?, ?, ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountId, mockNodeRegistry.RegistrationHeight,
			mockNodeRegistry.NodeAddress, mockNodeRegistry.LockedBalance, mockNodeRegistry.Queued,
			mockNodeRegistry.Latest, mockNodeRegistry.Height,
		}
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrations(t *testing.T) {
	t.Run("GetNodeRegistrations", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrations(0, 2)
		want := "SELECT node_public_key, account_id, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE height >= 0 AND latest=1 LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByNodePublicKey(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetNodeRegistrationByNodePublicKey([]byte{1})
		want := "SELECT node_public_key, account_id, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1"
		wantArg := []interface{}{[]byte{1}}
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if !reflect.DeepEqual(arg, wantArg) {
			t.Errorf("argument not match:\nget: %v\nwant: %v", arg, wantArg)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByAccountPublicKey(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetNodeRegistrationByAccountPublicKey([]byte{1})
		want := "SELECT node_public_key, account_id, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE account_id = [1] AND latest=1"
		wantArg := []interface{}{[]byte{1}}
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if !reflect.DeepEqual(arg, wantArg) {
			t.Errorf("argument not match:\nget: %v\nwant: %v", arg, wantArg)
		}
	})
}

func TestNodeRegistrationQuery_ExtractModel(t *testing.T) {
	t.Run("NodeRegistration:ExtractModel:success", func(t *testing.T) {
		res := mockNodeRegistrationQuery.ExtractModel(mockNodeRegistry)
		want := []interface{}{
			mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountId, mockNodeRegistry.RegistrationHeight,
			mockNodeRegistry.NodeAddress, mockNodeRegistry.LockedBalance, mockNodeRegistry.Queued,
			mockNodeRegistry.Latest, mockNodeRegistry.Height,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}

func TestNodeRegistrationQuery_BuildModel(t *testing.T) {
	t.Run("NodeRegistrationQuery-BuildModel:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows([]string{
			"NodePublicKey", "AccountId", "RegistrationHeight", "NodeAddress", "LockedBalance",
			"Queued", "Latest", "Height"}).
			AddRow(mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountId, mockNodeRegistry.RegistrationHeight,
				mockNodeRegistry.NodeAddress, mockNodeRegistry.LockedBalance, mockNodeRegistry.Queued,
				mockNodeRegistry.Latest, mockNodeRegistry.Height))
		rows, _ := db.Query("foo")
		var tempNode []*model.NodeRegistration
		res := mockNodeRegistrationQuery.BuildModel(tempNode, rows)
		if !reflect.DeepEqual(res[0], mockNodeRegistry) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockNodeRegistry)
		}
	})
}

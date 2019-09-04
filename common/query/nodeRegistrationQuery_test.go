package query

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockNodeRegistrationQuery = NewNodeRegistrationQuery()
	mockNodeRegistry          = &model.NodeRegistration{
		NodeID:             1,
		NodePublicKey:      []byte{1},
		AccountAddress:     "BCZ",
		RegistrationHeight: 1,
		NodeAddress:        "127.0.0.1",
		LockedBalance:      10000,
		Queued:             true,
		Latest:             true,
		Height:             0,
	}
)

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
		wantQ := "INSERT INTO node_registry (id,node_public_key,account_address,registration_height,node_address," +
			"locked_balance,queued,latest,height) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockNodeRegistry.NodeID, mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountAddress, mockNodeRegistry.RegistrationHeight,
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
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE height >= 0 AND latest=1 LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByNodePublicKey(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetNodeRegistrationByNodePublicKey([]byte{1})
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
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

func TestNodeRegistrationQuery_GetNodeRegistrationByAccountAddress(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetNodeRegistrationByAccountAddress("BCZ")
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE account_address = BCZ AND latest=1"
		wantArg := []interface{}{"BCZ"}
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
			mockNodeRegistry.NodeID, mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountAddress, mockNodeRegistry.RegistrationHeight,
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
			"id", "NodePublicKey", "AccountAddress", "RegistrationHeight", "NodeAddress", "LockedBalance",
			"Queued", "Latest", "Height"}).
			AddRow(mockNodeRegistry.NodeID, mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountAddress, mockNodeRegistry.RegistrationHeight,
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

func TestNodeRegistrationQuery_UpdateNodeRegistration(t *testing.T) {
	t.Run("UpdateNodeRegistration:success", func(t *testing.T) {

		q, args := mockNodeRegistrationQuery.UpdateNodeRegistration(mockNodeRegistry)
		wantQ0 := "UPDATE node_registry SET latest = 0 WHERE ID = 1"
		wantQ1 := "INSERT INTO node_registry (id,node_public_key,account_address,registration_height,node_address," +
			"locked_balance,queued,latest,height) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockNodeRegistry.NodeID, mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountAddress, mockNodeRegistry.RegistrationHeight,
			mockNodeRegistry.NodeAddress, mockNodeRegistry.LockedBalance, mockNodeRegistry.Queued,
			mockNodeRegistry.Latest, mockNodeRegistry.Height,
		}
		if q[0] != wantQ0 {
			t.Errorf("update query returned wrong: get: %s\nwant: %s", q, wantQ0)
		}
		if q[1] != wantQ1 {
			t.Errorf("insert query returned wrong: get: %s\nwant: %s", q, wantQ1)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByID(t *testing.T) {
	t.Run("GetNodeRegistrationByID:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetNodeRegistrationByID(1)
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance," +
			" queued, latest, height FROM node_registry WHERE id = ? AND latest=1"
		wantArg := []interface{}{int64(1)}
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if !reflect.DeepEqual(arg, wantArg) {
			t.Errorf("argument not match:\nget: %v\nwant: %v", arg, wantArg)
		}
	})
}

func TestNodeRegistrationQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantQueries [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountBalanceQuery),
			args:   args{height: uint32(1)},
			wantQueries: [][]interface{}{
				{
					"DELETE FROM account_balance WHERE height > ?",
					[]interface{}{uint32(1)},
				},
				{`
			UPDATE account_balance SET latest = ?
			WHERE height || '_' || id) IN (
				SELECT (MAX(height) || '_' || id) as con
				FROM account_balance
				WHERE latest = 0
				GROUP BY id
			)`,
					[]interface{}{1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nr := &NodeRegistrationQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQueries := nr.Rollback(tt.args.height)
			if !reflect.DeepEqual(gotQueries, tt.wantQueries) {
				t.Errorf("Rollback() gotQueries = \n%v, want \n%v", gotQueries, tt.wantQueries)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetNodeRegistrationsByHighestLockedBalance(t *testing.T) {
	t.Run("GetNodeRegistrationsByHighestLockedBalance", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(2)
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued," +
			" latest, height FROM node_registry WHERE locked_balance > 0 AND latest=1 ORDER BY locked_balance DESC LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

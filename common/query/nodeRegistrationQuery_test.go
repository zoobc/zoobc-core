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
		res := mockNodeRegistrationQuery.GetNodeRegistrationByNodePublicKey()
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByAccountAddress(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res, args := mockNodeRegistrationQuery.GetNodeRegistrationByAccountAddress("BCZ")
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE account_address = ? AND latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		wantArg := []interface{}{
			mockNodeRegistry.AccountAddress,
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
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

	t.Run("NodeRegistrationQuery-BuildModel-WithAggregation:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo-withAggregation").WillReturnRows(sqlmock.NewRows([]string{
			"id", "NodePublicKey", "AccountAddress", "RegistrationHeight", "NodeAddress", "LockedBalance",
			"Queued", "Latest", "Height", "max_height"}).
			AddRow(mockNodeRegistry.NodeID, mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountAddress, mockNodeRegistry.RegistrationHeight,
				mockNodeRegistry.NodeAddress, mockNodeRegistry.LockedBalance, mockNodeRegistry.Queued,
				mockNodeRegistry.Latest, mockNodeRegistry.Height, 123))
		rows, _ := db.Query("foo-withAggregation")
		var tempNode []*model.NodeRegistration
		res := mockNodeRegistrationQuery.BuildModel(tempNode, rows)
		if !reflect.DeepEqual(res[0], mockNodeRegistry) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockNodeRegistry)
		}
	})
}

func TestNodeRegistrationQuery_UpdateNodeRegistration(t *testing.T) {
	t.Run("UpdateNodeRegistration:success", func(t *testing.T) {

		q := mockNodeRegistrationQuery.UpdateNodeRegistration(mockNodeRegistry)
		wantQ0 := "UPDATE node_registry SET latest = 0 WHERE ID = ?"
		wantQ1 := "INSERT INTO node_registry (id,node_public_key,account_address,registration_height,node_address," +
			"locked_balance,queued,latest,height) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockNodeRegistry.NodeID, mockNodeRegistry.NodePublicKey, mockNodeRegistry.AccountAddress, mockNodeRegistry.RegistrationHeight,
			mockNodeRegistry.NodeAddress, mockNodeRegistry.LockedBalance, mockNodeRegistry.Queued,
			mockNodeRegistry.Latest, mockNodeRegistry.Height,
		}
		if q[0][0] != wantQ0 {
			t.Errorf("update query returned wrong: get: %s\nwant: %s", q, wantQ0)
		}
		if !reflect.DeepEqual(q[0][1], wantArg[0]) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", q[0][1], wantArg[0])
		}
		if q[1][0] != wantQ1 {
			t.Errorf("insert query returned wrong: get: %s\nwant: %s", q, wantQ1)
		}
		for idx, wanted := range wantArg {
			if !reflect.DeepEqual(q[1][idx+1], wanted) {
				t.Errorf("arguments returned wrong: get: %v\nwant: %v", q[1][idx+1], wanted)
			}
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
					uint32(1),
				},
				{`
			UPDATE account_balance SET latest = ?
			WHERE (height || '_' || id) IN (
				SELECT (MAX(height) || '_' || id) as con
				FROM account_balance
				WHERE latest = 0
				GROUP BY id
			)`,
					1,
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
		res := mockNodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(2, true)
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, " +
			"locked_balance, queued, latest, height FROM node_registry WHERE locked_balance > 0 " +
			"AND queued = 1 AND latest=1 ORDER BY locked_balance DESC LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationsWithZeroScore(t *testing.T) {
	t.Run("GetNodeRegistrationsWithZeroScore", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationsWithZeroScore(false)
		want := "SELECT A.id, A.node_public_key, A.account_address, A.registration_height, A.node_address, " +
			"A.locked_balance, A.queued, A.latest, A.height FROM node_registry as A " +
			"INNER JOIN participation_score as B ON A.id = B.node_id WHERE B.score = 0 AND A.latest=1 AND A.queued=0 AND B.latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetLastVersionedNodeRegistrationByPublicKey(t *testing.T) {
	t.Run("GetLastVersionedNodeRegistrationByPublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey([]byte{1}, uint32(1))
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"queued, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1"
		wantArg := []interface{}{[]byte{1}, uint32(1)}
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if !reflect.DeepEqual(arg, wantArg) {
			t.Errorf("argument not match:\nget: %v\nwant: %v", arg[0], wantArg[0])
		}
	})
}

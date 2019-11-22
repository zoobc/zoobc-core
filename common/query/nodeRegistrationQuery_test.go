package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockNodeRegistrationQuery = NewNodeRegistrationQuery()
	mockNodeAddress           = &model.NodeAddress{Address: "127.0.0.1", Port: 8000}
	mockNodeRegistry          = &model.NodeRegistration{
		NodeID:             1,
		NodePublicKey:      []byte{1},
		AccountAddress:     "BCZ",
		RegistrationHeight: 1,
		NodeAddress:        mockNodeAddress,
		LockedBalance:      10000,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
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

func TestNodeRegistrationQuery_GetNodeRegistrations(t *testing.T) {
	t.Run("GetNodeRegistrations", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrations(0, 2)
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE height >= 0 AND latest=1 LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByNodePublicKey(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationByNodePublicKey()
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByAccountAddress(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res, args := mockNodeRegistrationQuery.GetNodeRegistrationByAccountAddress("BCZ")
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC"
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
			mockNodeRegistry.NodeID,
			mockNodeRegistry.NodePublicKey,
			mockNodeRegistry.AccountAddress,
			mockNodeRegistry.RegistrationHeight,
			mockNodeRegistrationQuery.ExtractNodeAddress(mockNodeRegistry.GetNodeAddress()),
			mockNodeRegistry.LockedBalance,
			mockNodeRegistry.RegistrationStatus,
			mockNodeRegistry.Latest,
			mockNodeRegistry.Height,
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

		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(mockNodeRegistrationQuery.Fields).
			AddRow(
				mockNodeRegistry.NodeID,
				mockNodeRegistry.NodePublicKey,
				mockNodeRegistry.AccountAddress,
				mockNodeRegistry.RegistrationHeight,
				mockNodeRegistrationQuery.ExtractNodeAddress(mockNodeRegistry.GetNodeAddress()),
				mockNodeRegistry.LockedBalance,
				mockNodeRegistry.RegistrationStatus,
				mockNodeRegistry.Latest,
				mockNodeRegistry.Height,
			))
		rows, _ := db.Query("foo")
		var tempNode []*model.NodeRegistration
		res, _ := mockNodeRegistrationQuery.BuildModel(tempNode, rows)
		if !reflect.DeepEqual(res[0], mockNodeRegistry) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockNodeRegistry)
		}
	})

	t.Run("NodeRegistrationQuery-BuildModel-WithAggregation:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo-withAggregation").WillReturnRows(sqlmock.NewRows([]string{
			"id", "NodePublicKey", "AccountAddress", "RegistrationHeight", "NodeAddress", "LockedBalance",
			"RegistrationStatus", "Latest", "Height", "max_height"}).
			AddRow(
				mockNodeRegistry.NodeID,
				mockNodeRegistry.NodePublicKey,
				mockNodeRegistry.AccountAddress,
				mockNodeRegistry.RegistrationHeight,
				mockNodeRegistrationQuery.ExtractNodeAddress(mockNodeRegistry.GetNodeAddress()),
				mockNodeRegistry.LockedBalance,
				mockNodeRegistry.RegistrationStatus,
				mockNodeRegistry.Latest,
				mockNodeRegistry.Height,
				123,
			))
		rows, _ := db.Query("foo-withAggregation")
		var tempNode []*model.NodeRegistration
		res, _ := mockNodeRegistrationQuery.BuildModel(tempNode, rows)
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
			"locked_balance,registration_status,latest,height) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockNodeRegistry.NodeID,
			mockNodeRegistry.NodePublicKey,
			mockNodeRegistry.AccountAddress,
			mockNodeRegistry.RegistrationHeight,
			mockNodeRegistrationQuery.ExtractNodeAddress(mockNodeRegistry.GetNodeAddress()),
			mockNodeRegistry.LockedBalance,
			mockNodeRegistry.RegistrationStatus,
			mockNodeRegistry.Latest,
			mockNodeRegistry.Height,
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
			" registration_status, latest, height FROM node_registry WHERE id = ? AND latest=1"
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
			WHERE latest = ? AND (height || '_' || id) IN (
				SELECT (MAX(height) || '_' || id) as con
				FROM account_balance
				GROUP BY id
			)`,
					1, 0,
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
		res := mockNodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(2, model.NodeRegistrationState_NodeQueued)
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, " +
			"locked_balance, registration_status, latest, height FROM node_registry WHERE locked_balance > 0 " +
			"AND registration_status = 1 AND latest=1 ORDER BY locked_balance DESC LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationsWithZeroScore(t *testing.T) {
	t.Run("GetNodeRegistrationsWithZeroScore", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationsWithZeroScore(model.NodeRegistrationState_NodeRegistered)
		want := "SELECT A.id, A.node_public_key, A.account_address, A.registration_height, A.node_address, A.locked_balance, " +
			"A.registration_status, A.latest, A.height FROM node_registry as A INNER JOIN participation_score as B ON A.id = B.node_id " +
			"WHERE B.score <= 0 AND A.latest=1 AND A.registration_status=0 AND B.latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetLastVersionedNodeRegistrationByPublicKey(t *testing.T) {
	t.Run("GetLastVersionedNodeRegistrationByPublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey([]byte{1}, uint32(1))
		want := "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1"
		wantArg := []interface{}{[]byte{1}, uint32(1)}
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if !reflect.DeepEqual(arg, wantArg) {
			t.Errorf("argument not match:\nget: %v\nwant: %v", arg[0], wantArg[0])
		}
	})
}

type (
	mockQueryExecutorScan struct {
		Executor
	}
)

func (*mockQueryExecutorScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockNodeRegistrationQuery.Fields).AddRow(
			1,
			[]byte{1},
			"BCZ",
			1,
			"127.0.0.1:8000",
			10000,
			uint32(model.NodeRegistrationState_NodeQueued),
			true,
			0,
		),
	)
	return db.QueryRow("")
}
func TestNodeRegistrationQuery_Scan(t *testing.T) {

	var nodeRegistration model.NodeRegistration

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nr  *model.NodeRegistration
		row *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.NodeRegistration
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockNodeRegistrationQuery),
			args: args{
				nr:  &nodeRegistration,
				row: (&mockQueryExecutorScan{}).ExecuteSelectRow(""),
			},
			want: model.NodeRegistration{
				NodeID:             1,
				NodePublicKey:      []byte{1},
				AccountAddress:     "BCZ",
				RegistrationHeight: 1,
				NodeAddress:        mockNodeAddress,
				LockedBalance:      10000,
				RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
				Latest:             true,
				Height:             0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := &NodeRegistrationQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := nrq.Scan(tt.args.nr, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.want, nodeRegistration) {
				t.Errorf("Scan() want = \n%v, but got \n%v", tt.want, nodeRegistration)
			}
		})
	}
}
func TestNodeRegistrationQuery_GetActiveNodeRegistrations(t *testing.T) {
	t.Run("GetActiveNodeRegistrations", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetActiveNodeRegistrations()
		want := "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score " +
			"FROM node_registry AS nr INNER JOIN participation_score AS ps ON nr.id = ps.node_id " +
			"WHERE nr.registration_status = 0 AND nr.latest = 1 " +
			"AND ps.score > 0 AND ps.latest = 1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_ClearDeletedNodeRegistration(t *testing.T) {
	t.Run("ClearDeletedNodeRegistration", func(t *testing.T) {
		nr := &model.NodeRegistration{
			NodeID: 1,
		}
		res := mockNodeRegistrationQuery.ClearDeletedNodeRegistration(nr)
		want := "UPDATE node_registry SET latest = 0 WHERE ID = ? AND registration_status = 2"
		qry := res[0][0]
		args := res[0][1]
		if qry != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if args.(int64) != 1 {
			t.Errorf("args don't match:\nget: %d\nwant: %d", args, 1)
		}
	})
}

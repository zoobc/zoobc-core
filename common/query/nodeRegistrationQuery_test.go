package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockNodeRegistrationQuery = NewNodeRegistrationQuery()
	mockNodeRegistry          = &model.NodeRegistration{
		NodeID:        1,
		NodePublicKey: []byte{1},
		AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
			202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		RegistrationHeight: 1,
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
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE height >= 0 AND latest=1 LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByNodePublicKey(t *testing.T) {
	t.Run("GetNodeRegistrationByNodePublicKey:success", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationByNodePublicKey()
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationByAccountAddress(t *testing.T) {
	t.Run("GetNodeRegistrationByAccountAddress:success", func(t *testing.T) {
		res, args := mockNodeRegistrationQuery.GetNodeRegistrationByAccountAddress([]byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62,
			98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14})
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1"
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
			"id", "NodePublicKey", "AccountAddress", "RegistrationHeight", "LockedBalance",
			"RegistrationStatus", "Latest", "Height", "max_height"}).
			AddRow(
				mockNodeRegistry.NodeID,
				mockNodeRegistry.NodePublicKey,
				mockNodeRegistry.AccountAddress,
				mockNodeRegistry.RegistrationHeight,
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
		wantQ1 := fmt.Sprintf("INSERT INTO node_registry (id,node_public_key,account_address,registration_height,"+
			"locked_balance,registration_status,latest,height) VALUES(? , ?, ?, ?, ?, ?, ?, ?) "+
			"ON CONFLICT(id, height) DO UPDATE SET registration_status = %d, latest = 1", mockNodeRegistry.RegistrationStatus)
		wantArg := []interface{}{
			mockNodeRegistry.NodeID,
			mockNodeRegistry.NodePublicKey,
			mockNodeRegistry.AccountAddress,
			mockNodeRegistry.RegistrationHeight,
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
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance," +
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
			WHERE latest = ? AND (id, height) IN (
				SELECT t2.id, MAX(t2.height)
				FROM account_balance as t2
				GROUP BY t2.id
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
		want := "SELECT id, node_public_key, account_address, registration_height, " +
			"locked_balance, registration_status, latest, height FROM node_registry WHERE locked_balance > 0 " +
			"AND registration_status = 1 AND latest=1 ORDER BY locked_balance DESC LIMIT 2"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetLastVersionedNodeRegistrationByPublicKey(t *testing.T) {
	t.Run("GetLastVersionedNodeRegistrationByPublicKey:success", func(t *testing.T) {
		res, arg := mockNodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey([]byte{1}, uint32(1))
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
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
			mockNodeRegistry.AccountAddress,
			1,
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
		Fields                  []string
		JoinedAddressInfoFields []string
		TableName               string
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
				AccountAddress:     mockNodeRegistry.AccountAddress,
				RegistrationHeight: 1,
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

func TestNodeRegistrationQuery_GetActiveNodeRegistrationsByHeight(t *testing.T) {
	t.Run("GetActiveNodeRegistrations", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetActiveNodeRegistrationsByHeight(1)
		want := "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score, " +
			"max(nr.height) AS max_height FROM node_registry AS nr INNER JOIN participation_score AS ps ON nr.id = " +
			"ps.node_id WHERE nr.height <= 1 AND nr.registration_status = 0 AND nr.latest = 1 AND ps.score > 0 AND " +
			"ps.latest = 1 GROUP BY nr.id"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_GetNodeRegistrationsByBlockTimestampInterval(t *testing.T) {
	t.Run("GetActiveNodeRegistrations", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationsByBlockTimestampInterval(0, 1)
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE height >= (SELECT MIN(height) " +
			"FROM main_block AS mb1 WHERE mb1.timestamp >= 0) AND height <= (SELECT MAX(height) " +
			"FROM main_block AS mb2 WHERE mb2.timestamp < 1) AND registration_status != 1 AND latest=1 ORDER BY height"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_InsertNodeRegistration(t *testing.T) {
	t.Run("GetActiveNodeRegistrations", func(t *testing.T) {
		qry, _ := mockNodeRegistrationQuery.InsertNodeRegistration(&model.NodeRegistration{})
		want := "INSERT INTO node_registry (id, node_public_key, account_address, registration_height, " +
			"locked_balance, registration_status, latest, height) VALUES(? , ? , ? , ? , ? , ? , ? , ? )"
		if qry != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", qry, want)
		}
	})
}

func TestNodeRegistrationQuery_TrimDataBeforeSnapshot(t *testing.T) {
	t.Run("TrimDataBeforeSnapshot:success", func(t *testing.T) {
		res := mockNodeRegistrationQuery.TrimDataBeforeSnapshot(0, 10)
		want := "DELETE FROM node_registry WHERE height >= 0 AND height <= 10 AND height != 0"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestNodeRegistrationQuery_SelectDataForSnapshot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "SelectDataForSnapshot:success-{fromGenesis}",
			fields: fields{
				TableName: NewNodeRegistrationQuery().TableName,
				Fields:    NewNodeRegistrationQuery().Fields,
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "SELECT id,node_public_key,account_address,registration_height,locked_balance,registration_status,latest,height " +
				"FROM node_registry WHERE height >= 0 AND height <= 10 AND height != 0 ORDER BY height, id",
		},
		{
			name: "SelectDataForSnapshot:success-{fromArbitraryHeight}",
			fields: fields{
				TableName: NewNodeRegistrationQuery().TableName,
				Fields:    NewNodeRegistrationQuery().Fields,
			},
			args: args{
				fromHeight: 720,
				toHeight:   1440,
			},
			want: "SELECT id,node_public_key,account_address,registration_height,locked_balance,registration_status,latest,height " +
				"FROM node_registry WHERE (id, height) IN (SELECT t2.id, MAX(t2.height) " +
				"FROM node_registry as t2 WHERE t2.height > 0 AND t2.height < 720 GROUP BY t2.id) " +
				"UNION ALL SELECT id,node_public_key,account_address,registration_height,locked_balance,registration_status,latest,height " +
				"FROM node_registry WHERE height >= 720 AND height <= 1440 ORDER BY height, id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := &NodeRegistrationQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := nrq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("NodeRegistrationQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetNodeRegistryAtHeight(t *testing.T) {
	type args struct {
		height uint32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "GetNodeRegistryAtHeightQuery",
			args: args{
				height: 11120,
			},
			want: "SELECT id, node_public_key, account_address, registration_height, " +
				"locked_balance, registration_status, latest, height FROM node_registry " +
				"where registration_status = 0 AND (id,height) in " +
				"(SELECT id,MAX(height) FROM node_registry WHERE height <= 11120 GROUP BY id) " +
				"ORDER BY height DESC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := NewNodeRegistrationQuery()
			if got := nrq.GetNodeRegistryAtHeight(tt.args.height); got != tt.want {
				t.Errorf("NodeRegistrationQuery.GetNodeRegistryAtHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetNodeRegistryAtHeightWithNodeAddress(t *testing.T) {
	type fields struct {
		Fields                  []string
		JoinedAddressInfoFields []string
		TableName               string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetNodeRegistryAtHeightWithNodeAddress:success",
			fields: fields{
				TableName:               NewNodeRegistrationQuery().TableName,
				Fields:                  NewNodeRegistrationQuery().Fields,
				JoinedAddressInfoFields: NewNodeRegistrationQuery().JoinedAddressInfoFields,
			},
			args: args{
				height: 10,
			},
			want: "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, latest, height, " +
				"t2.address AS node_address, t2.port AS node_address_port, t2.status AS node_address_status " +
				"FROM node_registry INNER JOIN node_address_info AS t2 ON id = t2.node_id WHERE registration_status = 0 " +
				"AND (id,height) in (SELECT t1.id,MAX(t1.height) FROM node_registry AS t1 WHERE t1.height <= 10 " +
				"GROUP BY t1.id) ORDER BY id, t2.status",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := &NodeRegistrationQuery{
				Fields:                  tt.fields.Fields,
				TableName:               tt.fields.TableName,
				JoinedAddressInfoFields: tt.fields.JoinedAddressInfoFields,
			}
			if got := nrq.GetNodeRegistryAtHeightWithNodeAddress(tt.args.height); got != tt.want {
				t.Errorf("NodeRegistrationQuery.GetNodeRegistryAtHeightWithNodeAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetActiveNodeRegistrationsWithNodeAddress(t *testing.T) {
	type fields struct {
		Fields                  []string
		JoinedAddressInfoFields []string
		TableName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "GetActiveNodeRegistrationsWithNodeAddress:success",
			fields: fields{
				TableName:               NewNodeRegistrationQuery().TableName,
				Fields:                  NewNodeRegistrationQuery().Fields,
				JoinedAddressInfoFields: NewNodeRegistrationQuery().JoinedAddressInfoFields,
			},
			want: "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, latest, height, " +
				"t2.address AS node_address, t2.port AS node_address_port, t2.status AS node_address_status FROM node_registry " +
				"INNER JOIN node_address_info AS t2 ON id = t2.node_id WHERE registration_status = 0 ORDER BY height DESC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := &NodeRegistrationQuery{
				Fields:                  tt.fields.Fields,
				TableName:               tt.fields.TableName,
				JoinedAddressInfoFields: tt.fields.JoinedAddressInfoFields,
			}
			if got := nrq.GetActiveNodeRegistrationsWithNodeAddress(); got != tt.want {
				t.Errorf("NodeRegistrationQuery.GetActiveNodeRegistrationsWithNodeAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetLastVersionedNodeRegistrationByPublicKeyWithNodeAddress(t *testing.T) {
	type fields struct {
		Fields                  []string
		JoinedAddressInfoFields []string
		TableName               string
	}
	type args struct {
		nodePublicKey []byte
		height        uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetActiveNodeRegistrationsWithNodeAddress:success",
			fields: fields{
				TableName:               NewNodeRegistrationQuery().TableName,
				JoinedAddressInfoFields: NewNodeRegistrationQuery().JoinedAddressInfoFields,
				Fields:                  NewNodeRegistrationQuery().Fields,
			},
			args: args{
				height:        10,
				nodePublicKey: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
			wantStr: "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, latest, height, " +
				"t2.address AS node_address, t2.port AS node_address_port, t2.status AS node_address_status " +
				"FROM node_registry LEFT JOIN node_address_info AS t2 ON id = t2.node_id " +
				"WHERE (node_public_key = ? OR t2.node_id IS NULL) AND height <= ? ORDER BY height DESC LIMIT 1",
			wantArgs: []interface{}{[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1}, uint32(10)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := &NodeRegistrationQuery{
				Fields:                  tt.fields.Fields,
				JoinedAddressInfoFields: tt.fields.JoinedAddressInfoFields,
				TableName:               tt.fields.TableName,
			}
			gotStr, gotArgs := nrq.GetLastVersionedNodeRegistrationByPublicKeyWithNodeAddress(tt.args.nodePublicKey, tt.args.height)
			if gotStr != tt.wantStr {
				t.Errorf("GetLastVersionedNodeRegistrationByPublicKeyWithNodeAddress() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetLastVersionedNodeRegistrationByPublicKeyWithNodeAddress() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetPendingNodeRegistrations(t *testing.T) {
	type args struct {
		limit uint32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "wantSuccess",
			args: args{
				limit: 1,
			},
			want: "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
				"registration_status, latest, height FROM node_registry WHERE registration_status=1 AND latest=1 ORDER BY locked_balance DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrq := NewNodeRegistrationQuery()
			if got := nrq.GetPendingNodeRegistrations(tt.args.limit); got != tt.want {
				t.Errorf("NodeRegistrationQuery.GetPendingNodeRegistrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistrationQuery_GetNodeRegistrationsByBlockHeightInterval(t *testing.T) {
	t.Run("GetNodeRegistrationsByBlockHeightInterval", func(t *testing.T) {
		res := mockNodeRegistrationQuery.GetNodeRegistrationsByBlockHeightInterval(0, 1)
		want := "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
			"registration_status, latest, height FROM node_registry WHERE height >= 0 AND height <= 1 AND " +
			"registration_status != 1 AND latest=1 ORDER BY height"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

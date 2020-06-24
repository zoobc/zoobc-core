package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestNodeAddressInfoQuery_InsertNodeAddressInfo(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		peerAddress *model.NodeAddressInfo
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "InsertNodeAddressInfo:success",
			args: args{
				peerAddress: &model.NodeAddressInfo{
					NodeID:      111,
					Address:     "192.168.1.2",
					Port:        8080,
					BlockHeight: 100,
					BlockHash:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					Signature: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
						1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
						1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
			},
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			wantArgs: []interface{}{
				int64(111),
				"192.168.1.2",
				uint32(8080),
				uint32(100),
				[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
					1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
			wantStr: "INSERT OR REPLACE INTO node_address_info (node_id, address, port, block_height, block_hash, signature) " +
				"VALUES(? , ? , ? , ? , ? , ? )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := paq.InsertNodeAddressInfo(tt.args.peerAddress)
			if gotStr != tt.wantStr {
				t.Errorf("NodeAddressInfoQuery.InsertNodeAddressInfo() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("NodeAddressInfoQuery.InsertNodeAddressInfo() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestNodeAddressInfoQuery_UpdateNodeAddressInfo(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		peerAddress *model.NodeAddressInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name: "UpdateNodeAddressInfo:success",
			args: args{
				peerAddress: &model.NodeAddressInfo{
					NodeID:      111,
					Address:     "192.168.1.2",
					Port:        8080,
					BlockHeight: 100,
					BlockHash:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					Signature: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
						1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
			},
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			want: [][]interface{}{
				append([]interface{}{"UPDATE node_address_info SET address = ?, " +
					"port = ?, block_height = ?, block_hash = ?, signature = ? WHERE" +
					" node_id = ?"},
					"192.168.1.2",
					uint32(8080),
					uint32(100),
					[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
						1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					int64(111),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := paq.UpdateNodeAddressInfo(tt.args.peerAddress); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAddressInfoQuery.UpdateNodeAddressInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoQuery_DeleteNodeAddressInfoByNodeID(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nodeID int64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "DeleteNodeAddressInfo:success",
			args: args{
				nodeID: 111,
			},
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			wantArgs: []interface{}{
				int64(111),
			},
			wantStr: "DELETE FROM node_address_info WHERE node_id = ?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := paq.DeleteNodeAddressInfoByNodeID(tt.args.nodeID)
			if gotStr != tt.wantStr {
				t.Errorf("NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestNodeAddressInfoQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pa *model.NodeAddressInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "ExtractModel:success",
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			args: args{
				pa: &model.NodeAddressInfo{
					NodeID:      111,
					Address:     "192.168.1.2",
					Port:        8080,
					BlockHeight: 100,
					BlockHash:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					Signature: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
						1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
			},
			want: []interface{}{
				int64(111),
				"192.168.1.2",
				uint32(8080),
				uint32(100),
				[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
					1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := paq.ExtractModel(tt.args.pa); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAddressInfoQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorNodeAddressInfoBuildModel struct {
		Executor
	}
)

func (*mockQueryExecutorNodeAddressInfoBuildModel) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(NewNodeAddressInfoQuery().Fields).AddRow(
		int64(111),
		"192.168.1.2",
		uint32(8080),
		uint32(100),
		[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	))
	return db.Query("")
}

func (*mockQueryExecutorNodeAddressInfoBuildModel) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(NewNodeAddressInfoQuery().Fields).AddRow(
		int64(111),
		"192.168.1.2",
		uint32(8080),
		uint32(100),
		[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	))
	return db.QueryRow("")
}

func TestNodeAddressInfoQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pas  []*model.NodeAddressInfo
		rows *sql.Rows
	}
	rows, err := (&mockQueryExecutorNodeAddressInfoBuildModel{}).ExecuteSelect("", false, "")
	if err != nil {
		t.Errorf("Rows Failed: %s", err.Error())
		return
	}
	defer rows.Close()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.NodeAddressInfo
	}{
		{
			name:   "BuildModel:success",
			fields: fields(*NewNodeAddressInfoQuery()),
			args: args{
				pas:  []*model.NodeAddressInfo{},
				rows: rows,
			},
			want: []*model.NodeAddressInfo{
				{
					NodeID:      111,
					Address:     "192.168.1.2",
					Port:        8080,
					BlockHeight: 100,
					BlockHash:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					Signature: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
						1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got, _ := re.BuildModel(tt.args.pas, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pa  *model.NodeAddressInfo
		row *sql.Row
	}
	row := (&mockQueryExecutorNodeAddressInfoBuildModel{}).ExecuteSelectRow("", false, "")

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Scan:success",
			fields: fields(*NewNodeAddressInfoQuery()),
			args: args{
				pa:  &model.NodeAddressInfo{},
				row: row,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := paq.Scan(tt.args.pa, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("NodeAddressInfoQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoQuery_GetNodeAddressInfoByNodeIDs(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nodeIDs []int64
		status  model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantStr string
	}{
		{
			name: "GetNodeAddressInfoByNodeIDs:success-{statusPending}",
			args: args{
				nodeIDs: []int64{1, 2, 3, 4, 5, 6, 7, 100, 2, 3, 4, 6, 7},
				status:  model.NodeAddressStatus_NodeAddressPending,
			},
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			wantStr: "SELECT node_id, address, port, block_height, block_hash, signature FROM node_address_info WHERE node_id IN " +
				"(1, 2, 3, 4, 5, 6, 7, 100, 2, 3, 4, 6, 7) AND status = 0",
		},
		{
			name: "GetNodeAddressInfoByNodeIDs:success-{statusConfirmed}",
			args: args{
				nodeIDs: []int64{1, 2, 3, 4, 5, 6, 7, 100, 2, 3, 4, 6, 7},
				status:  model.NodeAddressStatus_NodeAddressConfirmed,
			},
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			wantStr: "SELECT node_id, address, port, block_height, block_hash, signature FROM node_address_info WHERE node_id IN " +
				"(1, 2, 3, 4, 5, 6, 7, 100, 2, 3, 4, 6, 7) AND status = 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr := paq.GetNodeAddressInfoByNodeIDs(tt.args.nodeIDs, tt.args.status)
			if gotStr != tt.wantStr {
				t.Errorf("NodeAddressInfoQuery.GetNodeAddressInfoByNodeIDs() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestNodeAddressInfoQuery_GetNodeAddressInfoByAddressPort(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		address string
		port    uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetNodeAddressInfoByAddressPort:success",
			args: args{
				port:    8001,
				address: "127.0.0.1",
			},
			fields: fields{
				Fields:    NewNodeAddressInfoQuery().Fields,
				TableName: NewNodeAddressInfoQuery().TableName,
			},
			wantArgs: []interface{}{
				"127.0.0.1",
				uint32(8001),
			},
			wantStr: "SELECT node_id, address, port, block_height, block_hash, signature FROM node_address_info WHERE node_id = ? " +
				"AND port = ? LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paq := &NodeAddressInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := paq.GetNodeAddressInfoByAddressPort(tt.args.address, tt.args.port)
			if gotStr != tt.wantStr {
				t.Errorf("NodeAddressInfoQuery.GetNodeAddressInfoByAddressPort() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("NodeAddressInfoQuery.GetNodeAddressInfoByAddressPort() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

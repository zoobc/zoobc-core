package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockBlockchainObjectQuery = NewBlockchainObjectQuery()
	mockBlockchainObject      = &model.BlockchainObject{
		ID: []byte{0, 0, 0, 0, 7, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		OwnerAccountAddress: []byte{1, 2},
		BlockHeight:         12,
		Latest:              true,
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
			wantStr: "INSERT INTO blockchain_object (id, owner_account_address, block_height, latest) VALUES (? , ?, ?, ?)",
			wantArgs: []interface{}{
				mockBlockchainObject.ID,
				mockBlockchainObject.OwnerAccountAddress,
				mockBlockchainObject.BlockHeight,
				mockBlockchainObject.Latest,
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

func TestBlockchainObjectQuery_InsertBlockcahinObjects(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockchainObjects []*model.BlockchainObject
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
				blockchainObjects: []*model.BlockchainObject{
					mockBlockchainObject,
					mockBlockchainObject,
				},
			},
			wantStr: "INSERT INTO blockchain_object (id, owner_account_address, block_height, latest) " +
				"VALUES (?,? ,? ,? ), (?,? ,? ,? )",
			wantArgs: []interface{}{
				mockBlockchainObject.ID,
				mockBlockchainObject.OwnerAccountAddress,
				mockBlockchainObject.BlockHeight,
				mockBlockchainObject.Latest,
				mockBlockchainObject.ID,
				mockBlockchainObject.OwnerAccountAddress,
				mockBlockchainObject.BlockHeight,
				mockBlockchainObject.Latest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := boq.InsertBlockcahinObjects(tt.args.blockchainObjects)
			if gotStr != tt.wantStr {
				t.Errorf("BlockchainObjectQuery.InsertBlockcahinObjects() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BlockchainObjectQuery.InsertBlockcahinObjects() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBlockchainObjectQuery_GetBlockchainObjectByID(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		id []byte
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
				id: mockBlockchainObject.ID,
			},
			wantStr: "SELECT id,owner_account_address,block_height,latest FROM blockchain_object " +
				"WHERE id = ? AND latest = 1 ORDER BY block_height DESC",
			wantArgs: []interface{}{
				mockBlockchainObject.ID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := boq.GetBlockchainObjectByID(tt.args.id)
			if gotStr != tt.wantStr {
				t.Errorf("BlockchainObjectQuery.GetBlockchainObjectByID() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BlockchainObjectQuery.GetBlockchainObjectByID() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBlockchainObjectQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockchainObject *model.BlockchainObject
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := b.ExtractModel(tt.args.blockchainObject); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainObjectQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectQuery_BuildModel(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rowsMock := sqlmock.NewRows(mockBlockchainObjectQuery.Fields)
	rowsMock.AddRow(
		mockBlockchainObject.GetID(),
		mockBlockchainObject.GetOwnerAccountAddress(),
		mockBlockchainObject.GetBlockHeight(),
		mockBlockchainObject.GetLatest(),
	)
	mock.ExpectQuery("").WillReturnRows(rowsMock)
	rows, _ := db.Query("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockchainObjects []*model.BlockchainObject
		rows              *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.BlockchainObject
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				blockchainObjects: []*model.BlockchainObject{},
				rows:              rows,
			},
			want: []*model.BlockchainObject{
				mockBlockchainObject,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := b.BuildModel(tt.args.blockchainObjects, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockchainObjectQuery.BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainObjectQuery.BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectQuery_Scan(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rowsMock := sqlmock.NewRows(mockBlockchainObjectQuery.Fields)
	rowsMock.AddRow(
		mockBlockchainObject.GetID(),
		mockBlockchainObject.GetOwnerAccountAddress(),
		mockBlockchainObject.GetBlockHeight(),
		mockBlockchainObject.GetLatest(),
	)
	mock.ExpectQuery("").WillReturnRows(rowsMock)
	row := db.QueryRow("")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockchainObject *model.BlockchainObject
		row              *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectQuery),
			args: args{
				blockchainObject: &model.BlockchainObject{},
				row:              row,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := b.Scan(tt.args.blockchainObject, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("BlockchainObjectQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchainObjectQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectQuery),
			args: args{
				height: 1,
			},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM blockchain_object WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := boq.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("BlockchainObjectQuery.Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

func TestBlockchainObjectQuery_RecalibrateVersionedTable(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectQuery),
			want: []string{
				"UPDATE blockchain_object SET latest = false " +
					"WHERE latest = true AND id NOT IN (SELECT MAX(t2.block_height) FROM blockchain_object t2",
				"UPDATE blockchain_object SET latest = true " +
					"WHERE latest = false AND id IN (SELECT MAX(t2.block_height) FROM blockchain_object t2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := boq.RecalibrateVersionedTable(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainObjectQuery.RecalibrateVersionedTable() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectQuery_SelectDataForSnapshot(t *testing.T) {
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
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectQuery),
			args: args{
				fromHeight: 1,
				toHeight:   2,
			},
			want: "SELECT id,owner_account_address,block_height,latest FROM blockchain_object " +
				"WHERE block_height >= 1 AND block_height <= 2 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := boq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("BlockchainObjectQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectQuery),
			args: args{
				fromHeight: 1,
				toHeight:   2,
			},
			want: "DELETE FROM blockchain_object " +
				"WHERE block_height >= 1 AND block_height <= 2 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boq := &BlockchainObjectQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := boq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("BlockchainObjectQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockBlockchainObjectPropertyQuery = NewBlockchainObjectPropertyQuery()
	mockBlockchainObjectProperty      = model.BlockchainObjectProperty{
		BlockchainObjectID: []byte{0, 0, 0, 0, 7, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		Key:         "mockKey",
		Value:       "mockvalue",
		BlockHeight: 2,
	}
)

func TestBlockchainObjectPropertyQuery_InsertBlockcahinObjectProperties(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		properties []*model.BlockchainObjectProperty
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
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				properties: []*model.BlockchainObjectProperty{
					&mockBlockchainObjectProperty,
					&mockBlockchainObjectProperty,
				},
			},
			wantStr: "INSERT INTO blockchain_object_property (blockchain_object_id, key, value, block_height) " +
				"VALUES (?,? ,? ,? ), (?,? ,? ,? )",
			wantArgs: []interface{}{
				mockBlockchainObjectProperty.BlockchainObjectID,
				mockBlockchainObjectProperty.Key,
				mockBlockchainObjectProperty.Value,
				mockBlockchainObjectProperty.BlockHeight,
				mockBlockchainObjectProperty.BlockchainObjectID,
				mockBlockchainObjectProperty.Key,
				mockBlockchainObjectProperty.Value,
				mockBlockchainObjectProperty.BlockHeight,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bopq := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := bopq.InsertBlockcahinObjectProperties(tt.args.properties)
			if gotStr != tt.wantStr {
				t.Errorf("BlockchainObjectPropertyQuery.InsertBlockcahinObjectProperties() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BlockchainObjectPropertyQuery.InsertBlockcahinObjectProperties() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBlockchainObjectPropertyQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		property *model.BlockchainObjectProperty
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				property: &mockBlockchainObjectProperty,
			},
			want: []interface{}{
				mockBlockchainObjectProperty.GetBlockchainObjectID(),
				mockBlockchainObjectProperty.GetKey(),
				mockBlockchainObjectProperty.GetValue(),
				mockBlockchainObjectProperty.GetBlockHeight(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := b.ExtractModel(tt.args.property); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainObjectPropertyQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectPropertyQuery_BuildModel(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rowsMock := sqlmock.NewRows(mockBlockchainObjectPropertyQuery.Fields)
	rowsMock.AddRow(
		mockBlockchainObjectProperty.GetBlockchainObjectID(),
		mockBlockchainObjectProperty.GetKey(),
		mockBlockchainObjectProperty.GetValue(),
		mockBlockchainObjectProperty.GetBlockHeight(),
	)
	mock.ExpectQuery("").WillReturnRows(rowsMock)
	rows, _ := db.Query("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		properties []*model.BlockchainObjectProperty
		rows       *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.BlockchainObjectProperty
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				properties: []*model.BlockchainObjectProperty{},
				rows:       rows,
			},
			want: []*model.BlockchainObjectProperty{
				&mockBlockchainObjectProperty,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := b.BuildModel(tt.args.properties, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockchainObjectPropertyQuery.BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainObjectPropertyQuery.BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectPropertyQuery_Rollback(t *testing.T) {
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
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				height: 1,
			},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM blockchain_object_property WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bopq := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := bopq.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("BlockchainObjectPropertyQuery.Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

func TestBlockchainObjectPropertyQuery_RecalibrateVersionedTable(t *testing.T) {
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
			fields: fields(*mockBlockchainObjectPropertyQuery),
			want:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bopq := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := bopq.RecalibrateVersionedTable(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockchainObjectPropertyQuery.RecalibrateVersionedTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectPropertyQuery_SelectDataForSnapshot(t *testing.T) {
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
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				fromHeight: 1,
				toHeight:   2,
			},
			want: "SELECT blockchain_object_id,key,value,block_height FROM blockchain_object_property " +
				"WHERE block_height >= 1 AND block_height <= 2 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bopq := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := bopq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("BlockchainObjectPropertyQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainObjectPropertyQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			fields: fields(*mockBlockchainObjectPropertyQuery),
			args: args{
				fromHeight: 1,
				toHeight:   2,
			},
			want: "DELETE FROM blockchain_object_property WHERE block_height >= 1 AND block_height <= 2 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bopq := &BlockchainObjectPropertyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := bopq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("BlockchainObjectPropertyQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

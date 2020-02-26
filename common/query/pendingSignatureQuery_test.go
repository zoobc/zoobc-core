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
	mockPendingSignatureQueryIntance = NewPendingSignatureQuery()
)

func TestNewPendingSignatureQuery(t *testing.T) {
	tests := []struct {
		name string
		want *PendingSignatureQuery
	}{
		{
			name: "NewPendingSignatureQuery-Success",
			want: &PendingSignatureQuery{
				Fields: []string{
					"transaction_hash",
					"account_address",
					"signature",
					"block_height",
				},
				TableName: "pending_signature",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPendingSignatureQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPendingSignatureQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mock build model rows getter
func getPendingSignatureQueryBuildModelRowsFail() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomFields"})
	mockRow.AddRow(
		make([]byte, 32),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}
func getPendingSignatureQueryBuildModelRowsSuccess() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockPendingSignatureQueryIntance.Fields)
	mockRow.AddRow(
		make([]byte, 32),
		"account_address",
		make([]byte, 64),
		uint32(10),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

// mock build model rows getter

func TestPendingSignatureQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pss  []*model.PendingSignature
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PendingSignature
		wantErr bool
	}{
		{
			name: "BuildModel-Fail",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				pss:  []*model.PendingSignature{},
				rows: getPendingSignatureQueryBuildModelRowsFail(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "BuildModel-Success",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				pss:  []*model.PendingSignature{},
				rows: getPendingSignatureQueryBuildModelRowsSuccess(),
			},
			want: []*model.PendingSignature{
				{
					TransactionHash: make([]byte, 32),
					AccountAddress:  "account_address",
					Signature:       make([]byte, 64),
					BlockHeight:     10,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := psq.BuildModel(tt.args.pss, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockExtractModelPendingSig = &model.PendingSignature{
		TransactionHash: make([]byte, 32),
		AccountAddress:  "A",
		Signature:       make([]byte, 64),
		BlockHeight:     10,
	}
)

func TestPendingSignatureQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingSig *model.PendingSignature
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "ExtractModel-Success",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				pendingSig: mockExtractModelPendingSig,
			},
			want: []interface{}{
				&mockExtractModelPendingSig.TransactionHash,
				&mockExtractModelPendingSig.AccountAddress,
				&mockExtractModelPendingSig.Signature,
				&mockExtractModelPendingSig.BlockHeight,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := pe.ExtractModel(tt.args.pendingSig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingSignatureQuery_GetPendingSignatureByHash(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		txHash []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetPendingSignatureByHash-Success",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				txHash: make([]byte, 32),
			},
			wantStr: "SELECT transaction_hash, account_address, signature, block_height FROM pending_signature " +
				"WHERE transaction_hash = ?",
			wantArgs: []interface{}{
				make([]byte, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := psq.GetPendingSignatureByHash(tt.args.txHash)
			if gotStr != tt.wantStr {
				t.Errorf("GetPendingSignatureByHash() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetPendingSignatureByHash() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

var (
	mockInsertPendingSignaturePendingSig = &model.PendingSignature{
		TransactionHash: nil,
		AccountAddress:  "",
		Signature:       nil,
		BlockHeight:     0,
	}
)

func TestPendingSignatureQuery_InsertPendingSignature(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingSig *model.PendingSignature
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "InsertPendingSignature-Success",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				pendingSig: mockInsertPendingSignaturePendingSig,
			},
			wantStr: "INSERT INTO pending_signature (transaction_hash, account_address, signature, " +
				"block_height) VALUES(? , ? , ? , ? )",
			wantArgs: mockPendingSignatureQueryIntance.ExtractModel(mockInsertPendingSignaturePendingSig),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := psq.InsertPendingSignature(tt.args.pendingSig)
			if gotStr != tt.wantStr {
				t.Errorf("InsertPendingSignature() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertPendingSignature() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestPendingSignatureQuery_Rollback(t *testing.T) {
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
			name: "Rollback",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				height: 10,
			},
			wantMultiQueries: [][]interface{}{
				{
					fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", mockPendingSignatureQueryIntance.TableName),
					uint32(10),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := psq.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

// Scan mocks
func getMockScanRowFail() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")

}
func getMockScanRowSuccess() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockPendingSignatureQueryIntance.Fields)
	mockRow.AddRow(
		make([]byte, 32),
		"account_address",
		make([]byte, 64),
		uint32(10),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")

}

// Scan mocks

func TestPendingSignatureQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingSig *model.PendingSignature
		row        *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Scan-Fail-WrongNumberOfColumn",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				pendingSig: &model.PendingSignature{},
				row:        getMockScanRowFail(),
			},
			wantErr: true,
		},
		{
			name: "Scan-Success",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				pendingSig: &model.PendingSignature{},
				row:        getMockScanRowSuccess(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := pe.Scan(tt.args.pendingSig, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPendingSignatureQuery_getTableName(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getTableName",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			want: mockPendingSignatureQueryIntance.TableName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := psq.getTableName(); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

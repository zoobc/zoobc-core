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
	mockPendingTransactionQueryInstance = NewPendingTransactionQuery()
)

func TestNewPendingTransactionQuery(t *testing.T) {
	tests := []struct {
		name string
		want *PendingTransactionQuery
	}{
		{
			name: "NewPendingTransactionQuery-Success",
			want: &PendingTransactionQuery{
				Fields: []string{
					"sender_address",
					"transaction_hash",
					"transaction_bytes",
					"status",
					"block_height",
				},
				TableName: "pending_transaction",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPendingTransactionQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPendingTransactionQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mock PendingTransactionQueryBuildModel
func getPendingTransactionQueryBuildModelFailRow() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"RandomField"})
	mockRow.AddRow(
		make([]byte, 32),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}
func getPendingTransactionQueryBuildModelSuccessRow() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockPendingTransactionQueryInstance.Fields)
	mockRow.AddRow(
		"",
		make([]byte, 32),
		make([]byte, 100),
		model.PendingTransactionStatus_PendingTransactionExecuted,
		uint32(10),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

// mock PendingTransactionQueryBuildModel

func TestPendingTransactionQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pts  []*model.PendingTransaction
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PendingTransaction
		wantErr bool
	}{
		{
			name: "BuildModel-Fail",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pts:  []*model.PendingTransaction{},
				rows: getPendingTransactionQueryBuildModelFailRow(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "BuildModel-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pts:  []*model.PendingTransaction{},
				rows: getPendingTransactionQueryBuildModelSuccessRow(),
			},
			want: []*model.PendingTransaction{
				{
					SenderAddress:    "",
					TransactionHash:  make([]byte, 32),
					TransactionBytes: make([]byte, 100),
					Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
					BlockHeight:      10,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := ptq.BuildModel(tt.args.pts, tt.args.rows)
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
	mockPendingTransactionExtractModel = &model.PendingTransaction{
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 100),
		Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
		BlockHeight:      10,
	}
)

func TestPendingTransactionQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTx *model.PendingTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "ExtractModel",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: mockPendingTransactionExtractModel,
			},
			want: []interface{}{
				&mockPendingTransactionExtractModel.SenderAddress,
				&mockPendingTransactionExtractModel.TransactionHash,
				&mockPendingTransactionExtractModel.TransactionBytes,
				&mockPendingTransactionExtractModel.Status,
				&mockPendingTransactionExtractModel.BlockHeight,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := pe.ExtractModel(tt.args.pendingTx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingTransactionQuery_GetPendingTransactionByHash(t *testing.T) {
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
			name: "GetPendingTransactionByHash-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				txHash: make([]byte, 32),
			},
			wantStr: "SELECT sender_address, transaction_hash, transaction_bytes, status, block_height FROM pending_transaction " +
				"WHERE transaction_hash = ?",
			wantArgs: []interface{}{
				make([]byte, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ptq.GetPendingTransactionByHash(tt.args.txHash)
			if gotStr != tt.wantStr {
				t.Errorf("GetPendingTransactionByHash() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetPendingTransactionByHash() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

var (
	mockInsertPendingTransaction = &model.PendingTransaction{
		SenderAddress:    "",
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 100),
		Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
		BlockHeight:      10,
	}
)

func TestPendingTransactionQuery_InsertPendingTransaction(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTx *model.PendingTransaction
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "InsertPendingTransaction-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: mockInsertPendingTransaction,
			},
			wantStr: "INSERT INTO pending_transaction (sender_address, transaction_hash, transaction_bytes, " +
				"status, block_height) VALUES(? , ? , ? , ? , ? )",
			wantArgs: mockPendingTransactionQueryInstance.ExtractModel(mockInsertPendingTransaction),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ptq.InsertPendingTransaction(tt.args.pendingTx)
			if gotStr != tt.wantStr {
				t.Errorf("InsertPendingTransaction() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertPendingTransaction() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestPendingTransactionQuery_Rollback(t *testing.T) {
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
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				height: 10,
			},
			wantMultiQueries: [][]interface{}{
				{
					fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", mockPendingTransactionQueryInstance.TableName),
					uint32(10),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := ptq.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

// mock PendingTransactionQuery Scan
func getPendingTransactionQueryScanFailRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}
func getPendingTransactionQueryScanSuccessRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockPendingTransactionQueryInstance.Fields)
	mockRow.AddRow(
		"",
		make([]byte, 32),
		make([]byte, 100),
		uint32(0),
		uint32(10),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}

// mock PendingTransactionQuery Scan

func TestPendingTransactionQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTx *model.PendingTransaction
		row       *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Scan-Fail",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: &model.PendingTransaction{},
				row:       getPendingTransactionQueryScanFailRow(),
			},
			wantErr: true,
		},
		{
			name: "Scan-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: &model.PendingTransaction{},
				row:       getPendingTransactionQueryScanSuccessRow(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := pe.Scan(tt.args.pendingTx, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPendingTransactionQuery_getTableName(t *testing.T) {
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
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			want: mockPendingTransactionQueryInstance.TableName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := ptq.getTableName(); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

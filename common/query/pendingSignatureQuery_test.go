package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockPendingSignatureQueryIntance = NewPendingSignatureQuery()
	pendingSigAccountAddress1        = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
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
					"multisig_address",
					"account_address",
					"signature",
					"block_height",
					"latest",
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
		[]byte{},
		pendingSigAccountAddress1,
		make([]byte, 64),
		uint32(10),
		true,
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
					TransactionHash:       make([]byte, 32),
					MultiSignatureAddress: []byte{},
					AccountAddress:        pendingSigAccountAddress1,
					Signature:             make([]byte, 64),
					BlockHeight:           10,
					Latest:                true,
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
		AccountAddress:  pendingSigAccountAddress1,
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
				&mockExtractModelPendingSig.MultiSignatureAddress,
				&mockExtractModelPendingSig.AccountAddress,
				&mockExtractModelPendingSig.Signature,
				&mockExtractModelPendingSig.BlockHeight,
				&mockExtractModelPendingSig.Latest,
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
		txHash               []byte
		currentHeight, limit uint32
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
				txHash:        make([]byte, 32),
				currentHeight: 0,
				limit:         constant.MinRollbackBlocks,
			},
			wantStr: "SELECT transaction_hash, multisig_address, account_address, signature, block_height, latest FROM " +
				"pending_signature WHERE transaction_hash = ? AND block_height >= ? AND latest = true",
			wantArgs: []interface{}{
				make([]byte, 32),
				uint32(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := psq.GetPendingSignatureByHash(
				tt.args.txHash,
				tt.args.currentHeight,
				tt.args.limit,
			)
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
		AccountAddress:  nil,
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
		name   string
		fields fields
		args   args
		want   [][]interface{}
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
			want: [][]interface{}{
				append([]interface{}{"INSERT OR REPLACE INTO pending_signature (transaction_hash, multisig_address, account_address, " +
					"signature, block_height, latest) VALUES(? , ? , ? , ? , ? , ? )"},
					mockPendingSignatureQueryIntance.ExtractModel(mockInsertPendingSignaturePendingSig)...),
				{
					"UPDATE pending_signature SET latest = false WHERE account_address = ? AND transaction_hash = " +
						"? AND block_height != 0 AND latest = true",
					mockInsertPendingSignaturePendingSig.AccountAddress, mockInsertPendingSignaturePendingSig.TransactionHash,
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
			got := psq.InsertPendingSignature(tt.args.pendingSig)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertPendingSignature() gotArgs = \n%v, want \n%v", got, tt.want)
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
					"DELETE FROM pending_signature WHERE block_height > ?",
					uint32(10),
				},
				{
					"UPDATE pending_signature SET latest = ? WHERE latest = ? AND (account_address, transaction_hash, " +
						"block_height) IN (SELECT t2.account_address, t2.transaction_hash, " +
						"MAX(t2.block_height) FROM pending_signature as t2 GROUP BY t2.account_address, t2.transaction_hash)",
					1, 0,
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
		[]byte{},
		"account_address",
		make([]byte, 64),
		uint32(10),
		true,
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

func TestPendingSignatureQuery_SelectDataForSnapshot(t *testing.T) {
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
			name: "SelectDataForSnapshot",
			fields: fields{
				Fields:    mockPendingSignatureQueryIntance.Fields,
				TableName: mockPendingSignatureQueryIntance.TableName,
			},
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "SELECT transaction_hash,multisig_address,account_address,signature,block_height,latest FROM pending_signature WHERE (account_address, " +
				"transaction_hash, block_height) IN (SELECT t2.account_address, t2.transaction_hash, " +
				"MAX(t2.block_height) FROM pending_signature as t2 WHERE t2.block_height >= 1 AND t2.block_height <= 10 AND t2.block_height != 0 " +
				"GROUP BY t2.account_address, t2.transaction_hash) ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := psq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("PendingSignatureQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingSignatureQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			name: "TrimDataBeforeSnapshot",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM pending_transaction WHERE block_height >= 0 AND block_height <= 10 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := psq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("PendingSignatureQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingSignatureQuery_InsertPendingSignatures(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingSigs []*model.PendingSignature
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewPendingSignatureQuery()),
			args: args{
				pendingSigs: []*model.PendingSignature{
					mockInsertPendingSignaturePendingSig,
				},
			},
			wantStr: "INSERT INTO pending_signature (transaction_hash, multisig_address, account_address, signature, " +
				"block_height, latest) VALUES (?, ?, ?, ?, ?, ?)",
			wantArgs: NewPendingSignatureQuery().ExtractModel(mockInsertPendingSignaturePendingSig),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psq := &PendingSignatureQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := psq.InsertPendingSignatures(tt.args.pendingSigs)
			if gotStr != tt.wantStr {
				t.Errorf("InsertPendingSignatures() gotStr = \n%v, want \n%v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertPendingSignatures() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

package query

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockMultisigInfoQueryInstance = NewMultisignatureInfoQuery()
)

// mocks fixtures for MultisignatureInfoQuery_BuildModel
func getBuildModelErrorMockRows() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

func getBuildModelSuccessMockRows() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockMultisigInfoQueryInstance.Fields)
	mockRow.AddRow(
		"multisig_address",
		uint32(1),
		int64(10),
		uint32(12),
		true,
		"address_1,address_2",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

func TestMultisignatureInfoQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		mss  []*model.MultiSignatureInfo
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.MultiSignatureInfo
		wantErr bool
	}{
		{
			name: "BuildModel-RowsError",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				mss:  []*model.MultiSignatureInfo{},
				rows: getBuildModelErrorMockRows(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "BuildModel-RowsError",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				mss:  []*model.MultiSignatureInfo{},
				rows: getBuildModelSuccessMockRows(),
			},
			want: []*model.MultiSignatureInfo{
				{
					MultisigAddress:   "multisig_address",
					MinimumSignatures: 1,
					Nonce:             10,
					BlockHeight:       12,
					Latest:            true,
					Addresses:         []string{"address_1", "address_2"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := msi.BuildModel(tt.args.mss, tt.args.rows)
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
	// Extract mocks
	mockExtractMultisignatureInfoMultisig = &model.MultiSignatureInfo{
		MinimumSignatures: 0,
		Nonce:             0,
		Addresses:         []string{"A", "B"},
		MultisigAddress:   "",
		BlockHeight:       0,
		Latest:            true,
	}
	// Extract mocks
)

func TestMultisignatureInfoQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
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
				Fields:    nil,
				TableName: "",
			},
			args: args{
				multisigInfo: mockExtractMultisignatureInfoMultisig,
			},
			want: []interface{}{
				&mockExtractMultisignatureInfoMultisig.MultisigAddress,
				&mockExtractMultisignatureInfoMultisig.MinimumSignatures,
				&mockExtractMultisignatureInfoMultisig.Nonce,
				&mockExtractMultisignatureInfoMultisig.BlockHeight,
				&mockExtractMultisignatureInfoMultisig.Latest,
				strings.Join(mockExtractMultisignatureInfoMultisig.Addresses, ","),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mu.ExtractModel(tt.args.multisigInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_GetMultisignatureInfoByAddress(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigAddress      string
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
			name: "GetMultisignatureInfoByAddress-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigAddress: "A",
				currentHeight:   0,
				limit:           constant.MinRollbackBlocks,
			},
			wantStr: "SELECT multisig_address, minimum_signatures, nonce, block_height, latest, (" +
				"SELECT GROUP_CONCAT(account_address, ',') FROM %s GROUP BY multisig_address, block_height ORDER BY account_address_index DESC" +
				") as addresses FROM multisignature_info WHERE multisig_address = ? AND block_height >= ? AND latest = true",
			wantArgs: []interface{}{"A", uint32(0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := msi.GetMultisignatureInfoByAddress(
				tt.args.multisigAddress,
				tt.args.currentHeight,
				tt.args.limit,
			)
			if gotStr != tt.wantStr {
				t.Errorf("GetMultisignatureInfoByAddress() gotStr = \n%v, want \n%v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetMultisignatureInfoByAddress() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

var (
	// InsertMultisignatureInfo mocks
	mockInsertMultisignatureInfoMultisig = &model.MultiSignatureInfo{
		MinimumSignatures: 0,
		Nonce:             0,
		Addresses:         nil,
		MultisigAddress:   "",
		BlockHeight:       0,
		Latest:            true,
	}
	// InsertMultisignatureInfo mocks
)

func TestMultisignatureInfoQuery_InsertMultisignatureInfo(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name: "InsertMultisigInfo-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigInfo: mockInsertMultisignatureInfoMultisig,
			},
			want: [][]interface{}{
				append([]interface{}{
					"INSERT OR REPLACE INTO multisignature_info (multisig_address, minimum_signatures, nonce, block_height, latest) " +
						"VALUES(? , ? , ? , ? , ? )",
				}, mockMultisigInfoQueryInstance.ExtractModel(
					mockInsertMultisignatureInfoMultisig)[:len(NewMultisignatureInfoQuery().Fields)-1]...),
				{
					"UPDATE multisignature_info SET latest = false WHERE multisig_address = ? AND " +
						"block_height != 0 AND latest = true", mockInsertMultisignatureInfoMultisig.MultisigAddress,
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got := msi.InsertMultisignatureInfo(tt.args.multisigInfo)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertMultisignatureInfo() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_Rollback(t *testing.T) {
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
			name: "Rollback-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				height: 10,
			},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM multisignature_info WHERE block_height > ?",
					uint32(10),
				},
				{
					"UPDATE multisignature_info SET latest = ? WHERE latest = ? AND (multisig_address, " +
						"block_height) IN (SELECT t2.multisig_address, MAX(t2.block_height) FROM multisignature_info as t2 GROUP BY t2." +
						"multisig_address)",
					1, 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := msi.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

// mocks fixtures for MultisignatureInfoQuery_Scan
func getNumberScanFailMockRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}

func getNumberScanSuccessMockRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockMultisigInfoQueryInstance.Fields)
	mockRow.AddRow(
		"multisig_address",
		uint32(123),
		int64(10),
		uint32(12),
		true,
		"addresses",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}

// mocks fixtures for MultisignatureInfoQuery_Scan

func TestMultisignatureInfoQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
		row          *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "MultisignatureInfoQuery_Scan-Fail-WrongNumberOfColumn",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigInfo: &model.MultiSignatureInfo{},
				row:          getNumberScanFailMockRow(),
			},
			wantErr: true,
		},
		{
			name: "MultisignatureInfoQuery_Scan-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigInfo: &model.MultiSignatureInfo{},
				row:          getNumberScanSuccessMockRow(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := mu.Scan(tt.args.multisigInfo, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultisignatureInfoQuery_getTableName(t *testing.T) {
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
				Fields:    nil,
				TableName: "X",
			},
			want: "X",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.getTableName(); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	// NewMultisignatureInfoQuery mocks
	mockNewMultisignatureInfoQueryResult = &MultisignatureInfoQuery{
		Fields: []string{
			"multisig_address",
			"minimum_signatures",
			"nonce",
			"block_height",
			"latest",
			"addresses",
		},
		TableName: "multisignature_info",
	}
	// NewMultisignatureInfoQuery mocks
)

func TestNewMultisignatureInfoQuery(t *testing.T) {
	tests := []struct {
		name string
		want *MultisignatureInfoQuery
	}{
		{
			name: "NewMultisignatureInfoQuery",
			want: mockNewMultisignatureInfoQueryResult,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMultisignatureInfoQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMultisignatureInfoQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_SelectDataForSnapshot(t *testing.T) {
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
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "SELECT multisig_address, minimum_signatures, nonce, block_height, latest, (" +
				"SELECT GROUP_CONCAT(account_address, ',') FROM multisignature_participant " +
				"GROUP BY multisig_address, block_height ORDER BY account_address_index ASC" +
				") as addresses FROM multisignature_info " +
				"WHERE (multisig_address, block_height) IN (" +
				"SELECT t2.multisig_address, MAX(t2.block_height) " +
				"FROM multisignature_info as t2 WHERE t2.block_height >= 1 AND t2.block_height <= 10 GROUP BY t2.multisig_address" +
				") ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("MultisignatureInfoQuery.SelectDataForSnapshot() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "DELETE FROM multisignature_info WHERE block_height >= 1 AND block_height <= 10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("MultisignatureInfoQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

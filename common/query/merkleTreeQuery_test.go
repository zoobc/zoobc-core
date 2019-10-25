package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var (
	mockMerkleTreeQuery = NewMerkleTreeQuery()
	mockRoot            = make([]byte, 32)
	mockTree            = make([]byte, 14*32)
)

func TestMerkleTreeQuery_SelectMerkleTree(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		lowerHeight uint32
		upperHeight uint32
		limit       uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "SelectMerkleTree:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				lowerHeight: 0,
				upperHeight: 10,
				limit:       20,
			},
			want: "SELECT id, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS (SELECT rmr_linked FROM " +
				"published_receipt AS pr WHERE mt.id = pr.rmr_linked AND block_height BETWEEN 0 AND 10 " +
				"ORDER BY block_height ASC) LIMIT 20",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mrQ.SelectMerkleTree(tt.args.lowerHeight, tt.args.upperHeight, tt.args.limit); got != tt.want {
				t.Errorf("SelectMerkleTree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMerkleTreeQuery(t *testing.T) {
	tests := []struct {
		name string
		want MerkleTreeQueryInterface
	}{
		{
			name: "NewMerkleTreeQuery:success",
			want: mockMerkleTreeQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMerkleTreeQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMerkleTreeQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_InsertMerkleTree(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		root      []byte
		tree      []byte
		timestamp int64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name: "InsertMerkleTreeQuery:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				root:      mockRoot,
				tree:      mockTree,
				timestamp: 0,
			},
			wantQStr: "INSERT INTO merkle_tree (id, tree, timestamp) VALUES(?,? ,? )",
			wantArgs: []interface{}{
				mockRoot,
				mockTree,
				int64(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := mrQ.InsertMerkleTree(tt.args.root, tt.args.tree, tt.args.timestamp)
			if gotQStr != tt.wantQStr {
				t.Errorf("InsertMerkleTree() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertMerkleTree() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestMerkleTreeQuery_GetMerkleTreeByRoot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		root []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name: "GetMerkleTreeByRoot:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				root: mockRoot,
			},
			wantQStr: "SELECT id, tree, timestamp FROM merkle_tree WHERE id = ?",
			wantArgs: []interface{}{mockRoot},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := mrQ.GetMerkleTreeByRoot(tt.args.root)
			if gotQStr != tt.wantQStr {
				t.Errorf("GetMerkleTreeByRoot() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetMerkleTreeByRoot() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestMerkleTreeQuery_GetLastMerkleRoot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name     string
		fields   fields
		wantQStr string
	}{
		{
			name: "GetLastMerkleRoot:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			wantQStr: "SELECT id, tree, timestamp FROM merkle_tree ORDER BY timestamp DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotQStr := mrQ.GetLastMerkleRoot(); gotQStr != tt.wantQStr {
				t.Errorf("GetLastMerkleRoot() = %v, want %v", gotQStr, tt.wantQStr)
			}
		})
	}
}

func TestMerkleTreeQuery_ScanTree(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockMerkleTreeQuery.Fields).AddRow(
		mockRoot,
		mockTree,
		int64(0),
	))
	mock.ExpectQuery("wrongQuery").WillReturnRows(sqlmock.NewRows([]string{"foo"}).AddRow(
		mockRoot,
	))
	row := db.QueryRow("")
	wrongRow := db.QueryRow("wrongQuery")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		row *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ScanTree:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: row,
			},
			want:    mockTree,
			wantErr: false,
		},
		{
			name: "ScanTree:wrongColumn",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: wrongRow,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := mrQ.ScanTree(tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanTree() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_ScanRoot(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockMerkleTreeQuery.Fields).AddRow(
		mockRoot,
		mockTree,
		int64(0),
	))
	mock.ExpectQuery("wrongQuery").WillReturnRows(sqlmock.NewRows([]string{"foo"}).AddRow(
		mockRoot,
	))
	row := db.QueryRow("")
	wrongRow := db.QueryRow("wrongQuery")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		row *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ScanRoot:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: row,
			},
			want:    mockRoot,
			wantErr: false,
		},
		{
			name: "ScanRoot:wrongColumn",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: wrongRow,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := mrQ.ScanRoot(tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanRoot() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_BuildTree(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockMerkleTreeQuery.Fields).AddRow(
		mockRoot,
		mockTree,
		int64(0),
	))

	rows, _ := db.Query("")
	defer rows.Close()
	mock.ExpectQuery("WrongQuery").WillReturnRows(sqlmock.NewRows([]string{"foo"}).AddRow(
		mockRoot,
	))

	rowsWrong, _ := db.Query("WrongQuery")
	defer rowsWrong.Close()
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string][]byte
		wantErr bool
	}{
		{
			name: "BuildTree:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				rows: rows,
			},
			want: map[string][]byte{
				string(mockRoot): mockTree,
			},
			wantErr: false,
		},
		{
			name: "BuildTree:fail - wrong number field",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				rows: rowsWrong,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := mrQ.BuildTree(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildTree() got = %v, want %v", got, tt.want)
			}
		})
	}
}

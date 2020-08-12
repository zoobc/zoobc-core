package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNodeAdmissionTimestampQuery_GetNextNodeAdmision(t *testing.T) {
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
			name:   "wantSuccess",
			fields: fields(*NewNodeAdmissionTimestampQuery()),
			want:   "SELECT timestamp, block_height, latest FROM node_admission_timestamp WHERE latest = true  ORDER BY block_height DESC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natq := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := natq.GetNextNodeAdmision(); got != tt.want {
				t.Errorf("NodeAdmissionTimestampQuery.GetNextNodeAdmision() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockodeAdmissionTimestampQuery = &NodeAdmissionTimestampQuery{
		Fields: []string{
			"timestamp",
			"block_height",
			"latest",
		},
		TableName: "node_admission_timestamp",
	}
	mockNextAdmission = model.NodeAdmissionTimestamp{
		Timestamp:   1,
		BlockHeight: 1,
		Latest:      true,
	}
)

func TestNodeAdmissionTimestampQuery_InsertNextNodeAdmission(t *testing.T) {

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nextNodeAdmission *model.NodeAdmissionTimestamp
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*NewNodeAdmissionTimestampQuery()),
			args: args{
				nextNodeAdmission: &mockNextAdmission,
			},
			want: [][]interface{}{
				{
					fmt.Sprintf(`
				UPDATE %s SET latest = ? 
				WHERE latest = ? AND block_height IN (
					SELECT MAX(t2.block_height) FROM %s as t2
				)`,
						mockodeAdmissionTimestampQuery.TableName,
						mockodeAdmissionTimestampQuery.TableName,
					),
					0,
					1,
				},
				append(
					[]interface{}{
						fmt.Sprintf(
							"INSERT INTO %s (%s) VALUES(%s)",
							mockodeAdmissionTimestampQuery.TableName,
							strings.Join(mockodeAdmissionTimestampQuery.Fields, ", "),
							fmt.Sprintf("? %s", strings.Repeat(", ?", len(mockodeAdmissionTimestampQuery.Fields)-1)),
						),
					},
					mockodeAdmissionTimestampQuery.ExtractModel(&mockNextAdmission)...,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natq := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := natq.InsertNextNodeAdmission(tt.args.nextNodeAdmission); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdmissionTimestampQuery.InsertNextNodeAdmission() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestNodeAdmissionTimestampQuery_BuildModel(t *testing.T) {
	var (
		mockQuery   = NewNodeAdmissionTimestampQuery()
		db, mock, _ = sqlmock.New()
	)
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(mockQuery.Fields).
		AddRow(mockNextAdmission.Timestamp, mockNextAdmission.BlockHeight, mockNextAdmission.Latest))
	mockRows, _ := db.Query("foo")
	defer db.Close()

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nodeAdmissionTimestamps []*model.NodeAdmissionTimestamp
		rows                    *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAdmissionTimestamp
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields{},
			args: args{
				nodeAdmissionTimestamps: []*model.NodeAdmissionTimestamp{},
				rows:                    mockRows,
			},
			want:    []*model.NodeAdmissionTimestamp{&mockNextAdmission},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := n.BuildModel(tt.args.nodeAdmissionTimestamps, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdmissionTimestampQuery.BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdmissionTimestampQuery.BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdmissionTimestampQuery_Scan(t *testing.T) {
	var (
		mockQuery   = NewNodeAdmissionTimestampQuery()
		db, mock, _ = sqlmock.New()
	)
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(mockQuery.Fields).
		AddRow(mockNextAdmission.Timestamp, mockNextAdmission.BlockHeight, mockNextAdmission.Latest))
	mockRow := db.QueryRow("foo")
	defer db.Close()

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nextNodeAdmission *model.NodeAdmissionTimestamp
		row               *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				nextNodeAdmission: &model.NodeAdmissionTimestamp{},
				row:               mockRow,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natq := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := natq.Scan(tt.args.nextNodeAdmission, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("NodeAdmissionTimestampQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAdmissionTimestampQuery_InsertNextNodeAdmissions(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		nodeAdmissionTimestamps []*model.NodeAdmissionTimestamp
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
			fields: fields(*NewNodeAdmissionTimestampQuery()),
			args: args{
				nodeAdmissionTimestamps: []*model.NodeAdmissionTimestamp{
					&mockNextAdmission,
				},
			},
			wantStr: "INSERT INTO node_admission_timestamp (timestamp, block_height, latest) VALUES (?, ?, ?)",
			wantArgs: []interface{}{
				mockNextAdmission.Timestamp,
				mockNextAdmission.BlockHeight,
				mockNextAdmission.Latest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natq := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := natq.InsertNextNodeAdmissions(tt.args.nodeAdmissionTimestamps)
			if gotStr != tt.wantStr {
				t.Errorf("NodeAdmissionTimestampQuery.InsertNextNodeAdmissions() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("NodeAdmissionTimestampQuery.InsertNextNodeAdmissions() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestNodeAdmissionTimestampQuery_SelectDataForSnapshot(t *testing.T) {
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
			name:   "wantSucess",
			fields: fields(*NewNodeAdmissionTimestampQuery()),
			args: args{
				fromHeight: 1,
				toHeight:   2,
			},
			want: "SELECT timestamp,block_height,latest FROM node_admission_timestamp " +
				"WHERE block_height >= 1 AND block_height <= 2 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natq := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := natq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("NodeAdmissionTimestampQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdmissionTimestampQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			fields: fields(*NewNodeAdmissionTimestampQuery()),
			args: args{
				fromHeight: 1,
				toHeight:   2,
			},
			want: "DELETE FROM node_admission_timestamp WHERE block_height >= 1 AND block_height <= 2 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			natq := &NodeAdmissionTimestampQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := natq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("NodeAdmissionTimestampQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestMultiSignatureParticipantQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		participant *model.MultiSignatureParticipant
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				participant: &model.MultiSignatureParticipant{
					MultiSignatureAddress: "MSG_",
					AccountAddressIndex:   0,
					AccountAddress:        "BCZ_",
					BlockHeight:           100,
				},
			},
			want: []interface{}{
				"MSG_",
				"BCZ_",
				uint32(0),
				false,
				uint32(100),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msq.ExtractModel(tt.args.participant); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_BuildModel(t *testing.T) {
	dbMock, sqlMock, _ := sqlmock.New()
	sqlMock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(NewMultiSignatureParticipantQuery().Fields).
		AddRow(
			"MZG_",
			"BCZ_",
			0,
			true,
			100,
		))
	rows, _ := dbMock.Query("")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantParticipants []*model.MultiSignatureParticipant
		wantErr          bool
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				rows: rows,
			},
			wantParticipants: []*model.MultiSignatureParticipant{
				{
					MultiSignatureAddress: "MZG_",
					AccountAddress:        "BCZ_",
					AccountAddressIndex:   0,
					Latest:                true,
					BlockHeight:           100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotParticipants, err := msq.BuildModel(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotParticipants, tt.wantParticipants) {
				t.Errorf("BuildModel() gotParticipants = %v, want %v", gotParticipants, tt.wantParticipants)
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_Scan(t *testing.T) {
	dbMock, sqlMock, _ := sqlmock.New()
	sqlMock.ExpectQuery("").WillReturnRows(sqlMock.NewRows(NewMultiSignatureParticipantQuery().Fields).
		AddRow(
			"MZG_",
			"BCZ_",
			0,
			true,
			100,
		))
	row := dbMock.QueryRow("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		participant *model.MultiSignatureParticipant
		row         *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				participant: &model.MultiSignatureParticipant{},
				row:         row,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := msq.Scan(tt.args.participant, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ok := reflect.DeepEqual(tt.args.participant, new(model.MultiSignatureParticipant)); ok {
				t.Errorf("Scan() did not update reference")
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_InsertMultisignatureParticipants(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		participants []*model.MultiSignatureParticipant
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantQueries [][]interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				participants: []*model.MultiSignatureParticipant{
					{
						MultiSignatureAddress: "MSG_",
						AccountAddressIndex:   0,
						AccountAddress:        "BCZ_0",
						BlockHeight:           100,
						Latest:                true,
					},
					{
						MultiSignatureAddress: "MSG_",
						AccountAddressIndex:   1,
						AccountAddress:        "BCZ_1",
						BlockHeight:           100,
						Latest:                true,
					},
				},
			},
			wantQueries: [][]interface{}{
				{
					"INSERT OR REPLACE INTO multisignature_participant (multisig_address, account_address, account_address_index, latest, block_height) " +
						"VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
					"MSG_", "BCZ_0", uint32(0), true, uint32(100), "MSG_", "BCZ_1", uint32(1), true, uint32(100),
				},
				{
					"UPDATE multisignature_participant SET latest = ? WHERE multisig_address = ? AND block_height != ? AND latest = ?",
					false, "MSG_", uint32(100), true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotQueries := msq.InsertMultisignatureParticipants(tt.args.participants); !reflect.DeepEqual(gotQueries, tt.wantQueries) {
				t.Errorf("InsertMultisignatureParticipants() = \n%v, want \n%v", gotQueries, tt.wantQueries)
			}
		})
	}
}

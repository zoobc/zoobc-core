// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package database

import (
	"database/sql"
	"github.com/zoobc/zoobc-core/common/queue"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorInitFailRows struct {
		query.Executor
	}
)

func (*mockExecutorInitFailRows) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {

	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(qe).WillReturnRows(
		sqlmock.NewRows([]string{"count"}),
	)
	return db.QueryRow(qe), nil
}

func (*mockExecutorInitFailRows) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Version",
	}).AddRow(1))

	rows, _ := db.Query(qe)
	return rows, nil
}

func TestMigration_Init(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	type fields struct {
		Versions []string
		Query    query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: query.NewQueryExecutor(db, queue.NewPriorityPreferenceLock()),
			},
			wantErr: false,
		},
		{
			name: "wantError",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: nil,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:versionNotNil",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: &mockExecutorInitFailRows{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				Versions: tt.fields.Versions,
				Query:    tt.fields.Query,
			}
			if err := m.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Migration.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var dbMock, mock, _ = sqlmock.New()

type (
	mockQueryExecutorVersionNil struct {
		query.Executor
	}
)

func (*mockQueryExecutorVersionNil) BeginTx(params ...int) error {
	return nil
}
func (*mockQueryExecutorVersionNil) ExecuteTransaction(qStr string, args ...interface{}) error {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectPrepare(regexp.QuoteMeta(qStr)).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	stmt, err := db.Prepare(qStr)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(args...)
	return err
}
func (*mockQueryExecutorVersionNil) CommitTx() error {
	return nil
}
func TestMigration_Apply(t *testing.T) {
	currentVersion := 1

	type fields struct {
		CurrentVersion *int
		Versions       []string
		Query          query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "accounts" (
						id	INTEGER,
						public_key	BLOB  NOT NULL,
						PRIMARY KEY("public_key")
					);`,
					`CREATE TABLE IF NOT EXISTS "accounts" (
						id	INTEGER,
						public_key	BLOB  NOT NULL,
						PRIMARY KEY("public_key")
					);`,
				},
				CurrentVersion: &currentVersion,
				Query: &mockQueryExecutorVersionNil{
					query.Executor{Db: dbMock, Lock: queue.NewPriorityPreferenceLock()},
				},
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:VersionNil",
			fields: fields{
				CurrentVersion: nil,
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "account" (
						id	INTEGER,
						public_key	BLOB  NOT NULL,
						PRIMARY KEY("public_key")
					);`,
				},
				Query: &mockQueryExecutorVersionNil{
					query.Executor{Db: dbMock, Lock: queue.NewPriorityPreferenceLock()},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				CurrentVersion: tt.fields.CurrentVersion,
				Versions:       tt.fields.Versions,
				Query:          tt.fields.Query,
			}
			if err := m.Apply(); (err != nil) != tt.wantErr {
				t.Errorf("Migration.Apply() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Migration.Apply() query: \n%s, want: \n%s", tt.fields.Versions[0], err)
			}
		})
	}
}

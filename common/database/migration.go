package database

import (
	"fmt"

	"github.com/zoobc/zoobc-core/common/query"
)

/*
Migration is struct that included:
	- Init() initialization
	- Apply() run migrations
migration should be has `query.Executor` interface
*/
type Migration struct {
	CurrentVersion *int
	Versions       []string
	Query          query.ExecutorInterface
}

/*
Init function must be call at the first time before call `Apply()`.
That just for make sure no error that caused by `query.Executor` not `nil`
and initialize versions
*/
func (m *Migration) Init() error {

	if m.Query != nil {
		rows, _ := m.Query.ExecuteSelect("SELECT version FROM migration;")
		if rows != nil {
			var version int
			_ = rows.Scan(&version)
			m.CurrentVersion = &version
		}

		m.Versions = []string{
			`CREATE TABLE IF NOT EXISTS "migration" (
				"version" INTEGER DEFAULT 0 NOT NULL,
				"created_date" TIMESTAMP NOT NULL
			);`,
			`
			CREATE TABLE IF NOT EXISTS "mempool" (
				"id"	INTEGER,
				"fee_per_byte"	INTEGER,
				"arrival_timestamp"	INTEGER,
				"transaction_bytes"	BLOB,
				PRIMARY KEY("id")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "transaction" (
				"id"	INTEGER,
				"block_id"	INTEGER,
				"block_height"	INTEGER,
				"sender_account_type"	INTEGER,
				"sender_account_address"	TEXT,
				"recipient_account_type"	INTEGER,
				"recipient_account_address"	TEXT,
				"transaction_type"	INTEGER,
				"fee"	INTEGER,
				"timestamp"	INTEGER,
				"transaction_hash"	BLOB,
				"transaction_body_length"	INTEGER,
				"transaction_body_bytes"	BLOB,
				"signature"	BLOB,
				"version" INTEGER,
				PRIMARY KEY("id")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "account" (
				"id"	BLOB,
				"account_type"	INTEGER,
				"address"	TEXT,
				PRIMARY KEY("id")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "account_balance" (
				"account_id"	BLOB,
				"block_height"	INTEGER,
				"spendable_balance"	INTEGER,
				"balance"	INTEGER,
				"pop_revenue"	INTEGER,
				"latest"	INTEGER,
				PRIMARY KEY("account_id","block_height"),
				FOREIGN KEY("account_id") REFERENCES account(id)
			);`,
			`
			CREATE TABLE IF NOT EXISTS "main_block" (
				"id" INTEGER,
				"previous_block_hash" BLOB,
				"height" INTEGER,
				"timestamp" INTEGER,
				"block_seed" BLOB,
				"block_signature" BLOB,
				"cumulative_difficulty" TEXT,
				"smith_scale" INTEGER,
				"blocksmith_id" BLOB,
				"total_amount" INTEGER,
				"total_fee" INTEGER,
				"total_coinbase" INTEGER,
				"version" INTEGER,
				"payload_length" INTEGER,
				"payload_hash" BLOB,
				PRIMARY KEY("id")
			);`,
		}
		return nil
	}
	return fmt.Errorf("make sure have add query.Executor")

}

/*
Apply for applying migrations that had initialize on `Init()`.
And this will create migration table included version of migration
*/
func (m *Migration) Apply() error {

	var (
		migrations = m.Versions
	)

	if m.CurrentVersion != nil {
		migrations = m.Versions[*m.CurrentVersion:]
	}

	for version, query := range migrations {
		version := version
		_ = m.Query.BeginTx()
		_ = m.Query.ExecuteTransaction(query)

		if m.CurrentVersion != nil {
			_ = m.Query.ExecuteTransaction(`UPDATE "migration"
				SET "version" = ?, "created_date" = datetime('now');`, *m.CurrentVersion)
		} else {
			_ = m.Query.ExecuteTransaction(`
				INSERT INTO "migration" (
					"version",
					"created_date"
				)
				VALUES (
					0,
					datetime('now')
				);
				`)
		}
		err := m.Query.CommitTx()
		m.CurrentVersion = &version
		if err != nil {
			return err
		}
	}
	return nil
}

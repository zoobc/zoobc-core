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
	Query          *query.Executor
}

/*
Init function must be call at the first time before call `Apply()`.
That just for make sure no error that caused by `query.Executor` not `nil`
and initialize versions
*/
func (m *Migration) Init(qe *query.Executor) error {

	if qe != nil {
		rows, _ := qe.ExecuteSelect("SELECT version FROM migration;")
		if rows != nil {
			var version int
			_ = rows.Scan(&version)
			m.CurrentVersion = &version
		}

		m.Query = qe
		m.Versions = []string{
			`CREATE TABLE IF NOT EXISTS "migration" (
				"version" INTEGER DEFAULT 0 NOT NULL,
				"created_date" TIMESTAMP NOT NULL
			);`,
			`
			CREATE TABLE IF NOT EXISTS "mempool" (
				"id"	BLOB,
				"fee_per_byte"	INTEGER,
				"arrival_timestamp"	INTEGER,
				"transaction_bytes"	BLOB,
				PRIMARY KEY("id")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "transaction" (
				"id"	BLOB,
				"block_id"	INTEGER,
				"block_height"	INTEGER,
				"sender_account_id"	BLOB,
				"recipient_account_id"	BLOB,
				"transaction_type"	INTEGER,
				"fee"	INTEGER,
				"timestamp"	INTEGER,
				"transaction_hash"	BLOB,
				"transaction_body_length"	INTEGER,
				"transaction_body_bytes"	BLOB,
				"signature"	BLOB,
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
				"id"	BLOB,
				"block_height"	INTEGER,
				"spendable_balance"	INTEGER,
				"balance"	INTEGER,
				"pop_revenue"	INTEGER,
				"latest"	INTEGER,
				PRIMARY KEY("id","block_height"),
				FOREIGN KEY("id") REFERENCES account(id)
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
			`
			CREATE TABLE IF NOT EXISTS "account_balance" (
				"account_id"	BLOB,
				"balance"	INTEGER,
				"spendable_balance"	INTEGER,
				"pop_revenue"	INTEGER,
				"block_height"	INTEGER,
				"latest"	INTEGER
			);
			`,
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
		queries := []string{
			query,
		}
		if m.CurrentVersion != nil {
			queries = append(queries, fmt.Sprintf(`
				UPDATE "migration"
				SET "version" = %d, "created_date" = datetime('now');
			`, *m.CurrentVersion))
		} else {
			queries = append(queries, `
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
		_, err := m.Query.ExecuteTransactions(queries)
		m.CurrentVersion = &version
		if err != nil {
			return err
		}
	}
	return nil
}

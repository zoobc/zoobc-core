package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
		var version int
		row, _ := m.Query.ExecuteSelectRow("SELECT version FROM migration;", false)
		err := row.Scan(&version)
		if err == nil {
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
				"sender_account_address" VARCHAR(255),
				"recipient_account_address" VARCHAR(255),
				PRIMARY KEY("id")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "transaction" (
				"id"	INTEGER,
				"block_id"	INTEGER,
				"block_height"	INTEGER,
				"sender_account_address"	VARCHAR(255),
				"recipient_account_address"	VARCHAR(255),
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
			CREATE TABLE IF NOT EXISTS "account_balance" (
				"account_address"	VARCHAR(255),
				"block_height"	INTEGER,
				"spendable_balance"	INTEGER,
				"balance"	INTEGER,
				"pop_revenue"	INTEGER,
				"latest"	INTEGER,
				PRIMARY KEY("account_address","block_height")
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
				"blocksmith_public_key" BLOB,
				"total_amount" INTEGER,
				"total_fee" INTEGER,
				"total_coinbase" INTEGER,
				"version" INTEGER,
				"payload_length" INTEGER,
				"payload_hash" BLOB,
				UNIQUE("height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "node_registry" (
				"id" INTEGER,
				"node_public_key" BLOB,
				"account_address" VARCHAR(255),
				"registration_height" INTEGER,
				"node_address" VARCHAR(255),
				"locked_balance" INTEGER,
				"queued" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("id", "height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "account_dataset"(
				"setter_account_address" VARCHAR(255),
				"recipient_account_address" VARCHAR(255),
				"property" TEXT,
				"value" TEXT,
				"timestamp_starts" INTEGER,
				"timestamp_expires" INTEGER,
				"height" INTEGER,
				"latest" INTEGER,
				PRIMARY KEY("setter_account_address","recipient_account_address", "property", "height")
			);`,
			`
			ALTER TABLE "transaction"
				ADD COLUMN "transaction_index" INTEGER AFTER "version"
			`,
			`
			CREATE TABLE IF NOT EXISTS "participation_score"(
				"node_id" INTEGER,
				"score" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("node_id", "height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "node_receipt" (
				"sender_public_key" BLOB,
				"recipient_public_key" BLOB,
				"datum_type" INTEGER,
				"datum_hash" BLOB,
				"reference_block_height" INTEGER,
				"reference_block_hash" BLOB,
				"rmr_linked" BLOB,
				"recipient_signature" BLOB,
				"rmr" BLOB,
				"rmr_index" INTEGER
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "batch_receipt" (
				"sender_public_key" BLOB,
				"recipient_public_key" BLOB,
				"datum_type" INTEGER,
				"datum_hash" BLOB,
				"reference_block_height" INTEGER,
				"reference_block_hash" BLOB,
				"rmr_linked" BLOB,
				"recipient_signature" BLOB
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "merkle_tree" (
				"id" BLOB,
				"tree" BLOB
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "published_receipt" (
				"sender_public_key" BLOB,
				"recipient_public_key" BLOB,
				"datum_type" INTEGER,
				"datum_hash" BLOB,
				"reference_block_height" INTEGER,
				"reference_block_hash" BLOB,
				"rmr_linked" BLOB,
				"recipient_signature" BLOB,
				"intermediate_hashes" BLOB,
				"block_height" INTEGER,
				"receipt_index" INTEGER,
				"published_index" INTEGER
			)
			`,
			`
			ALTER TABLE "merkle_tree"
				ADD COLUMN "timestamp" INTEGER AFTER "tree"
			`,
			`
			ALTER TABLE "node_registry" 
				RENAME COLUMN "queued" TO "registration_status"
			`,
			`
			AlTER TABLE "main_block"
				ADD COLUMN "block_hash" BLOB AFTER "id"
			`,
			`
			CREATE TABLE IF NOT EXISTS "skipped_blocksmith" (
				"blocksmith_public_key" BLOB,
				"pop_change" INTEGER,
				"block_height" INTEGER,
				"blocksmith_index" INTEGER
			)
			`,
			`
			ALTER TABLE "mempool"
				ADD COLUMN "block_height" INTEGER AFTER "id"
			`,
			`
			ALTER TABLE "merkle_tree"
				ADD COLUMN "block_height" INTEGER AFTER "id"
			`,
			`
			CREATE TABLE IF NOT EXISTS "spine_block" (
				"id" INTEGER,
				"block_hash" BLOB,
				"previous_block_hash" BLOB,
				"height" INTEGER,
				"timestamp" INTEGER,
				"block_seed" BLOB,
				"block_signature" BLOB,
				"cumulative_difficulty" TEXT,
				"blocksmith_public_key" BLOB,
				"total_amount" INTEGER,
				"total_fee" INTEGER,
				"total_coinbase" INTEGER,
				"version" INTEGER,
				"payload_length" INTEGER,
				"payload_hash" BLOB,
				UNIQUE("height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "spine_public_key"(
				"node_public_key" BLOB,
				"block_id"	INTEGER,
				"public_key_action" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("node_public_key", "height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "account_ledger" (
				"account_address" VARCHAR(255) NULL,
				"balance_change" INTEGER,
				"block_height" INTEGER,
				"transaction_id" INTEGER NULL,
				"event_type" INTEGER
			)
			`,
			`
			ALTER TABLE "account_ledger"
				ADD COLUMN "timestamp" INTEGER
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
		err        error
	)

	if m.CurrentVersion != nil {
		migrations = m.Versions[*m.CurrentVersion+1:]
	}

	for v, qStr := range migrations {
		version := v
		err = m.Query.BeginTx()
		if err != nil {
			return err
		}
		err = m.Query.ExecuteTransaction(qStr)
		if err != nil {
			rollbackErr := m.Query.RollbackTx()
			if rollbackErr != nil {
				log.Errorln(rollbackErr.Error())
			}
			return err
		}
		if m.CurrentVersion != nil {
			err = m.Query.ExecuteTransaction(`UPDATE "migration"
				SET "version" = ?, "created_date" = datetime('now');`, *m.CurrentVersion+1)
			if err != nil {
				return err
			}
		} else {
			err = m.Query.ExecuteTransaction(`
				INSERT INTO "migration" (
					"version",
					"created_date"
				)
				VALUES (
					0,
					datetime('now')
				);
			`)
			if err != nil {
				return err
			}
		}

		m.CurrentVersion = &version
		err = m.Query.CommitTx()
		if err != nil {
			return err
		}
	}
	return nil
}

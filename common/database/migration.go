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
				"block_height" INTEGER,
				"fee_per_byte"	INTEGER,
				"arrival_timestamp"	INTEGER,
				"transaction_bytes"	BLOB,
				"sender_account_address" VARCHAR(255),
				"recipient_account_address" VARCHAR(255),
				PRIMARY KEY("id")
			);`,
			`CREATE INDEX "idx_sender_mempool" 
			ON "mempool" ("sender_account_address");`,
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
				"transaction_index" INTEGER,
				PRIMARY KEY("id")
			);`,

			`CREATE INDEX "idx_blockid_transaction" 
			ON "transaction" ("block_id");`,
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
			`CREATE INDEX "idx_blockheight_account_balance" 
			ON "account_balance" ("block_height");`,
			`
			CREATE TABLE IF NOT EXISTS "main_block" (
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
			`CREATE INDEX "idx_height_main_block" 
			ON "main_block" ("height");`,
			`
			CREATE TABLE IF NOT EXISTS "node_registry" (
				"id" INTEGER,
				"node_public_key" BLOB,
				"account_address" VARCHAR(255),
				"registration_height" INTEGER,
				"node_address" VARCHAR(255),
				"locked_balance" INTEGER,
				"registration_status" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("id")
			);`,
			`CREATE INDEX "idx_height_node_registry" 
			ON "node_registry" ("height");`,
			`CREATE INDEX "idx_nodeaddr_node_registry" 
			ON "node_registry" ("node_address");`,
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
			`CREATE INDEX "idx_recipient_account_dataset" 
			ON "account_dataset" ("recipient_account_address");`,
			`CREATE INDEX "idx_property_account_dataset" 
			ON "account_dataset" ("property");`,
			`CREATE INDEX "idx_height_account_dataset" 
			ON "account_dataset" ("height");`,
			`
			CREATE TABLE IF NOT EXISTS "participation_score"(
				"node_id" INTEGER,
				"score" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("node_id","height")
			);`,
			`CREATE INDEX "idx_height_participation_score" 
			ON "participation_score" ("height");`,
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
				"rmr_index" INTEGER,
				PRIMARY KEY ("sender_public_key")
			)
			`,
			`CREATE INDEX "idx_height_node_receipt" 
			ON "node_receipt" ("sender_public_key");`,
			`CREATE INDEX "idx_rmr_node_receipt" 
			ON "node_receipt" ("rmr_index");`,
			`
			CREATE TABLE IF NOT EXISTS "batch_receipt" (
				"sender_public_key" BLOB,
				"recipient_public_key" BLOB,
				"datum_type" INTEGER,
				"datum_hash" BLOB,
				"reference_block_height" INTEGER,
				"reference_block_hash" BLOB,
				"rmr_linked" BLOB,
				"recipient_signature" BLOB,
				PRIMARY KEY ("sender_public_key")
			)
			`,
			`CREATE INDEX "idx_sender_batch_receipt" 
			ON "batch_receipt" ("sender_public_key");`,
			`
			CREATE TABLE IF NOT EXISTS "merkle_tree" (
				"id" BLOB,
				"block_height" INTEGER,
				"tree" BLOB,
				"timestamp" INTEGER,
				PRIMARY KEY("id")
			)
			`,
			`CREATE INDEX "idx_merkle" 
			ON "merkle_tree" ("id");`,
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
			`CREATE INDEX "idx_pubindex_published_receipt" 
			ON "published_receipt" ("published_index");
			`,
			`CREATE INDEX "idx_sender_published_receipt" 
			ON "published_receipt" ("sender_public_key");
			`,
			`
			CREATE TABLE IF NOT EXISTS "skipped_blocksmith" (
				"blocksmith_public_key" BLOB,
				"pop_change" INTEGER,
				"block_height" INTEGER,
				"blocksmith_index" INTEGER
			)
			`,
			`CREATE INDEX "idx_pubkey_skipped_blocksmith" 
			ON "skipped_blocksmith" ("blocksmith_public_key");
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
			`CREATE INDEX "idx_id_spine_block" 
			ON "spine_block" ("id");
			`,
			`
			CREATE TABLE IF NOT EXISTS "spine_public_key"(
				"node_public_key" BLOB,
				"public_key_action" INTEGER,
				"main_block_height"	INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("node_public_key", "height")
			);`,
			`CREATE INDEX "idx_nodepubkey_spine_pubkey" 
			ON "spine_public_key" ("node_public_key");
			`,
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
			`
			CREATE TABLE IF NOT EXISTS "escrow_transaction" (
				"id" INTEGER,
				"sender_address" VARCHAR(255),
				"recipient_address" VARCHAR(255),
				"approver_address" VARCHAR(255),
				"amount" INTEGER,
				"commission" INTEGER,
				"timeout" INTEGER,
				"status" INTEGER,
				"block_height" INTEGER,
				"latest" INTEGER,
				"instruction" TEXT
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "spine_block_manifest" (
				"id" INTEGER,				-- little endian of hash of all spine_block_manifest fields but itself
				"full_file_hash" BLOB,			-- hash of the (snapshot) file content
				"file_chunk_hashes" BLOB,		-- sorted sequence file chunks hashes referenced by the spine_block_manifest
				"manifest_reference_height" INTEGER NOT NULL,	-- height at which the snapshot was taken on the (main)chain
				"chain_type" INTEGER NOT NULL,		-- chain type this spine_block_manifest reference to
				"manifest_type" INTEGER NOT NULL,	-- type of spine_block_manifest (as of now only snapshot)
				"manifest_timestamp" INTEGER NOT NULL,	-- timestamp that marks the end of file chunks processing 
				PRIMARY KEY("id")
				UNIQUE("id")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "pending_transaction" (
				"transaction_hash" BLOB,		-- transaction hash of pending transaction
				"transaction_bytes" BLOB,		-- full transaction bytes of the pending transaction
				"status" INTEGER,			-- execution status of the pending transaction
				"block_height" INTEGER,			-- height when pending transaction inserted/updated
				PRIMARY KEY("transaction_hash", "block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "pending_signature" (
				"transaction_hash" INTEGER,		-- transaction hash of pending transaction being signed
				"account_address" TEXT,			-- account address of the respective signature
				"signature" BLOB,			-- full transaction bytes of the pending transaction
				"block_height" INTEGER,			-- height when pending signature inserted/updated
				PRIMARY KEY("account_address", "transaction_hash")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "multisignature_info" (
				"multisig_address" TEXT,		-- address of multisig account / hash of multisignature_info 
				"minimum_signatures" INTEGER,		-- account address of the respective signature 
				"nonce" INTEGER,			-- full transaction bytes of the pending transaction
				"addresses" TEXT,			-- list of addresses / participants of the multisig account
				"block_height" INTEGER,			-- height when multisignature_info inserted / updated
				PRIMARY KEY("multisig_address", "block_height")
			)
			`,
			`
			CREATE INDEX "node_registry_height_idx" ON "node_registry" ("height")
			`,
			`
			CREATE INDEX "skipped_blocksmith_block_height_idx" ON "skipped_blocksmith" ("block_height")
			`,
			`
			CREATE INDEX "published_receipt_block_height_idx" ON "published_receipt" ("block_height")
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
			*m.CurrentVersion++
			err = m.Query.ExecuteTransaction(`UPDATE "migration"
				SET "version" = ?, "created_date" = datetime('now');`, m.CurrentVersion)
			if err != nil {
				return err
			}
		} else {
			m.CurrentVersion = &version // should 0 value not nil anymore
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

		err = m.Query.CommitTx()
		if err != nil {
			return err
		}
	}
	return nil
}

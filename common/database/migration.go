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
				"sender_account_address_type" INTEGER,
				"sender_account_address" BLOB,
				"recipient_account_address_type" INTEGER,
				"recipient_account_address" BLOB,
				PRIMARY KEY("id")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "transaction" (
				"id"	INTEGER,
				"block_id"	INTEGER,
				"block_height"	INTEGER,
				"sender_account_address_type" INTEGER,
				"sender_account_address" BLOB,
				"recipient_account_address_type" INTEGER,
				"recipient_account_address" BLOB,
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
				"account_address_type" INTEGER,
				"account_address" BLOB,
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
				"account_address_type" INTEGER,
				"account_address" BLOB,
				"registration_height" INTEGER,
				"locked_balance" INTEGER,
				"queued" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("id", "height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "account_dataset"(
				"setter_account_address_type" INTEGER,
				"setter_account_address" BLOB,
				"recipient_account_address_type" INTEGER,
				"recipient_account_address" BLOB,
				"property" TEXT,
				"value" TEXT,
				"is_active" INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
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
				"public_key_action" INTEGER,
				"main_block_height"	INTEGER,
				"latest" INTEGER,
				"height" INTEGER,
				PRIMARY KEY("node_public_key", "height")
			);`,
			`
			CREATE TABLE IF NOT EXISTS "account_ledger" (
				"account_address_type" INTEGER,
				"account_address" BLOB,
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
				"sender_address_type" INTEGER,
				"sender_address" BLOB,
				"recipient_address_type" INTEGER,
				"recipient_address" BLOB,
				"approver_address_type" INTEGER,
				"approver_address" BLOB,
				"amount" INTEGER,
				"commission" INTEGER,
				"timeout" INTEGER,
				"status" INTEGER,
				"block_height" INTEGER,
				"latest" INTEGER,
				"instruction" TEXT,
				PRIMARY KEY("id", "block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "spine_block_manifest" (
				"id" INTEGER,				-- little endian of hash of all spine_block_manifest fields but itself
				"full_file_hash" BLOB,			-- hash of the (snapshot) file content
				"file_chunk_hashes" BLOB,		-- sorted sequence file chunks hashes referenced by the spine_block_manifest
				"manifest_reference_height" INTEGER NOT NULL,	-- height at which the snapshot was taken on the (main)chain
				"manifest_spine_block_height" INTEGER NOT NULL,	-- height at which the snapshot was taken on the (main)chain
				"chain_type" INTEGER NOT NULL,		-- chain type this spine_block_manifest reference to
				"manifest_type" INTEGER NOT NULL,	-- type of spine_block_manifest (as of now only snapshot)
				"expiration_timestamp" INTEGER NOT NULL,	-- timestamp that marks the end of file chunks processing
				PRIMARY KEY("id")
				UNIQUE("id")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "pending_transaction" (
				"sender_address_type" INTEGER,	-- sender account type
				"sender_address" BLOB,			-- sender of transaction
				"transaction_hash" BLOB,		-- transaction hash of pending transaction
				"transaction_bytes" BLOB,		-- full transaction bytes of the pending transaction
				"status" INTEGER,			-- execution status of the pending transaction
				"block_height" INTEGER,			-- height when pending transaction inserted/updated
				"latest" INTEGER,			-- latest flag for pending transaction
				PRIMARY KEY("transaction_hash", "block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "pending_signature" (
				"transaction_hash" BLOB,		-- transaction hash of pending transaction being signed
				"account_address_type" INTEGER,	-- signing account type
				"account_address" BLOB,			-- account address of the respective signature
				"signature" BLOB,			-- full transaction bytes of the pending transaction
				"block_height" INTEGER,			-- height when pending signature inserted/updated
				"latest" INTEGER,			-- latest flag for pending signature
				PRIMARY KEY("account_address", "transaction_hash", "block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "multisignature_info" (
				"multisig_address_type" INTEGER,	-- multisig account type
				"multisig_address" BLOB,		-- address of multisig account / hash of multisignature_info
				"minimum_signatures" INTEGER,		-- account address of the respective signature
				"nonce" INTEGER,			-- full transaction bytes of the pending transaction
				"block_height" INTEGER,			-- height when multisignature_info inserted / updated
				"latest" INTEGER,			-- latest flag for pending signature
				PRIMARY KEY("multisig_address", "block_height")
			)
			`,
			`
			ALTER TABLE "transaction"
				ADD COLUMN "multisig_child" INTEGER DEFAULT 0
			`,
			`
			CREATE INDEX "node_registry_height_idx" ON "node_registry" ("height")
			`,
			`
			CREATE INDEX "node_public_key_idx" ON "node_registry" ("node_public_key")
			`,
			`
			CREATE INDEX "skipped_blocksmith_block_height_idx" ON "skipped_blocksmith" ("block_height")
			`,
			`
			CREATE INDEX "published_receipt_block_height_idx" ON "published_receipt" ("block_height")
			`,
			`
			CREATE INDEX "main_block_id_idx" ON "main_block" ("id")
			`,
			`
			CREATE INDEX "main_block_height_idx" ON "main_block" ("height")
			`,
			`
			CREATE INDEX "published_receipt_rmr_linked_idx" ON "published_receipt" ("rmr_linked")
			`,
			`
			CREATE INDEX "merkle_tree_id_idx" ON "merkle_tree" ("id")
			`,
			`
			CREATE INDEX "merkle_tree_block_height_idx" ON "merkle_tree" ("block_height")
			`,
			`
			CREATE INDEX "node_receipt_rmr_idx" ON "node_receipt" ("rmr")
			`,
			`
			CREATE INDEX "node_receipt_recipient_public_key_idx" ON "node_receipt" ("recipient_public_key")
			`,
			`
			CREATE INDEX "node_receipt_reference_block_height_idx" ON "node_receipt" ("reference_block_height")
			`,
			`
			CREATE INDEX "published_receipt_datum_hash_idx" ON "published_receipt" ("datum_hash")
			`,
			`
			CREATE INDEX "spine_block_manifest_spine_block_height_idx" ON "spine_block_manifest" ("manifest_spine_block_height")
			`,
			`
			CREATE INDEX "spine_block_manifest_reference_height_idx" ON "spine_block_manifest" ("manifest_reference_height")
			`,
			`
			CREATE INDEX "pending_transaction_transaction_hash_idx" ON "pending_transaction" ("transaction_hash")
			`,
			`
			CREATE INDEX "pending_transaction_status_idx" ON "pending_transaction" ("status")
			`,
			`
			CREATE INDEX "pending_signature_transaction_hash_idx" ON "pending_signature" ("transaction_hash")
			`,
			`
			CREATE TABLE IF NOT EXISTS "liquid_payment_transaction" (
				"id" INTEGER,
				"sender_address_type" INTEGER,	-- sender account type
				"sender_address" BLOB,			-- sender of transaction
				"recipient_address_type" INTEGER,	-- recipient account type
				"recipient_address" BLOB,			-- recipient of transaction
				"amount" INTEGER,
				"applied_time" INTEGER,
				"complete_minutes" INTEGER,
				"status" INTEGER,
				"block_height" INTEGER,
				"latest" INTEGER,
				PRIMARY KEY("id", "block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "fee_vote_commitment_vote" (
				"vote_hash" BLOB,		-- hash of fee vote object
				"voter_address_type" INTEGER,	-- sender account type of commit vote
				"voter_address" BLOB, -- sender account address of commit vote
				"block_height" INTEGER,	-- height when commit vote inserted
				PRIMARY KEY("vote_hash", "block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "fee_scale" (
				"fee_scale" INTEGER,		-- current fee scale
				"block_height" INTEGER,		-- block_height when the fee scale apply
				"latest" INTEGER,
				PRIMARY KEY("block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "fee_vote_reveal_vote" (
				"recent_block_hash" BLOB, 
				"recent_block_height" INTEGER,
				"fee_vote" INTEGER, -- fee value voted
				"voter_address_type" INTEGER,	-- sender account type as voter
				"voter_address" BLOB, -- sender account address as voter
				"voter_signature" BLOB, -- signed block_hash,block_height,fee_vote
				"block_height" INTEGER, -- height when revealed
				PRIMARY KEY("block_height", "voter_address")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "node_admission_timestamp" (
				"timestamp" INTEGER,	-- timestamp to remind the next node admission for queued node
				"block_height" INTEGER,		-- block height when the next node admission timestamp set
				"latest" INTEGER,
				PRIMARY KEY("block_height")
			)
			`,
			`
			CREATE TABLE IF NOT EXISTS "multisignature_participant" (
				"multisig_address_type" INTEGER,	-- address type of multisig account (TODO: use custom type if instead of an address we have a hash of multisig_info in multisig_address?)
				"multisig_address" BLOB, -- address of multisig account / hash of multisignature_info
				"account_address_type" INTEGER, -- account type of field below
				"account_address" BLOB, -- exists in addresses / participants of the multisig account
				"account_address_index" INTEGER, -- index / position of participants
				"latest" INTEGER,
				"block_height" INTEGER,
				PRIMARY KEY("multisig_address", "account_address", "block_height")
			)
			`,
			`
			ALTER TABLE "spine_public_key"
				ADD COLUMN "node_id" INTEGER AFTER "node_public_key"
			`,
			`CREATE TABLE IF NOT EXISTS "node_address_info" (
				"node_id"		INTEGER,					-- node_id relative to this node address
				"address"		VARCHAR(255),				-- peer/node address
				"port"			INTEGER,					-- peer rpc port
				"block_height"	INTEGER,					-- last blockchain height when broadcasting the address
				"block_hash"	BLOB,						-- hash of last block when broadcasting the address
				"signature"		BLOB,						-- signature of above fields (signed using node private key)
				"status" 		INTEGER,					-- pending or confirmed
				PRIMARY KEY("node_id","address","port")		-- primary key
			)
			`,
			`
			CREATE INDEX "node_address_info_address_idx" ON "node_address_info" ("address")
			`,
			`
			CREATE INDEX "account_balance_latest_idx" ON "account_balance" ("latest")
			`,
			`
			CREATE INDEX "account_balance_account_address_idx" ON "account_balance" ("account_address")
			`,
			`
			CREATE INDEX "account_ledger_event_type_idx" ON "account_ledger" ("event_type")
			`,
			`
			CREATE INDEX "escrow_transaction_latest_idx" ON "escrow_transaction" ("latest")
			`,
			`
			CREATE INDEX "escrow_transaction_block_height_idx" ON "escrow_transaction" ("block_height")
			`,
			`
			CREATE INDEX "transaction_block_id_idx" ON "transaction" ("block_id")
			`,
			`
			ALTER TABLE "main_block"
				ADD COLUMN "merkle_root" BLOB AFTER "payload_hash"
			`,
			`
			ALTER TABLE "main_block"
				ADD COLUMN "merkle_tree" BLOB AFTER "merkle_root"
			`,
			`
			ALTER TABLE "main_block"
				ADD COLUMN "reference_block_height" INTEGER AFTER "merkle_tree"
			`,
			`
			ALTER TABLE "spine_block"
				ADD COLUMN "merkle_root" BLOB AFTER "payload_hash"
			`,
			`
			ALTER TABLE "spine_block"
				ADD COLUMN "merkle_tree" BLOB AFTER "merkle_root"
			`,
			`
			ALTER TABLE "spine_block"
				ADD COLUMN "reference_block_height" INTEGER AFTER "merkle_tree"
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

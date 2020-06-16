package snapshot

import (
	"database/sql"
	"math/rand"
	"os"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
)

func init() {
	snapshotCmd.Flags().StringVarP(&dbPath, "db-path", "p", "resource", "Database path target")
	snapshotCmd.Flags().StringVarP(&dbName, "db-name", "n", "zoobc.db", "Database name target")
	snapshotCmd.Flags().StringVarP(&snapshotFile, "file", "f", "resource/snapshot", "Snapshot file location")
	/*
		New snapshot file
	*/
	newSnapshotCommand.Flags().Uint32VarP(&snapshotHeight, "height", "b", 0, "Block height target to snapshot")

	/*
		Storing payload
	*/
}

func Commands() *cobra.Command {
	newSnapshotCommand.Run = newSnapshotProcess()
	snapshotCmd.AddCommand(newSnapshotCommand)

	importSnapshotCommand.Run = storingPayloadProcess()
	snapshotCmd.AddCommand(importSnapshotCommand)
	return snapshotCmd
}

func newSnapshotProcess() func(ccmd *cobra.Command, args []string) {
	return func(ccmd *cobra.Command, args []string) {
		var (
			snapshotFileInfo *model.SnapshotFileInfo
			sqliteInstance   = database.NewSqliteDB()
			mainChain        = &chaintype.MainChain{}
			sqliteDB         *sql.DB
			logger           = logrus.New()
			err              error
		)

		err = sqliteInstance.InitializeDB(dbPath, dbName)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}
		sqliteDB, err = sqliteInstance.OpenDB(
			dbPath,
			dbName,
			constant.SQLMaxOpenConnetion,
			constant.SQLMaxIdleConnections,
			constant.SQLMaxConnectionLifetime,
		)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}

		fileService := service.NewFileService(
			logger,
			new(codec.CborHandle),
			snapshotFile,
		)
		executor := query.NewQueryExecutor(sqliteDB)
		snapshotMainService := service.NewSnapshotMainBlockService(
			snapshotFile,
			executor,
			logger,
			service.NewSnapshotBasicChunkStrategy(
				constant.SnapshotChunkSize,
				fileService,
			),
			query.NewAccountBalanceQuery(),
			query.NewNodeRegistrationQuery(),
			query.NewParticipationScoreQuery(),
			query.NewAccountDatasetsQuery(),
			query.NewEscrowTransactionQuery(),
			query.NewPublishedReceiptQuery(),
			query.NewPendingTransactionQuery(),
			query.NewPendingSignatureQuery(),
			query.NewMultisignatureInfoQuery(),
			query.NewSkippedBlocksmithQuery(),
			query.NewBlockQuery(mainChain),
			query.GetSnapshotQuery(mainChain),
			query.GetBlocksmithSafeQuery(mainChain),
			query.GetDerivedQuery(mainChain),
			&transaction.Util{},
			&transaction.TypeSwitcher{Executor: executor},
		)
		snapshotService := service.NewSnapshotService(
			service.NewSpineBlockManifestService(
				executor,
				query.NewSpineBlockManifestQuery(),
				query.NewBlockQuery(&chaintype.SpineChain{}),
				logger,
			),
			service.NewBlockchainStatusService(true, logger),
			map[int32]service.SnapshotBlockServiceInterface{
				(&chaintype.MainChain{}).GetTypeInt(): snapshotMainService,
			},
			logger,
		)
		snapshotFileInfo, err = snapshotService.GenerateSnapshot(&model.Block{
			ID:     rand.Int63n(int64(snapshotHeight)),
			Height: snapshotHeight,
		},
			mainChain,
			constant.SnapshotChunkSize,
		)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}

		_, err = snapshotService.SpineBlockManifestService.CreateSpineBlockManifest(
			snapshotFileInfo.SnapshotFileHash,
			snapshotFileInfo.Height,
			snapshotFileInfo.ProcessExpirationTimestamp,
			snapshotFileInfo.FileChunksHashes,
			&chaintype.MainChain{},
			model.SpineBlockManifestType_Snapshot,
		)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}
	}
}

func storingPayloadProcess() func(ccmd *cobra.Command, args []string) {
	return func(ccmd *cobra.Command, args []string) {
		var (
			snapshotFileInfo   *model.SnapshotFileInfo
			sqliteInstance     = database.NewSqliteDB()
			mainChain          = &chaintype.MainChain{}
			spineBlockManifest *model.SpineBlockManifest
			sqliteDB           *sql.DB
			logger             = logrus.New()
			err                error
		)

		err = sqliteInstance.InitializeDB(dbPath, dbName)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}
		sqliteDB, err = sqliteInstance.OpenDB(
			dbPath,
			dbName,
			constant.SQLMaxOpenConnetion,
			constant.SQLMaxIdleConnections,
			constant.SQLMaxConnectionLifetime,
		)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}

		fileService := service.NewFileService(
			logger,
			new(codec.CborHandle),
			snapshotFile,
		)
		executor := query.NewQueryExecutor(sqliteDB)
		snapshotMainService := service.NewSnapshotMainBlockService(
			snapshotFile,
			executor,
			logger,
			service.NewSnapshotBasicChunkStrategy(
				constant.SnapshotChunkSize,
				fileService,
			),
			query.NewAccountBalanceQuery(),
			query.NewNodeRegistrationQuery(),
			query.NewParticipationScoreQuery(),
			query.NewAccountDatasetsQuery(),
			query.NewEscrowTransactionQuery(),
			query.NewPublishedReceiptQuery(),
			query.NewPendingTransactionQuery(),
			query.NewPendingSignatureQuery(),
			query.NewMultisignatureInfoQuery(),
			query.NewSkippedBlocksmithQuery(),
			query.NewBlockQuery(mainChain),
			query.GetSnapshotQuery(mainChain),
			query.GetBlocksmithSafeQuery(mainChain),
			query.GetDerivedQuery(mainChain),
			&transaction.Util{},
			&transaction.TypeSwitcher{Executor: executor},
		)

		spineBlockManifestService := service.NewSpineBlockManifestService(
			executor,
			query.NewSpineBlockManifestQuery(),
			query.NewBlockQuery(&chaintype.SpineChain{}),
			logger,
		)
		spineBlockManifest, err = spineBlockManifestService.GetLastSpineBlockManifest(mainChain, model.SpineBlockManifestType_Snapshot)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)

		}

		fileChunkHashes, err := fileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), sha3.New256().Size())
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}

		snapshotFileInfo = &model.SnapshotFileInfo{
			SnapshotFileHash:           spineBlockManifest.GetFullFileHash(),
			FileChunksHashes:           fileChunkHashes,
			ChainType:                  mainChain.GetTypeInt(),
			Height:                     spineBlockManifest.ManifestReferenceHeight,
			ProcessExpirationTimestamp: spineBlockManifest.ExpirationTimestamp,
			SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
		}
		err = snapshotMainService.ImportSnapshotFile(snapshotFileInfo)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}
	}
}

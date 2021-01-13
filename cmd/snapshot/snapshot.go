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
package snapshot

import (
	"crypto/sha256"
	"database/sql"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/queue"
	"math/rand"
	"os"

	"github.com/zoobc/zoobc-core/common/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"golang.org/x/crypto/sha3"
)

func init() {
	snapshotCmd.PersistentFlags().StringVarP(&dbPath, "db-path", "p", "resource", "Database path target")
	snapshotCmd.PersistentFlags().StringVarP(&dbName, "db-name", "n", "zoobc.db", "Database name target")
	snapshotCmd.PersistentFlags().StringVarP(&snapshotFile, "file", "f", "resource/snapshot", "Snapshot file location")
	snapshotCmd.PersistentFlags().BoolVarP(&dump, "dump", "d", true, "Dump result out")
	/*
		New snapshot file
	*/
	newSnapshotCommand.Flags().Uint32VarP(&snapshotHeight, "height", "b", 0, "Block height target to snapshot")

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
			signature        = crypto.NewSignature()
			snapshotFileInfo *model.SnapshotFileInfo
			sqliteInstance   = database.NewSqliteDB()
			snapshotService  *service.SnapshotService
			mainChain        = &chaintype.MainChain{}
			executor         *query.Executor
			sqliteDB         *sql.DB
			logger           = logrus.New()
			err              error
		)

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
		executor = query.NewQueryExecutor(sqliteDB, queue.NewPriorityPreferenceLock())
		mempoolStorage := storage.NewMempoolStorage()
		nodeAuthValidation := auth.NewNodeAuthValidation(signature)
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
			query.NewMultiSignatureParticipantQuery(),
			query.NewSkippedBlocksmithQuery(&chaintype.MainChain{}),
			query.NewFeeScaleQuery(),
			query.NewFeeVoteCommitmentVoteQuery(),
			query.NewFeeVoteRevealVoteQuery(),
			query.NewLiquidPaymentTransactionQuery(),
			query.NewNodeAdmissionTimestampQuery(),
			query.NewBlockQuery(mainChain),
			query.GetSnapshotQuery(mainChain),
			query.GetBlocksmithSafeQuery(mainChain),
			query.GetDerivedQuery(mainChain),
			&transaction.Util{},
			&transaction.TypeSwitcher{
				Executor:            executor,
				MempoolCacheStorage: mempoolStorage,
				NodeAuthValidation:  nodeAuthValidation,
			},
			nil,
			nil,
			nil,
		)
		nodeShardStorage := storage.NewNodeShardCacheStorage()
		snapshotChunkUtil := util.NewChunkUtil(sha256.Size, nodeShardStorage, logger)

		spinePublicKeyService := service.NewBlockSpinePublicKeyService(
			crypto.NewSignature(),
			executor,
			query.NewNodeRegistrationQuery(),
			query.NewSpinePublicKeyQuery(),
			logger,
		)
		snapshotService = service.NewSnapshotService(
			service.NewSpineBlockManifestService(
				executor,
				query.NewSpineBlockManifestQuery(),
				query.NewBlockQuery(&chaintype.SpineChain{}),
				logger,
			),
			spinePublicKeyService,
			service.NewBlockchainStatusService(true, logger),
			map[int32]service.SnapshotBlockServiceInterface{
				(&chaintype.MainChain{}).GetTypeInt(): snapshotMainService,
			},
			snapshotChunkUtil,
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

		if dump {
			_ = sqliteInstance.CloseDB()

			dbPath = snapshotFile
			dbName = "dump.db"
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
			executor = query.NewQueryExecutor(sqliteDB, queue.NewPriorityPreferenceLock())
			migration := database.Migration{
				Query: executor,
			}
			err = migration.Init()
			if err != nil {
				logger.Errorf("Snapshot Failed: %s", err.Error())
				os.Exit(0)
			}

			err = migration.Apply()
			if err != nil {
				logger.Errorf("Snapshot Failed: %s", err.Error())
				os.Exit(0)
			}

			snapshotService = service.NewSnapshotService(
				service.NewSpineBlockManifestService(
					executor,
					query.NewSpineBlockManifestQuery(),
					query.NewBlockQuery(&chaintype.SpineChain{}),
					logger,
				),
				spinePublicKeyService,
				service.NewBlockchainStatusService(true, logger),
				map[int32]service.SnapshotBlockServiceInterface{
					(&chaintype.MainChain{}).GetTypeInt(): snapshotMainService,
				},
				snapshotChunkUtil,
				logger,
			)
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
			signature                 = crypto.NewSignature()
			nodeAuthValidationService = auth.NewNodeAuthValidation(signature)
			mempoolStorage            = storage.NewMempoolStorage()
			snapshotFileInfo          *model.SnapshotFileInfo
			sqliteInstance            = database.NewSqliteDB()
			mainChain                 = &chaintype.MainChain{}
			spineBlockManifest        *model.SpineBlockManifest
			sqliteDB                  *sql.DB
			executor                  *query.Executor
			logger                    = logrus.New()
			err                       error
		)

		if dump {
			dbPath = snapshotFile
			dbName = "dump.db"
		}
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
		executor = query.NewQueryExecutor(sqliteDB, queue.NewPriorityPreferenceLock())
		typeSwitcher := &transaction.TypeSwitcher{
			Executor:            executor,
			NodeAuthValidation:  nodeAuthValidationService,
			MempoolCacheStorage: mempoolStorage,
		}
		mainBlockService := service.NewBlockMainService(
			mainChain,
			executor,
			query.NewBlockQuery(mainChain),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			typeSwitcher,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			storage.NewBlockStateStorage(),
			nil,
			nil,
			nil,
		)
		err = mainBlockService.UpdateLastBlockCache(nil)
		if err != nil {
			logger.Errorf("Snapshot Failed: %s", err.Error())
			os.Exit(0)
		}
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
			query.NewMultiSignatureParticipantQuery(),
			query.NewSkippedBlocksmithQuery(&chaintype.MainChain{}),
			query.NewFeeScaleQuery(),
			query.NewFeeVoteCommitmentVoteQuery(),
			query.NewFeeVoteRevealVoteQuery(),
			query.NewLiquidPaymentTransactionQuery(),
			query.NewNodeAdmissionTimestampQuery(),
			query.NewBlockQuery(mainChain),
			query.GetSnapshotQuery(mainChain),
			query.GetBlocksmithSafeQuery(mainChain),
			query.GetDerivedQuery(mainChain),
			&transaction.Util{},
			typeSwitcher,
			mainBlockService,
			nil,
			nil,
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

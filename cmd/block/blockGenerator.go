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
package block

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/queue"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/core/smith"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	mockBlockchainStatusService struct {
		service.BlockchainStatusService
	}
)

var (
	blocksmith              *model.Blocksmith
	chainType               chaintype.ChainType
	blockProcessor          smith.BlockchainProcessorInterface
	blockService            service.BlockServiceInterface
	nodeRegistrationService service.NodeRegistrationServiceInterface
	blockSmithStrategy      strategy.BlocksmithStrategyInterface
	blocksmithStrategy      strategy.BlocksmithStrategyInterface
	queryExecutor           query.ExecutorInterface
	migration               database.Migration

	numberOfBlocks         int
	blocksmithSecretPhrase string
	outputPath             string

	blockCmd = &cobra.Command{
		Use:   "block",
		Short: "block command used to manipulate block of node",
		Long: `
			block command is use to manipulate block creation or broadcasting in the node
		`,
	}

	fakeBlockCmd = &cobra.Command{
		Use:   "fake-blocks",
		Short: "fake-blocks command used to create fake blocks",
		Run: func(cmd *cobra.Command, args []string) {
			generateBlocks(numberOfBlocks, blocksmithSecretPhrase, outputPath, chainType)
		},
	}
)

func (*mockBlockchainStatusService) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return true
}

func (*mockBlockchainStatusService) IsDownloading(ct chaintype.ChainType) bool {
	return true
}

func init() {
	fakeBlockCmd.Flags().IntVar(
		&numberOfBlocks,
		"numberOfBlocks",
		100,
		"number of account to generate",
	)
	fakeBlockCmd.Flags().StringVar(
		&blocksmithSecretPhrase,
		"blocksmithSecretPhrase",
		"",
		"secret phrase of blocksmith | required",
	)
	fakeBlockCmd.Flags().StringVar(
		&outputPath,
		"out",
		"./testdata/zoobc.db",
		"output path of the database",
	)
	blockCmd.AddCommand(fakeBlockCmd)
}

func Commands() *cobra.Command {
	return blockCmd
}

func initialize(
	secretPhrase, outputPath string,
) {
	signature := crypto.NewSignature()
	nodeAuthValidation := auth.NewNodeAuthValidation(signature)
	transactionUtil := &transaction.Util{}
	receiptUtil := &coreUtil.ReceiptUtil{}
	paths := strings.Split(outputPath, "/")
	dbPath, dbName := strings.Join(paths[:len(paths)-1], "/")+"/", paths[len(paths)-1]
	chainType = &chaintype.MainChain{}
	observerInstance := observer.NewObserver()
	blocksmith = model.NewBlocksmith(secretPhrase, signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase), 0)
	// initialize/open db and queryExecutor
	dbInstance := database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		panic(err)
	}
	db, err := dbInstance.OpenDB(dbPath, dbName, 10, 10, 20*time.Minute)
	if err != nil {
		panic(err)
	}
	queryExecutor = query.NewQueryExecutor(db, queue.NewPriorityPreferenceLock())
	mempoolStorage := storage.NewMempoolStorage()

	actionSwitcher := &transaction.TypeSwitcher{
		Executor:            queryExecutor,
		NodeAuthValidation:  nodeAuthValidation,
		MempoolCacheStorage: mempoolStorage,
	}
	blockStorage := storage.NewBlockStateStorage()
	blocksStorage := storage.NewBlocksStorage(monitoring.TypeMainBlocksCacheStorage)
	nodeAddressInfoStorage := storage.NewNodeAddressInfoStorage()
	receiptService := service.NewReceiptService(
		query.NewBatchReceiptQuery(),
		query.NewMerkleTreeQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewBlockQuery(chainType),
		queryExecutor,
		nodeRegistrationService,
		crypto.NewSignature(),
		nil,
		receiptUtil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	mempoolService := service.NewMempoolService(
		transactionUtil,
		chainType,
		queryExecutor,
		query.NewMempoolQuery(chainType),
		query.NewMerkleTreeQuery(),
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewTransactionQuery(chainType),
		crypto.NewSignature(),
		observerInstance,
		log.New(),
		receiptUtil,
		receiptService,
		nil,
		blocksStorage,
		mempoolStorage,
		nil,
	)
	nodeAddressInfoService := service.NewNodeAddressInfoService(
		queryExecutor,
		query.NewNodeAddressInfoQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewBlockQuery(chainType),
		nil,
		nodeAddressInfoStorage,
		nil,
		nil,
		nil,
		log.New(),
	)
	activeNodeRegistryCacheStorage := storage.NewNodeRegistryCacheStorage(monitoring.TypeActiveNodeRegistryStorage, nil)
	pendingNodeRegistryCacheStorage := storage.NewNodeRegistryCacheStorage(monitoring.TypePendingNodeRegistryStorage, nil)
	nodeRegistrationService := service.NewNodeRegistrationService(
		queryExecutor,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
		query.NewNodeAdmissionTimestampQuery(),
		log.New(),
		&mockBlockchainStatusService{},
		nodeAddressInfoService,
		nil,
		activeNodeRegistryCacheStorage,
		pendingNodeRegistryCacheStorage,
	)

	blocksmithStrategy = strategy.NewBlocksmithStrategyMain(
		log.New(),
		nil,
		activeNodeRegistryCacheStorage,
		query.NewSkippedBlocksmithQuery(&chaintype.MainChain{}),
		query.NewBlockQuery(&chaintype.MainChain{}),
		nil,
		queryExecutor,
		crypto.NewRandomNumberGenerator(),
		nil,
	)
	publishedReceiptUtil := coreUtil.NewPublishedReceiptUtil(
		query.NewPublishedReceiptQuery(),
		queryExecutor,
	)
	feeScaleService := fee.NewFeeScaleService(
		query.NewFeeScaleQuery(),
		blockStorage,
		queryExecutor,
	)
	blockService = service.NewBlockMainService(
		chainType,
		queryExecutor,
		query.NewBlockQuery(chainType),
		query.NewMempoolQuery(chainType),
		query.NewTransactionQuery(chainType),
		query.NewSkippedBlocksmithQuery(&chaintype.MainChain{}),
		crypto.NewSignature(),
		mempoolService,
		receiptService,
		nodeRegistrationService,
		nodeAddressInfoService,
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewParticipationScoreQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewFeeVoteRevealVoteQuery(),
		observerInstance,
		blocksmithStrategy,
		log.New(),
		query.NewAccountLedgerQuery(),
		service.NewBlockIncompleteQueueService(chainType, observerInstance),
		transactionUtil,
		receiptUtil,
		publishedReceiptUtil,
		service.NewTransactionCoreService(
			nil,
			queryExecutor,
			nil,
			nil,
			query.NewTransactionQuery(chainType),
			nil,
			nil,
		), nil, nil, nil, nil, nil, nil, feeScaleService, query.GetPruneQuery(chainType), nil, nil, nil, nil)

	migration = database.Migration{Query: queryExecutor}
}

// generateBlocks used to generate dummy block for testing
// note: now only support mainchain, will implement spinechain implementation details later.
func generateBlocks(numberOfBlocks int, blocksmithSecretPhrase, outputPath string, ct chaintype.ChainType) {
	fmt.Println("initializing dependency and database...")
	initialize(blocksmithSecretPhrase, outputPath)
	fmt.Println("done initializing database")
	blockProcessor = smith.NewBlockchainProcessor(
		blockService.GetChainType(),
		blocksmith,
		blockService,
		log.New(),
		&mockBlockchainStatusService{},
		nodeRegistrationService,
		blockSmithStrategy,
	)
	startTime := time.Now().UnixNano() / 1e6
	fmt.Printf("generating %d blocks\n", numberOfBlocks)
	fmt.Println("initializing database schema migration")
	if err := migration.Init(); err != nil {
		panic(err)
	}

	fmt.Println("applying database schema migration")
	if err := migration.Apply(); err != nil {
		panic(err)
	}
	fmt.Println("checking genesis...")

	if exist, _ := blockService.CheckGenesis(); !exist { // Add genesis if not exist
		fmt.Println("genesis does not exist, adding genesis")
		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			panic(err)
		}

		if err := blockService.AddGenesis(); err != nil {
			panic(err)
		}

		fmt.Println("begin generating blocks")
		if err := blockProcessor.FakeSmithing(numberOfBlocks, true, ct); err != nil {
			panic(err)
		}
	} else {
		// start from last block's timestamp
		fmt.Println("continuing from last database...")
		if err := blockProcessor.FakeSmithing(numberOfBlocks, false, ct); err != nil {
			panic("error in appending block to existing database")
		}
	}
	fmt.Printf("database generated in %s", outputPath)
	fmt.Printf("block generation success in %d miliseconds", (time.Now().UnixNano()/1e6)-startTime)
}

package block

import (
	"fmt"
	"log"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/common/database"

	"github.com/zoobc/zoobc-core/core/smith"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func GenerateBlocks(logger *logrus.Logger,
	blockProcessor smith.BlockchainProcessorInterface,
	blockService service.BlockServiceInterface,
	queryExecutor query.ExecutorInterface,
	migration database.Migration,
) *cobra.Command {
	var (
		numberOfBlocks int
	)
	var blockCmd = &cobra.Command{
		Use:   "block",
		Short: "block command used to manipulate block of node",
		Long: `
			block command is use to manipulate block creation or broadcasting in the node
		`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "generate-fake" {
				fmt.Printf("number of blocks: %d\n\n", numberOfBlocks)
				if err := migration.Init(); err != nil {
					log.Fatal(err)
				}

				if err := migration.Apply(); err != nil {
					log.Fatal(err)
				}
				if !blockService.CheckGenesis() { // Add genesis if not exist
					// genesis account will be inserted in the very beginning
					if err := service.AddGenesisAccount(queryExecutor); err != nil {
						log.Fatal("Fail to add genesis account")
					}

					if err := blockService.AddGenesis(); err != nil {
						log.Fatalf("error in adding genesis: %v", err)
					}

					if err := blockProcessor.FakeSmithing(numberOfBlocks); err != nil {
						log.Fatalf("error in fake smithing: %v", err)
					}
				} else {
					log.Fatal("previous generated database still exist, move them")
				}
				log.Printf("database generated")
			} else {
				logger.Error("unknown command")
			}
		},
	}
	blockCmd.Flags().IntVar(&numberOfBlocks, "numberOfBlocks", 100, "number of account to generate")
	return blockCmd
}

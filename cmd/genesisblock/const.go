package genesisblock

import "github.com/spf13/cobra"

var (
	genesisCmd = &cobra.Command{
		Use:   "genesis",
		Short: "command used to generate a new genesis block.",
	}

	/*
		// for genesis generate command
	*/
	withDbLastState bool
	dbPath          string
	extraNodesCount int

	/*
		// for genesis generate-consul-kv command
	*/
	logLevels              string
	wellKnownPeers         string
	deploymentName         string
	kvFileCustomConfigFile string
)

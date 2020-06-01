package genesisblock

import "github.com/spf13/cobra"

const (
	envTargetDevelop envTargetType = 0
	envTargetStaging envTargetType = 1
	envTargetAlpha   envTargetType = 2
	envTargetLocal   envTargetType = 3
)

type (
	/*
		ENV Target enum
	*/
	envTargetType uint32
)

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

	/*
		ENV Target
	*/
	envTarget       string
	output          string
	envTarget_value = map[string]uint32{
		"develop": 0,
		"staging": 1,
		"alpha":   2,
		"local":   3,
	}
)

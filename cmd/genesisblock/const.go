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
	withDbLastState                                 bool
	dbPath, applicationCodeName, applicationVersion string
	extraNodesCount                                 int
	genesisTimestamp                                int

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
	envTarget      string
	output         string
	envTargetValue = map[string]uint32{
		"develop":      0,
		"staging":      1,
		"alpha":        2,
		"local":        3,
		"experimental": 4,
		"beta":         5,
	}
)

type (
	genesisEntry struct {
		AccountAddress     string
		AccountSeed        string
		AccountBalance     int64
		NodeSeed           string
		NodePublicKey      string
		NodePublicKeyBytes []byte
		LockedBalance      int64
		ParticipationScore int64
		Smithing           bool
	}
	clusterConfigEntry struct {
		NodePublicKey  string `json:"NodePublicKey"`
		NodeSeed       string `json:"NodeSeed"`
		AccountAddress string `json:"AccountAddress"`
		NodeAddress    string `json:"MyAddress,omitempty"`
		Smithing       bool   `json:"Smithing,omitempty"`
	}
	accountNodeEntry struct {
		NodePublicKey  string
		AccountAddress string
	}
	parseErrorLog struct {
		AccountAddress    string `json:"AccountAddress"`
		ConfigPublicKey   string `json:ConfigPublicKey`
		ComputedPublicKey string `json:ComputedPublicKey`
	}
)

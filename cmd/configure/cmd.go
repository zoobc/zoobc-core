package configure

import (
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/admin"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "configure command to checking and generate configuration file",
		Long:  "configure command to checking, validate and generate configuration file",
	}
)

func init() {
	configureCmd.Flags().StringVarP(&target, "target", "t", "dev", "target configuration dev | alpha | beta")
}
func Commands() *cobra.Command {
	configureCmd.Run = GenerateConfigFileCommand
	return configureCmd
}
func GenerateConfigFileCommand(*cobra.Command, []string) {
	var (
		err        error
		configFile = "config.toml"
		shell      = ishell.New()
		config     model.Config
	)

	_, err = os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			color.Cyan("Will generate config file with prompt")
			_ = generateConfig(config)
		}
	} else {
		update := shell.MultiChoice([]string{"Yes", "No"}, "Config file exists, want to update ? [ENTER to exit]")
		if update == 0 {
			err = util.LoadConfig("./", "config", "toml")
			if err != nil {
				color.Red(err.Error())
				return
			}
			shell.Close()
			config.LoadConfigurations()
			_ = generateConfig(config)
		}
	}

}

func generateConfig(config model.Config) error {
	var (
		shell    = ishell.New()
		port     int
		err      error
		inputStr string
		alpha    = []string{
			"n0.alpha.proofofparticipation.network:8001",
			"n1.alpha.proofofparticipation.network:8001",
			"n2.alpha.proofofparticipation.network:8001",
			"172.105.37.61:8001",
			"80.85.84.163:8001",
		}
		dev = []string{
			"172.104.34.10:8001",
			"45.79.39.58:8001",
			"85.90.246.90:8001",
		}
	)

	// SET DEFAULT
	config.MonitoringPort = 9090
	config.CPUProfilingPort = 6060
	config.MaxAPIRequestPerSecond = 10
	config.Smithing = true
	config.ResourcePath = "./resource"
	config.DatabaseFileName = "zoobc.db"
	config.BadgerDbName = "zoobc_kv"
	config.CliMonitoring = true
	config.SnapshotPath = "./resource/snapshots"

	// WELL KNOWN PEERS
	switch target {
	case "alpha":
		config.WellknownPeers = alpha
	case "dev":
		config.WellknownPeers = dev
	default:

	}
	// PEER PORT
	color.White("PEER PORT [default 8001], Enter for default value: ")
	inputStr = shell.ReadLine()
	if strings.TrimSpace(inputStr) != "" {
		port, err = strconv.Atoi(inputStr)
		if err != nil {
			return err
		}
		config.PeerPort = uint32(port)
	} else {
		config.PeerPort = 8001
	}

	// API RPC PORT
	color.White("API RPC PORT [default 7000], Enter for default value: ")
	inputStr = shell.ReadLine()
	if strings.TrimSpace(inputStr) != "" {
		config.RPCAPIPort, err = strconv.Atoi(inputStr)
		if err != nil {
			return err
		}
	} else {
		config.RPCAPIPort = 7000
	}

	// API HTTP PORT
	color.White("API HTTP PORT [default 7001], Enter for default value: ")
	inputStr = shell.ReadLine()
	if strings.TrimSpace(inputStr) != "" {
		config.HTTPAPIPort, err = strconv.Atoi(inputStr)
		if err != nil {
			return err
		}
	} else {
		config.HTTPAPIPort = 7001
	}

	// OWNER ACCOUNT ADDRESS
	color.Cyan("! Create one on zoobc.one")
	color.White("OWNER ACCOUNT ADDRESS: ")
	inputStr = shell.ReadLine()
	if strings.TrimSpace(inputStr) != "" {
		config.OwnerAccountAddress = inputStr
	} else {
		if config.OwnerAccountAddress != "" {
			color.Cyan("previous ownerAccountAddress won't be replaced")
		} else {
			color.Yellow("! Node won't running when owner account address is empty.")
		}
	}

	color.White("NODE SEED: [Enter to let us generate a random for you]")
	inputStr = shell.ReadLine()
	if strings.TrimSpace(inputStr) != "" {
		config.NodeSeed = inputStr
	} else {
		admin.GenerateNodeKeysFile(nil, nil)
	}

	color.Yellow("! Please don't do anything, configuration will save")
	err = config.SaveConfig()
	if err != nil {
		color.Red(err.Error())
	}
	color.Cyan("Configuration saved")
	return nil
}

package configure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/admin"
	"github.com/zoobc/zoobc-core/cmd/helper"
	"github.com/zoobc/zoobc-core/common/crypto"
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
	configureCmd.Run = generateConfigFileCommand
	return configureCmd
}
func generateConfigFileCommand(*cobra.Command, []string) {
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
			err = generateConfig(config)
			if err == nil {
				color.Cyan("Configuration saved")
			}

		}
	} else { // Try another step
		update := shell.MultiChoice([]string{"Yes", "No"}, "Config file exists, want to update ? [ENTER to exit]")
		if update == 0 {
			err = util.LoadConfig("./", "config", "toml")
			if err != nil {
				color.Red(err.Error())
				return
			}
			shell.Close()
			config.LoadConfigurations()
			err = generateConfig(config)
			if err != nil {
				color.Red(err.Error())
			} else {
				color.Cyan("Configuration saved")
			}
		}
	}

}

func readCertFile(config *model.Config, fileName string) error {
	var (
		inputStr            string
		shell               = ishell.New()
		certFile, err       = os.Open(path.Join(helper.GetAbsDBPath(), fileName))
		readBuff, certBytes []byte
	)

	if err != nil {
		return fmt.Errorf("a wallet certificate has been found, failed to open it, %s", err.Error())
	}
	defer certFile.Close()

	readBuff, err = ioutil.ReadAll(certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %s", err.Error())
	}
	var certMap map[string]interface{}
	for i := 0; i <= 3; i++ {
		if i > 3 {
			return fmt.Errorf("maximum numbers of attempts exceeded")
		}
		color.Cyan("! A wallet certificate has been found. Enter the password to decrypt and import from it: ")
		inputStr = shell.ReadPassword()
		certBytes, err = crypto.OpenSSLDecrypt(inputStr, string(readBuff))
		if err != nil {
			color.Red("Attempt n. %d decrypting certificate failed", i)
			continue
		} else {
			err = json.Unmarshal(certBytes, &certMap)
			if err != nil {
				return fmt.Errorf("failed to assert certificate, %s", err.Error())
			}
			if ownerAccountAddress, ok := certMap["ownerAccount"]; ok {
				config.OwnerAccountAddress = fmt.Sprintf("%s", ownerAccountAddress)
			} else {
				return fmt.Errorf("invalid certificate format, ownerAccount not found")
			}
			if nodeSeed, ok := certMap["nodeKey"]; ok {
				config.NodeSeed = fmt.Sprintf("%s", nodeSeed)
			} else {
				return fmt.Errorf("invalid certificate format, nodeSeed not found")
			}
			break
		}
	}
	return nil
}
func generateConfig(config model.Config) error {
	var (
		shell    = ishell.New()
		port     int
		err      error
		inputStr string
		beta     = []string{
			"172.104.117.98:8002",
			"172.105.185.12:8002",
			"45.79.145.167:8002",
			"172.105.18.138:8002",
			"45.79.218.142:8002",
			"198.58.111.41:8002",
			"96.126.100.16:8002",
			"45.79.167.148:8002",
			"172.105.166.14:8002",
			"173.255.248.8:8002",
			"172.105.149.84:8002",
			"139.162.71.117:8002",
			"176.58.111.94:8002",
			"173.255.202.86:8002",
			"139.162.214.77:8002",
			"139.162.27.172:8002",
		}
		alpha = []string{
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
		config.WellknownPeers = beta

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

	// OWNER ACCOUNT ADDRESS & NODE SEED
	if config.WalletCertFileName != "" {
		if _, err = os.Stat(path.Join(helper.GetAbsDBPath(), config.WalletCertFileName)); err == nil {
			err = readCertFile(&config, config.WalletCertFileName)
			if err != nil {
				color.Red(err.Error())
			}
		}
	} else {
		haveCertFile := shell.MultiChoice(
			[]string{"Yes, already put in root app.", "No, want to manually input the owner address and node seed."},
			"Have wallet certificate ?",
		)
		switch haveCertFile {
		case 0:
			color.White("Wallet certificate name: ")
			inputStr = shell.ReadLine()
			if _, err = os.Stat(path.Join(helper.GetAbsDBPath(), inputStr)); err != nil {
				color.Red("%s not found on the root app. Please input manual", inputStr)
				break
			}
			err = readCertFile(&config, inputStr)
			if err != nil {
				color.Red(err.Error())
				return err
			}
			config.WalletCertFileName = inputStr
		default:
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
			}
		}
	}

	color.Yellow("! Please don't do anything, configuration will save")
	admin.GenerateNodeKeysFile(config.NodeSeed)
	err = config.SaveConfig()
	if err != nil {
		color.Red(err.Error())
	}
	return nil
}

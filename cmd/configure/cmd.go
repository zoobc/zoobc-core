package configure

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/cmd/admin"
	"github.com/zoobc/zoobc-core/cmd/helper"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"gopkg.in/abiosoft/ishell.v2"
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
		err                   error
		configFile            = "config.toml"
		shell                 = ishell.New()
		config                model.Config
		encodedAccountAddress string
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
			err = util.LoadConfig("./", "config", "toml", "")
			if err != nil {
				color.Red(err.Error())
				return
			}
			shell.Close()
			config.LoadConfigurations()
			// decode owner account address
			if config.OwnerAccountAddressHex != "" {
				config.OwnerAccountAddress, err = hex.DecodeString(config.OwnerAccountAddressHex)
				if err != nil {
					panic(err)
				}
				// double check that the decoded account address is valid
				accType, err := accounttype.NewAccountTypeFromAccount(config.OwnerAccountAddress)
				if err != nil {
					panic(err)
				}
				// TODO: move to crypto package in a function
				switch accType.GetTypeInt() {
				case 0:
					ed25519 := crypto.NewEd25519Signature()
					encodedAccountAddress, err = ed25519.GetAddressFromPublicKey(accType.GetAccountPrefix(), accType.GetAccountPublicKey())
					if err != nil {
						panic(err)
					}
				case 1:
					bitcoinSignature := crypto.NewBitcoinSignature(crypto.DefaultBitcoinNetworkParams(), crypto.DefaultBitcoinCurve())
					encodedAccountAddress, err = bitcoinSignature.GetAddressFromPublicKey(accType.GetAccountPublicKey())
					if err != nil {
						panic(err)
					}
				default:
					panic("Invalid Owner Account Type")
				}
				config.OwnerEncodedAccountAddress = encodedAccountAddress
				config.OwnerAccountAddressTypeInt = accType.GetTypeInt()
			}

			err = generateConfig(config)
			if err != nil {
				color.Red(err.Error())
			} else {
				color.Cyan("Configuration saved")
			}
		}
	}

}

// readCertFile read the certificate file *.zbc encrypt and extract the value and also verify the values
func readCertFile(config *model.Config, fileName string) error {
	var (
		readBuff, certBytes []byte
		inputStr            string
		certFile, err       = os.Open(fileName)
		certMap             map[string]interface{}
		shell               = ishell.New()
	)

	if err != nil {
		// there is not certificate file and need to input the base64 version
		color.Cyan("CERTIFICATE BASE64: ")
		for i := 0; i <= maxAttemptPromptFailed; i++ {
			if i >= maxAttemptPromptFailed {
				return fmt.Errorf("maximum numbers of attempts exceeded")
			}
			color.White("Input multiple lines and end with semicolon ';'.")
			inputStr = shell.ReadMultiLinesFunc(func(s string) bool {
				if ok := strings.HasSuffix(s, ";"); ok {
					return false
				}
				return true
			})
			if strings.TrimSpace(inputStr) == "" {
				color.Red("Attempt n. %d bad input value", i)
				continue
			}
			inputStr = strings.Trim(inputStr, ";")
			if inputStr == "" {
				return fmt.Errorf("certificate base64 empty")
			}
			readBuff = bytes.NewBufferString(inputStr).Bytes()
			break
		}
	} else {
		// read from certificate file
		defer certFile.Close()
		readBuff, err = ioutil.ReadAll(certFile)
		if err != nil {
			return fmt.Errorf("failed to read certificate file: %s", err.Error())
		}
	}

	color.Cyan("CERTIFICATE PASSWORD: ")
	for i := 0; i <= maxAttemptPromptFailed; i++ {
		if i >= maxAttemptPromptFailed {
			return fmt.Errorf("maximum numbers of attempts exceeded")
		}
		inputStr = shell.ReadPassword()
		certBytes, err = crypto.OpenSSLDecrypt(inputStr, string(readBuff))
		if err != nil {
			color.Red("Attempt n. %d decrypting certificate failed", i)
			continue
		}

		err = json.Unmarshal(certBytes, &certMap)
		if err != nil {
			return fmt.Errorf("failed to assert certificate, %s", err.Error())
		}
		if ownerAccountAddress, ok := certMap["ownerAccount"]; ok {
			// TODO: if in future we accept any account type in genesis, certificate must contain a full account address in hex format.
			//  As of now we have to parse an encoded address into its full hex format to be put in config.toml
			config.OwnerAccountAddress, err = accounttype.ParseEncodedAccountToAccountAddress(
				int32(model.AccountType_ZbcAccountType),
				fmt.Sprintf("%s", ownerAccountAddress),
			)
			if err != nil {
				color.Cyan("Invalid Owner Account Address: It must be a valid ZBC account")
				return err
			}
			config.OwnerAccountAddressHex = hex.EncodeToString(config.OwnerAccountAddress)
		} else {
			return fmt.Errorf("invalid certificate format, ownerAccount not found")
		}

		var (
			nodeSeed, nodePublicKey string
		)
		if nodePub, ok := certMap["nodePublicKey"]; ok {
			nodePublicKey, ok = nodePub.(string)
			if !ok {
				return fmt.Errorf("invalid certificate format, nodePublicKey should a string")
			}
		} else {
			return fmt.Errorf("invalid certificate format, nodePublicKey not found")
		}

		if seed, ok := certMap["nodeSeed"]; ok {
			nodeSeed, ok = seed.(string)
			if !ok {
				return fmt.Errorf("invalid certificate format, nodeSeed should a string")
			}
		} else {
			return fmt.Errorf("invalid certificate format, nodeSeed not found")
		}

		// verifying NodeSeed
		publicKey := crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
		compareNodeAddress, compareErr := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, publicKey)
		if compareErr != nil {
			return compareErr
		}
		if eq := strings.Compare(nodePublicKey, compareNodeAddress); eq != 0 {
			return fmt.Errorf("invalid certificate format, node seed is wrong format")
		}

		config.NodeSeed = nodeSeed
		break
	}
	return nil
}

func generateConfig(config model.Config) error {
	var (
		shell    = ishell.New()
		port     int
		err      error
		inputStr string
	)

	// SET DEFAULT
	config.MonitoringPort = 9090
	config.CPUProfilingPort = 6060
	config.MaxAPIRequestPerSecond = 10
	config.Smithing = true
	config.ResourcePath = "./resource"
	config.DatabaseFileName = "zoobc.db"
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

	fileInfos, err := filepath.Glob(path.Join(helper.GetAbsDBPath(), "*.zbc"))
	if err != nil {
		return err
	}
	if len(fileInfos) > 1 {
		return errors.New("there are several .zbc files, please delete some and leave one of them")
	}

	if len(fileInfos) != 0 {
		err = readCertFile(&config, fileInfos[0])
		if err != nil {
			return err
		}
		_ = os.Remove(fileInfos[0])
	} else {
		choice := shell.MultiChoice([]string{
			"Input the base64 version of certificate",
			"Input manual the OWNER ADDRESS and NODE SEED",
		}, "Certificate file [*.zbc] not found")
		if choice == 0 {
			err = readCertFile(&config, "*")
			if err != nil {
				return err
			}
		} else {
			color.Cyan("! Create one on zoobc.one")
			color.White("OWNER ACCOUNT ADDRESS: ")
			inputStr = shell.ReadLine()
			inputStr = strings.TrimSpace(inputStr)
			if inputStr != "" {
				// TODO: if in future we accept any account type in genesis, certificate must contain a full account address in hex format.
				//  As of now we have to parse an encoded address into its full hex format to be put in config.toml
				config.OwnerAccountAddress, err = accounttype.ParseEncodedAccountToAccountAddress(
					int32(model.AccountType_ZbcAccountType),
					inputStr,
				)
				if err != nil {
					color.Cyan("Invalid Owner Account Address: It must be a valid ZBC account")
					return err
				}
				config.OwnerAccountAddressHex = hex.EncodeToString(config.OwnerAccountAddress)
			} else {
				if config.OwnerAccountAddressHex != "" {
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

	admin.GenerateNodeKeysFile(config.NodeSeed)
	color.Cyan("Saving configuration ...")
	err = config.SaveConfig("./")
	if err != nil {
		color.Red(err.Error())
	}
	return nil
}

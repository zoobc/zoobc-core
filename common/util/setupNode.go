package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiosoft/ishell"

	"github.com/spf13/viper"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/fatih/color"
)

const (
	ErrNoConfigFile        = "ErrNoConfigFile"
	ErrFatal               = "ErrFatal"               // fatal error, abort process
	ErrFailSavingNewConfig = "ErrFailSavingNewConfig" // fatal error, abort process
)

type SetupNode struct {
}

func NewSetupNode() *SetupNode {
	return &SetupNode{}
}
func (sn *SetupNode) discoverNodeAddress(config *model.Config) error {
	ipAddr, err := (&IPUtil{}).DiscoverNodeAddress()
	if ipAddr == nil {
		// panic if we can't set an IP address for the node
		return errors.New(
			"node's ip can't be discovered, consider using static ip by setting myAddress in config.toml",
		)
	} else if err != nil {
		// notify user that something went wrong in net address discovery process and its node might not behave properly on the network
		color.Yellow("something goes wrong when discovering node's public IP, some feature might not work as expected"+
			"consider using static ip by setting myAddress\nerror: %s\n", err.Error())
	}
	color.Green("discovered node's address: %s", ipAddr.String())
	config.MyAddress = ipAddr.String()
	config.IsNodeAddressDynamic = true
	return nil
}

func (sn *SetupNode) checkConfig(config *model.Config) error {
	if !config.ConfigFileExist {
		return errors.New(ErrNoConfigFile)
	}

	if config.MyAddress == "" {
		color.Cyan("node's address not set, discovering...")
		if err := sn.discoverNodeAddress(config); err != nil {
			return err
		}
	}
	if config.Smithing {
		_, err := os.Stat(filepath.Join(config.ResourcePath, config.NodeKeyFileName))
		if err != nil {
			if ok := os.IsNotExist(err); ok {
				if config.NodeSeed != "" {
					config.NodeKey = &model.NodeKey{
						Seed: config.NodeSeed,
					}
				} else {
					color.Cyan("node keys has not been setup")
					sn.nodeKeysPrompt(config)
				}
			} else {
				color.Red("unknown error occurred when scanning for node keys file")
				return err
			}
		}

	} else {
		color.Yellow("node is not smithing")
	}
	if len(config.WellknownPeers) == 0 {
		color.Yellow("no wellknown peers found, set it in config.toml:wellknownPeers if you are starting " +
			"from scratch.")
	}

	return nil
}

func (sn *SetupNode) nodeKeysPrompt(config *model.Config) {
	c := ishell.New()
	var (
		nodeSeed string
	)
	result := c.MultiChoice([]string{
		"YES", "NO",
	}, "Do you have node's seed you want to use?")
	if result == 0 {
		c.Print("input your node's seed: ")
		nodeSeed = c.ReadLine()
	} else {
		color.Green("Generating secure random seed for node...\n")
		nodeSeed = GetSecureRandomSeed()
		color.Green("Node seed generated\n")
	}
	config.NodeKey = &model.NodeKey{
		Seed: nodeSeed,
	}
}

func (sn *SetupNode) ownerAddressPrompt(config *model.Config) {
	c := ishell.New()
	var (
		ownerAddress string
	)

	c.Print("input your account address to be set as owner of this node: ")
	ownerAddress = c.ReadLine()
	config.OwnerAccountAddress = ownerAddress
	viper.Set("ownerAddress", ownerAddress)
}

func (sn *SetupNode) wellknownPeersPrompt(config *model.Config) {
	c := ishell.New()
	var (
		wellknownPeerString string
	)
	c.Print("provide the peers (space separated) you prefer to connect to (ip:port): ")
	wellknownPeerString = c.ReadLine()
	config.WellknownPeers = strings.Split(strings.TrimSpace(wellknownPeerString), " ")
	viper.Set("wellknownPeers", config.WellknownPeers)
}

func (sn *SetupNode) generateConfig(config *model.Config) error {
	c := ishell.New()
	color.Cyan("generating config\n")
	if err := sn.discoverNodeAddress(config); err != nil {
		return err
	}
	// ask if want to run as blocksmith

	result := c.MultiChoice([]string{
		"YES", "NO",
	}, "Do you want to run node as blocksmith?")
	if result == 0 {
		// node keys prompt
		config.Smithing = true
		viper.Set("smithing", true)
		_, err := os.Stat(filepath.Join(config.ResourcePath, config.NodeKeyFileName))
		if ok := os.IsNotExist(err); ok {
			color.Cyan("node keys has not been setup")
			sn.nodeKeysPrompt(config)
		}
		// ask if have account address prepared as owner
		sn.ownerAddressPrompt(config)
	}
	sn.wellknownPeersPrompt(config)
	// todo: checking port availability and accessibility
	return nil
}

func (sn *SetupNode) WizardFirstSetup(config *model.Config) error {
	color.Green("WELCOME TO ZOOBC\n\n")
	color.Yellow("Checking existing configuration...\n")
	// todo: check config if everything ok, return
	err := sn.checkConfig(config)
	if err != nil {
		if err.Error() == ErrNoConfigFile {
			color.Cyan("no config file found, generating one...\n")
			err := sn.generateConfig(config)
			if err != nil {
				return err
			}
			// save generated config file
			_, err = os.Stat(config.ResourcePath)
			if ok := os.IsNotExist(err); ok {
				color.Cyan("resource folder not found, creating directory...")
				if err := os.Mkdir("resource", os.ModePerm); err != nil {
					return errors.New("fail to create directory")
				}
				color.Green("resource directory created")
			}
			color.Yellow("saving new configurations")
			err = viper.SafeWriteConfigAs("./config.toml")
			if err != nil {
				return errors.New(ErrFailSavingNewConfig)
			}
			color.Green("configuration saved successfully in ./config.toml")
			color.Green("continue to run node with provided configurations")
		} else {
			color.Red("failed reading / creating config file, error: %s\tstopping node...\n", err.Error())
			return errors.New(ErrFatal)
		}
	} else {
		color.Green("continue to run node with ./config.toml configurations")
	}
	return nil
}

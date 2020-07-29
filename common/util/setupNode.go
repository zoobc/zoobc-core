package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/model"
)

const (
	ErrNoConfigFile        = "ErrNoConfigFile"
	ErrFatal               = "ErrFatal"               // fatal error, abort process
	ErrFailSavingNewConfig = "ErrFailSavingNewConfig" // fatal error, abort process
)

type SetupNode struct {
	Config *model.Config
}

func NewSetupNode(config *model.Config) *SetupNode {
	return &SetupNode{Config: config}
}
func (sn *SetupNode) discoverNodeAddress() error {
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
	sn.Config.MyAddress = ipAddr.String()
	sn.Config.IsNodeAddressDynamic = true
	return nil
}

func (sn *SetupNode) checkConfig() error {
	if !sn.Config.ConfigFileExist {
		return errors.New(ErrNoConfigFile)
	}

	if sn.Config.MyAddress == "" {
		color.Cyan("node's address not set, discovering...")
		if err := sn.discoverNodeAddress(); err != nil {
			return err
		}
	}
	_, err := os.Stat(filepath.Join(sn.Config.ResourcePath, sn.Config.NodeKeyFileName))
	if err != nil {
		if ok := os.IsNotExist(err); ok {
			if sn.Config.NodeSeed != "" {
				sn.Config.NodeKey = &model.NodeKey{
					Seed: sn.Config.NodeSeed,
				}
			} else {
				color.Cyan("node keys has not been setup")
				sn.nodeKeysPrompt()
			}
		} else {
			color.Red("unknown error occurred when scanning for node keys file")
			return err
		}
	}

	if len(sn.Config.WellknownPeers) == 0 {
		color.Yellow("no wellknown peers found, set it in config.toml:wellknownPeers if you are starting " +
			"from scratch.")
	}

	sn.Config.PeerPort = uint32(sn.portAvailability("PEER", int(sn.Config.PeerPort)))
	sn.Config.RPCAPIPort = sn.portAvailability("API", sn.Config.RPCAPIPort)
	sn.Config.HTTPAPIPort = sn.portAvailability("API PROXY", sn.Config.HTTPAPIPort)

	return nil
}

func (sn *SetupNode) nodeKeysPrompt() {
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
	sn.Config.NodeKey = &model.NodeKey{
		Seed: nodeSeed,
	}
}

func (sn *SetupNode) ownerAddressPrompt() {
	c := ishell.New()
	var (
		ownerAddress string
	)

	c.Print("input your account address to be set as owner of this node: ")
	ownerAddress = c.ReadLine()
	sn.Config.OwnerAccountAddress = ownerAddress
	viper.Set("ownerAddress", ownerAddress)
}

func (sn *SetupNode) wellknownPeersPrompt() {
	c := ishell.New()
	var (
		wellknownPeerString string
	)
	c.Print("provide the peers (space separated) you prefer to connect to (ip:port): ")
	wellknownPeerString = c.ReadLine()
	sn.Config.WellknownPeers = strings.Split(strings.TrimSpace(wellknownPeerString), " ")
	viper.Set("wellknownPeers", sn.Config.WellknownPeers)
}

func (sn *SetupNode) generateConfig() error {
	c := ishell.New()
	color.Cyan("generating config\n")
	if err := sn.discoverNodeAddress(); err != nil {
		return err
	}
	// ask if want to run as blocksmith

	result := c.MultiChoice([]string{
		"YES", "NO",
	}, "Do you want to run node as blocksmith?")
	if result == 0 {
		// node keys prompt
		sn.Config.Smithing = true
		viper.Set("smithing", true)
		_, err := os.Stat(filepath.Join(sn.Config.ResourcePath, sn.Config.NodeKeyFileName))
		if ok := os.IsNotExist(err); ok {
			color.Cyan("node keys has not been setup")
			sn.nodeKeysPrompt()
		}
		// ask if have account address prepared as owner
		sn.ownerAddressPrompt()
	}
	sn.wellknownPeersPrompt()
	// todo: checking port availability and accessibility
	return nil
}

func (sn *SetupNode) WizardFirstSetup() error {
	color.Green("WELCOME TO ZOOBC\n\n")
	color.Yellow("Checking existing configuration...\n")
	// todo: check config if everything ok, return
	err := sn.checkConfig()
	if err != nil {
		if err.Error() == ErrNoConfigFile {
			color.Cyan("no config file found, generating one...\n")
			err := sn.generateConfig()
			if err != nil {
				return err
			}
			// save generated config file
			_, err = os.Stat(sn.Config.ResourcePath)
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

// portAvailability allowing to check port availability for peer and api grpc connection need to use
func (sn *SetupNode) portAvailability(portType model.PortType, port int) (availablePort int) {
	portCollection := map[model.PortType]int{
		model.PeerPort:    int(sn.Config.PeerPort),
		model.RPCAPIPort:  sn.Config.RPCAPIPort,
		model.HTTPAPIPort: sn.Config.HTTPAPIPort,
	}

	if _, ok := portCollection[portType]; ok {
		port++
		sn.portAvailability(portType, port)
	}
	for i := 0; i <= 5; i++ {
		if i == 5 {
			color.Red("too long, pick another port for %v manually!", portType)
			break
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			color.Yellow("%s", err.Error())
			port++
			continue
		}
		_ = ln.Close()
		break
	}
	return port
}

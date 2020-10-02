package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/zoobc/zoobc-core/common/model"
)

const (
	ErrNoConfigFile = "ErrNoConfigFile"
)

type SetupNode struct {
	Config *model.Config
}

func NewSetupNode(config *model.Config) *SetupNode {
	return &SetupNode{
		Config: config,
	}
}

// CheckConfig checking configuration files, validate keys of it that would important to start the app
func (sn *SetupNode) CheckConfig() error {
	if !sn.Config.ConfigFileExist {
		return errors.New(ErrNoConfigFile)
	}

	if sn.Config.MyAddress == "" {
		color.Cyan("node's address not set, discovering...")
		if err := sn.discoverNodeAddress(); err != nil {
			return err
		}
	}
	err := sn.checkResourceFolder()
	if err != nil {
		return err
	}
	_, err = os.Stat(filepath.Join(sn.Config.ResourcePath, sn.Config.NodeKeyFileName))
	if err != nil {
		if ok := os.IsNotExist(err); ok {
			if sn.Config.NodeSeed != "" {
				sn.Config.NodeKey = &model.NodeKey{
					Seed: sn.Config.NodeSeed,
				}
			} else {
				return errors.New("nod keys has not been setup")
			}
		} else {
			return err
		}
	}

	if len(sn.Config.WellknownPeers) < 1 {
		return errors.New("no wellknown peers found")
	}
	if sn.Config.Smithing && sn.Config.OwnerAccountAddress == nil {
		return errors.New("no owner account address found")
	}
	peePort, err := sn.portAvailability("PEER", int(sn.Config.PeerPort))
	if err != nil {
		return err
	}
	sn.Config.PeerPort = uint32(peePort)
	rpcAPIPort, err := sn.portAvailability("API", sn.Config.RPCAPIPort)
	if err != nil {
		return err
	}
	sn.Config.RPCAPIPort = rpcAPIPort

	httpAPIPort, err := sn.portAvailability("API PROXY", sn.Config.HTTPAPIPort)
	if err != nil {
		return err
	}
	sn.Config.HTTPAPIPort = httpAPIPort
	return nil
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

func (sn *SetupNode) checkResourceFolder() error {
	_, err := os.Stat(sn.Config.ResourcePath)
	if ok := os.IsNotExist(err); ok {
		color.Cyan("resource folder not found in %s, creating directory...", sn.Config.ResourcePath)
		if err := os.Mkdir("resource", os.ModePerm); err != nil {
			return errors.New("fail to create directory")
		}
		color.Green("resource directory created")
	}
	return nil
}

// portAvailability allowing to check port availability for peer and api grpc connection need to use
func (sn *SetupNode) portAvailability(portType model.PortType, port int) (availablePort int, err error) {
	portCollection := map[model.PortType]int{
		model.PeerPort:    int(sn.Config.PeerPort),
		model.RPCAPIPort:  sn.Config.RPCAPIPort,
		model.HTTPAPIPort: sn.Config.HTTPAPIPort,
	}

	if _, ok := portCollection[portType]; ok {
		port++
		port, err = sn.portAvailability(portType, port)
		if err != nil {
			return port, err
		}
	}
	for i := 0; i <= model.PortChangePeriod; i++ {
		if i == model.PortChangePeriod {
			return port, fmt.Errorf("too long, pick another port for %v manually", portType)
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
	return port, err
}

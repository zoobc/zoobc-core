// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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

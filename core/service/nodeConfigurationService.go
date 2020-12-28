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
package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/signaturetype"
)

type (
	NodeConfigurationServiceInterface interface {
		SetMyAddress(nodeAddress string, port uint32)
		GetMyAddress() (string, error)
		GetMyPeerPort() (uint32, error)
		SetIsMyAddressDynamic(nodeAddressDynamic bool)
		IsMyAddressDynamic() bool
		GetHost() *model.Host
		SetHostID(nodeID int64)
		GetHostID() (int64, error)
		GetNodeSecretPhrase() string
		GetNodePublicKey() []byte
		SetHost(host *model.Host)
		SetNodeSeed(seed string)
	}
)

type (
	NodeConfigurationService struct {
		Logger *log.Logger
		host   *model.Host
	}
	NodeConfigurationServiceHelper struct{}
)

var (
	secretPhrase                     string
	isMyAddressDynamic               bool
	NodeConfigurationServiceInstance *NodeConfigurationService
)

func NewNodeConfigurationService(logger *log.Logger) *NodeConfigurationService {
	if NodeConfigurationServiceInstance == nil {
		NodeConfigurationServiceInstance = &NodeConfigurationService{
			Logger: logger,
		}
		return NodeConfigurationServiceInstance
	}
	NodeConfigurationServiceInstance.Logger = logger
	return NodeConfigurationServiceInstance
}

func (nss *NodeConfigurationService) SetNodeSeed(seed string) {
	secretPhrase = seed
}

func (nss *NodeConfigurationService) GetNodeSecretPhrase() string {
	return secretPhrase
}

func (nss *NodeConfigurationService) GetNodePublicKey() []byte {
	if sp := nss.GetNodeSecretPhrase(); sp == "" {
		return []byte{}
	}
	return signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nss.GetNodeSecretPhrase())
}

func (nss *NodeConfigurationService) SetHost(host *model.Host) {
	nss.host = host
}

func (nss *NodeConfigurationService) SetMyAddress(nodeAddress string, nodePort uint32) {
	nss.host.Info.Address = nodeAddress
	nss.host.Info.Port = nodePort
}

func (nss *NodeConfigurationService) SetHostID(nodeID int64) {
	if nss.host != nil {
		nss.host.Info.ID = nodeID
	}
}

func (nss *NodeConfigurationService) GetHostID() (int64, error) {
	if nss.host == nil || nss.host.Info == nil || nss.host.Info.ID == 0 {
		return 0, blocker.NewBlocker(blocker.AppErr, "host id not set")
	}
	return nss.host.Info.ID, nil
}

func (nss *NodeConfigurationService) GetMyAddress() (string, error) {
	if nss.host != nil {
		return nss.host.Info.Address, nil
	}
	return "", blocker.NewBlocker(blocker.AppErr, "node address not set")
}

func (nss *NodeConfigurationService) GetMyPeerPort() (uint32, error) {
	if nss.host != nil && nss.host.Info.Port > 0 {
		return nss.host.Info.Port, nil
	}
	return 0, blocker.NewBlocker(blocker.AppErr, "node peer port not set")
}

func (nss *NodeConfigurationService) SetIsMyAddressDynamic(nodeAddressDynamic bool) {
	isMyAddressDynamic = nodeAddressDynamic
}

func (nss *NodeConfigurationService) IsMyAddressDynamic() bool {
	return isMyAddressDynamic
}

func (nss *NodeConfigurationService) GetHost() *model.Host {
	return nss.host
}

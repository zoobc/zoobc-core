package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
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
	return crypto.NewEd25519Signature().GetPublicKeyFromSeed(nss.GetNodeSecretPhrase())
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

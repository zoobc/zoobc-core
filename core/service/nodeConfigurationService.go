package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeConfigurationServiceInterface interface {
		SetMyAddress(nodeAddress string)
		GetMyAddress() (string, error)
		GetMyPeerPort() (uint32, error)
		SetIsMyAddressDynamic(nodeAddressDynamic bool)
		IsMyAddressDynamic() bool
		GetHost() *model.Host
		SetHost(myHost *model.Host)
		GetNodeSecretPhrase() string
	}
)

type (
	NodeConfigurationService struct {
		Logger *log.Logger
	}
)

var (
	secretPhrase       string
	isMyAddressDynamic bool
	host               *model.Host
)

func NewNodeConfigurationService(
	nodeAddressDynamic bool,
	sp string,
	logger *log.Logger,
) *NodeConfigurationService {
	var nss = &NodeConfigurationService{
		Logger: logger,
	}
	secretPhrase = sp
	isMyAddressDynamic = nodeAddressDynamic
	return nss
}

func (nss *NodeConfigurationService) GetNodeSecretPhrase() string {
	return secretPhrase
}

func (nss *NodeConfigurationService) SetMyAddress(nodeAddress string) {
	if host != nil {
		host.Info.Address = nodeAddress
	}
}

func (nss *NodeConfigurationService) GetMyAddress() (string, error) {
	if host != nil {
		return host.Info.Address, nil
	}
	return "", blocker.NewBlocker(blocker.AppErr, "node address not set")
}

func (nss *NodeConfigurationService) GetMyPeerPort() (uint32, error) {
	if host != nil && host.Info.Port > 0 {
		return host.Info.Port, nil
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
	return host
}

func (nss *NodeConfigurationService) SetHost(myHost *model.Host) {
	host = myHost
}

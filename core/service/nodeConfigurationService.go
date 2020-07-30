package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abiosoft/ishell"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"io/ioutil"
	"os"
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
		ImportWalletCertificate(config *model.Config) error
	}
	NodeConfigurationServiceHelperInterface interface {
		ReadPassword(c *ishell.Shell) string
	}
)

type (
	NodeConfigurationService struct {
		Logger        *log.Logger
		host          *model.Host
		ServiceHelper NodeConfigurationServiceHelperInterface
	}
	NodeConfigurationServiceHelper struct{}
)

var (
	secretPhrase                     string
	isMyAddressDynamic               bool
	NodeConfigurationServiceInstance *NodeConfigurationService
)

func NewNodeConfigurationService(
	nodeAddressDynamic bool,
	sp string,
	logger *log.Logger,
	host *model.Host,
	serviceHelper NodeConfigurationServiceHelperInterface,
) *NodeConfigurationService {
	secretPhrase = sp
	isMyAddressDynamic = nodeAddressDynamic

	if NodeConfigurationServiceInstance == nil {
		NodeConfigurationServiceInstance = &NodeConfigurationService{
			Logger:        logger,
			host:          host,
			ServiceHelper: serviceHelper,
		}
		return NodeConfigurationServiceInstance
	}
	NodeConfigurationServiceInstance.Logger = logger
	NodeConfigurationServiceInstance.host = host
	return NodeConfigurationServiceInstance
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

func (nss *NodeConfigurationService) ImportWalletCertificate(config *model.Config) error {
	var (
		certMap               map[string]interface{}
		nodeKeyFieldName      = "nodeKey"
		ownerAccountFieldName = "ownerAccount"
	)

	jsonFile, err := os.Open(config.WalletCertFileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	c := ishell.New()
	c.Print("A wallet certificate has been found. Please enter the password to decrypt and import it:")
	var i int
	for i = 0; i < 4; i++ {
		if i > 3 {
			return errors.New("maximum numbers of attempts exceeded")
		}
		pwd := nss.ServiceHelper.ReadPassword(c)
		nodeKeyDecryptedBytes, err := crypto.OpenSSLDecrypt(pwd, string(byteValue))
		if err != nil {
			continue
		}

		if err := json.Unmarshal(nodeKeyDecryptedBytes, &certMap); err != nil {
			return err
		}
		// validate certificate fields
		nodeSeed, ok := certMap[nodeKeyFieldName]
		if !ok {
			return errors.New("wallet certificate malformed: nodeKey not found")
		}
		ownerAccount, ok := certMap[ownerAccountFieldName]
		if !ok {
			return errors.New("wallet certificate malformed: ownerAccount not found")
		}
		// import into node configuration
		config.NodeSeed = fmt.Sprintf("%s", nodeSeed)
		config.NodeKey = &model.NodeKey{
			Seed: config.NodeSeed,
		}
		config.OwnerAccountAddress = fmt.Sprintf("%s", ownerAccount)
		break
	}
	return nil
}

// ReadPassword wrapper around the ishell ReadPassword method, which is not testable otherwise
func (nssH *NodeConfigurationServiceHelper) ReadPassword(c *ishell.Shell) string {
	return c.ReadPassword()
}

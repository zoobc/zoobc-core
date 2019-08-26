package util

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	NodeKeysInterface interface {
		ParseKeysFile() ([]*NodeKey, error)
		GetLastNodeKey() (*NodeKey, error)
		GenerateNodeKey(seed string) ([]byte, error)
	}
	NodeKeyConfig struct {
		filePath string
	}

	NodeKey struct {
		ID        int64
		PublicKey string
		Seed      string
	}
)

func NewNodeKeyConfig() *NodeKeyConfig {
	configPath := viper.GetString("configPath")
	nodeKeysFileName := viper.GetString("nodeKeyFile")
	if nodeKeysFileName == "" {
		return nil
	}
	return &NodeKeyConfig{
		filePath: filepath.Join(configPath, nodeKeysFileName),
	}
}

// ParseNodeKeysFile read the node key file and parses it into an array of NodeKey stuct
func (nk *NodeKeyConfig) ParseKeysFile() ([]*NodeKey, error) {
	file, err := ioutil.ReadFile(nk.filePath)
	if err != nil && os.IsNotExist(err) {
		return nil, blocker.NewBlocker(blocker.AppErr, "NodeKeysFileNotExist")
	}
	data := make([]*NodeKey, 0)
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "InvalidNodeKeysFile")
	}
	return data, nil
}

// GetLastNodeKey retrieves the last node key object from the node_key configuration file
func (*NodeKeyConfig) GetLastNodeKey(nodeKeys []*NodeKey) *NodeKey {
	max := nodeKeys[0]
	for _, nodeKey := range nodeKeys {
		if nodeKey.ID > max.ID {
			max = nodeKey
		}
	}
	return max
}

// GenerateNodeKey generates a new node ket from its seed and store it, together with relative public key into node_keys file
func (nk *NodeKeyConfig) GenerateNodeKey(seed string) ([]byte, error) {
	publicKey := util.GetPublicKeyFromSeed(seed)
	nodeKey := &NodeKey{
		Seed:      seed,
		PublicKey: hex.EncodeToString(publicKey),
	}

	nodeKeys := make([]*NodeKey, 0)
	_, err := os.Stat(nk.filePath)
	if !(err != nil && os.IsNotExist(err)) {
		// if there are previous keys, get the new id
		nodeKeys, err = nk.ParseKeysFile()
		if err != nil {
			return nil, err
		}
		lastNodeKey := nk.GetLastNodeKey(nodeKeys)
		nodeKey.ID = lastNodeKey.ID + 1
	}

	// append generated key to previous keys array
	nodeKeys = append(nodeKeys, nodeKey)
	file, _ := json.MarshalIndent(nodeKeys, "", " ")
	err = ioutil.WriteFile(nk.filePath, file, 0644)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorWritingNodeKeysFile")
	}

	return publicKey, nil
}

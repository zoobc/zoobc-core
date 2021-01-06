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
	"encoding/json"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"io/ioutil"
	"os"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GenerateProofOfOwnership(accountAddress []byte) (*model.ProofOfOwnership, error)
		ParseKeysFile() ([]*model.NodeKey, error)
		GetLastNodeKey(nodeKeys []*model.NodeKey) *model.NodeKey
		GenerateNodeKey(seed string) ([]byte, error)
	}

	// NodeAdminServiceHelpersInterface mockable service methods
	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
		FilePath      string
	}
)

func NewNodeAdminService(
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	signature crypto.SignatureInterface,
	blockService BlockServiceInterface,
	nodeKeyFilePath string) *NodeAdminService {
	return &NodeAdminService{
		QueryExecutor: queryExecutor,
		BlockQuery:    blockQuery,
		Signature:     signature,
		BlockService:  blockService,
		FilePath:      nodeKeyFilePath,
	}
}

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(
	accountAddress []byte) (*model.ProofOfOwnership, error) {

	// get the node seed (private key)
	nodeKeys, err := nas.ParseKeysFile()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorParseKeysFile: ,"+err.Error())
	}
	nodeKey := nas.GetLastNodeKey(nodeKeys)
	if nodeKey == nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "MissingNodePrivateKey")
	}

	lastBlock, err := nas.BlockService.GetLastBlock()
	if err != nil {
		return nil, err
	}
	// get the blockhash of a block that most likely have been already downloaded by all nodes
	// so that, every node will be able to validate it
	if lastBlock.Height > constant.BlockHeightOffset {
		lastBlock, err = nas.BlockService.GetBlockByHeight(lastBlock.Height - constant.BlockHeightOffset)
		if err != nil {
			return nil, err
		}
	}
	lastBlockHash, err := commonUtils.GetBlockHash(lastBlock, nas.BlockService.GetChainType())
	if err != nil {
		return nil, err
	}

	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: accountAddress,
		BlockHash:      lastBlockHash,
		BlockHeight:    lastBlock.Height,
	}

	messageBytes := commonUtils.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(messageBytes, nodeKey.Seed)
	return &model.ProofOfOwnership{
		MessageBytes: messageBytes,
		Signature:    poownSignature,
	}, nil
}

// ParseNodeKeysFile read the node key file and parses it into an array of NodeKey struct
func (nas *NodeAdminService) ParseKeysFile() ([]*model.NodeKey, error) {
	file, err := ioutil.ReadFile(nas.FilePath)
	if err != nil && os.IsNotExist(err) {
		return nil, blocker.NewBlocker(blocker.AppErr, "NodeKeysFileNotExist")
	}

	data := make([]*model.NodeKey, 0)
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "InvalidNodeKeysFile")
	}
	return data, nil
}

// GetLastNodeKey retrieves the last node key object from the node_key configuration file
func (*NodeAdminService) GetLastNodeKey(nodeKeys []*model.NodeKey) *model.NodeKey {
	if len(nodeKeys) == 0 {
		return nil
	}
	max := nodeKeys[0]
	for _, nodeKey := range nodeKeys {
		if nodeKey.ID > max.ID {
			max = nodeKey
		}
	}
	return max
}

// GenerateNodeKey generates a new node key from its seed and store it, together with relative public key into node_keys file
func (nas *NodeAdminService) GenerateNodeKey(seed string) ([]byte, error) {
	publicKey := signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(seed)
	nodeKey := &model.NodeKey{
		Seed:      seed,
		PublicKey: publicKey,
	}
	nodeKeys := make([]*model.NodeKey, 0)

	_, err := os.Stat(nas.FilePath)
	if !(err != nil && os.IsNotExist(err)) {
		// if there are previous keys, get the new id
		nodeKeys, err = nas.ParseKeysFile()
		if err != nil {
			return nil, err
		}
		lastNodeKey := nas.GetLastNodeKey(nodeKeys)
		nodeKey.ID = lastNodeKey.ID + 1
	}

	// append generated key to previous keys array
	nodeKeys = append(nodeKeys, nodeKey)
	file, err := json.MarshalIndent(nodeKeys, "", " ")
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorMarshalingNodeKeys: "+err.Error())
	}
	err = ioutil.WriteFile(nas.FilePath, file, 0644)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorWritingNodeKeysFile: "+err.Error())
	}

	return publicKey, nil
}

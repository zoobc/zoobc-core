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
package strategy

import (
	"sort"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

func GetActiveSpinePublicKeysByBlockHeight(QueryExecutor query.ExecutorInterface, SpinePublicKeyQuery query.SpinePublicKeyQueryInterface, height uint32) (spinePublicKeys []*model.SpinePublicKey, err error) {
	rows, err := QueryExecutor.ExecuteSelect(SpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(0, height), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	spinePublicKeys, err = SpinePublicKeyQuery.BuildModel(spinePublicKeys, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return spinePublicKeys, nil
}

// GetActiveNodesInSpineBlocks get the active nodes either from cache (
// main blocks), or from from spine_pub_keys if spine blocks
func GetActiveNodesInSpineBlocks(QueryExecutor query.ExecutorInterface, SpinePublicKeyQuery query.SpinePublicKeyQueryInterface, block *model.Block) (activeNodeRegistry []storage.NodeRegistry, err error) {
	var (
		spinePubKeys []*model.SpinePublicKey
	)
	spinePubKeys, err = GetActiveSpinePublicKeysByBlockHeight(QueryExecutor,
		SpinePublicKeyQuery, block.GetHeight())
	if err != nil {
		return
	}
	// update local spine pub keys with the ones in downloaded block in case there are newly added/removed nodes from registry since prev
	// block
	for _, blockPubKey := range block.GetSpinePublicKeys() {
		switch blockPubKey.GetPublicKeyAction() {
		case model.SpinePublicKeyAction_RemoveKey:
			for idx, spinePubKey := range spinePubKeys {
				if blockPubKey.GetNodeID() == spinePubKey.GetNodeID() {
					// remove element from spinePubKeys
					spinePubKeys = append(spinePubKeys[:idx], spinePubKeys[idx+1:]...)
					break
				}
			}
		case model.SpinePublicKeyAction_AddKey:
			var found = false
			for idx, spinePubKey := range spinePubKeys {
				if blockPubKey.GetNodeID() == spinePubKey.GetNodeID() {
					// update element from spinePubKeys with new one (already registered node, updated node pub key)
					spinePubKeys[idx] = blockPubKey
					found = true
					break
				}
			}
			if !found {
				// add new spine pub key (node registered after previous spine block)
				spinePubKeys = append(spinePubKeys, blockPubKey)
			}
		}

	}
	// sort by nodeID (same sort as in ActiveNodeRegistryCacheStorage.activeNodeRegistry)
	sort.SliceStable(spinePubKeys, func(i, j int) bool {
		// sort by nodeID lowest - highest
		return spinePubKeys[i].GetNodeID() < spinePubKeys[j].GetNodeID()
	})
	for _, spinePubKey := range spinePubKeys {
		var anr = storage.NodeRegistry{
			Node: model.NodeRegistration{
				NodeID:        spinePubKey.GetNodeID(),
				NodePublicKey: spinePubKey.GetNodePublicKey(),
			},
			// mock this value since we don't have it in spine public keys
			// anyway if a public key is in spine pub keys it means that node has positive score
			ParticipationScore: constant.DefaultParticipationScore,
		}
		activeNodeRegistry = append(activeNodeRegistry, anr)
	}
	return activeNodeRegistry, nil
}

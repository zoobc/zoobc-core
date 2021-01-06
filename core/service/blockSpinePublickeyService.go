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
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockSpinePublicKeyServiceInterface interface {
		BuildSpinePublicKeysFromNodeRegistry(
			mainFromHeight,
			mainToHeight,
			spineHeight uint32,
		) (spinePublicKeys []*model.SpinePublicKey, err error)
		GetSpinePublicKeysByBlockHeight(height uint32) (spinePublicKeys []*model.SpinePublicKey, err error)
		GetValidSpinePublicKeyByBlockHeightInterval(
			fromHeight, toHeight uint32,
		) (
			[]*model.SpinePublicKey, error,
		)
		InsertSpinePublicKeys(block *model.Block) error
	}

	BlockSpinePublicKeyService struct {
		Signature             crypto.SignatureInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
	}
)

func NewBlockSpinePublicKeyService(
	signature crypto.SignatureInterface,
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	logger *log.Logger,
) *BlockSpinePublicKeyService {
	return &BlockSpinePublicKeyService{
		Signature:             signature,
		QueryExecutor:         queryExecutor,
		NodeRegistrationQuery: nodeRegistrationQuery,
		SpinePublicKeyQuery:   spinePublicKeyQuery,
		Logger:                logger,
	}
}

// GetValidSpinePublicKeyByBlockHeightInterval return the spine_public_key rows that were valid
func (bsf *BlockSpinePublicKeyService) GetValidSpinePublicKeyByBlockHeightInterval(
	fromHeight, toHeight uint32,
) (
	[]*model.SpinePublicKey, error,
) {
	var validSpinePublicKeys []*model.SpinePublicKey
	// get all registered nodes with participation score > 0
	rows, err := bsf.QueryExecutor.ExecuteSelect(bsf.SpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(fromHeight, toHeight), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	validSpinePublicKeys, err = bsf.SpinePublicKeyQuery.BuildModel(validSpinePublicKeys, rows)
	return validSpinePublicKeys, err
}

func (bsf *BlockSpinePublicKeyService) GetSpinePublicKeysByBlockHeight(height uint32) (spinePublicKeys []*model.SpinePublicKey, err error) {
	rows, err := bsf.QueryExecutor.ExecuteSelect(bsf.SpinePublicKeyQuery.GetSpinePublicKeysByBlockHeight(height), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	spinePublicKeys, err = bsf.SpinePublicKeyQuery.BuildModel(spinePublicKeys, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return spinePublicKeys, nil
}

// BuildSpinePublicKeysFromNodeRegistry build the list of spine public keys from the node registry
func (bsf *BlockSpinePublicKeyService) BuildSpinePublicKeysFromNodeRegistry(
	mainFromHeight,
	mainToHeight,
	spineHeight uint32,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	var (
		nodeRegistrations []*model.NodeRegistration
	)
	qry := bsf.NodeRegistrationQuery.GetNodeRegistrationsByBlockHeightInterval(mainFromHeight, mainToHeight)
	rows, err := bsf.QueryExecutor.ExecuteSelect(
		qry,
		false,
	)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	nodeRegistrations, err = bsf.NodeRegistrationQuery.BuildModel(nodeRegistrations, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	spinePublicKeys = make([]*model.SpinePublicKey, 0)
	for _, nr := range nodeRegistrations {
		spinePublicKey := &model.SpinePublicKey{
			NodePublicKey:   nr.NodePublicKey,
			NodeID:          nr.NodeID,
			PublicKeyAction: util.GetAddRemoveSpineKeyAction(nr.RegistrationStatus),
			MainBlockHeight: nr.Height, // (node registration) transaction's height
			Height:          spineHeight,
			Latest:          true,
		}
		spinePublicKeys = append(spinePublicKeys, spinePublicKey)
	}
	return spinePublicKeys, nil
}

// InsertSpinePublicKeys insert all spine block publicKeys into spinePublicKey table
// Note: at this stage the spine pub keys have already been parsed into their model struct
func (bsf *BlockSpinePublicKeyService) InsertSpinePublicKeys(block *model.Block) error {
	queries := make([][]interface{}, 0)
	for _, spinePublicKey := range block.SpinePublicKeys {
		insertSpkQry := bsf.SpinePublicKeyQuery.InsertSpinePublicKey(spinePublicKey)
		queries = append(queries, insertSpkQry...)
	}
	if err := bsf.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}

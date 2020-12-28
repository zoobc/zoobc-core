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
// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// BlockServiceInterface represents interface for BlockService
	BlockServiceInterface interface {
		GetBlockByID(chainType chaintype.ChainType, ID int64) (*model.GetBlockResponse, error)
		GetBlockByHeight(chainType chaintype.ChainType, Height uint32) (*model.GetBlockResponse, error)
		GetBlocks(chainType chaintype.ChainType, Count uint32, Height uint32) (*model.GetBlocksResponse, error)
	}

	// BlockService represents struct of BlockService
	BlockService struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
		isDebugMode       bool
	}
)

var blockServiceInstance *BlockService

// NewBlockService create a singleton instance of BlockService
func NewBlockService(queryExecutor query.ExecutorInterface, blockCoreServices map[int32]coreService.BlockServiceInterface,
	isDebugMode bool) *BlockService {
	if blockServiceInstance == nil {
		blockServiceInstance = &BlockService{Query: queryExecutor}
	}
	blockServiceInstance.BlockCoreServices = blockCoreServices
	blockServiceInstance.isDebugMode = isDebugMode
	return blockServiceInstance
}

// GetBlockByID fetch a single block from Blockchain by providing block ID
func (bs *BlockService) GetBlockByID(chainType chaintype.ChainType, id int64) (*model.GetBlockResponse, error) {
	var (
		err   error
		block model.Block
		row   *sql.Row
	)
	blockQuery := query.NewBlockQuery(chainType)

	row, _ = bs.Query.ExecuteSelectRow(blockQuery.GetBlockByID(id), false)
	err = blockQuery.Scan(&block, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &model.GetBlockResponse{
		ChainType: chainType.GetTypeInt(),
		Block:     &block,
	}, nil
}

// GetBlockByHeight fetches a single block from Blockchain by providing block size
func (bs *BlockService) GetBlockByHeight(chainType chaintype.ChainType, height uint32) (*model.GetBlockResponse, error) {
	var (
		err   error
		block model.Block
		row   *sql.Row
	)
	blockQuery := query.NewBlockQuery(chainType)
	row, _ = bs.Query.ExecuteSelectRow(blockQuery.GetBlockByHeight(height), false)
	err = blockQuery.Scan(&block, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &model.GetBlockResponse{
		ChainType: chainType.GetTypeInt(),
		Block:     &block,
	}, nil
}

// GetBlocks fetches multiple blocks from Blockchain system
func (bs *BlockService) GetBlocks(chainType chaintype.ChainType, blockSize, height uint32) (*model.GetBlocksResponse, error) {
	var (
		rows   *sql.Rows
		err    error
		blocks []*model.Block
	)
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlocks(height, blockSize), false)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	blocks, err = blockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed build block into model")
	}
	blocksResponse := &model.GetBlocksResponse{
		Blocks: blocks,
		Height: height,
		Count:  uint32(len(blocks)),
	}
	return blocksResponse, nil
}

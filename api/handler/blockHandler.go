package handler

import (
	"context"
	"fmt"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BlockHandler to handle request related to Blocks from client
type BlockHandler struct {
	Service service.BlockServiceInterface // Use Blockservice Interface
}

// GetBlock handles request to get data of a single Block
func (bs *BlockHandler) GetBlock(ctx context.Context, req *model.GetBlockRequest) (*model.GetBlockResponse, error) {
	var (
		blockResponse *model.GetBlockResponse
		err           error
	)
	chainType := chaintype.GetChainType(req.ChainType)
	if req.ID != 0 {
		blockResponse, err = bs.Service.GetBlockByID(chainType, req.ID)
	}
	if req.Height != 0 {
		blockResponse, err = bs.Service.GetBlockByHeight(chainType, req.Height)
	}
	if err != nil {
		return nil, err
	}

	return blockResponse, nil
}

// GetBlocks handles request to get data of multiple blocks
func (bs *BlockHandler) GetBlocks(ctx context.Context, req *model.GetBlocksRequest) (*model.GetBlocksResponse, error) {
	if req.Limit > constant.MaxAPILimitPerPage {
		return nil, status.Error(codes.OutOfRange, fmt.Sprintf("limit exceeded, max. %d", constant.MaxAPILimitPerPage))
	}

	chainType := chaintype.GetChainType(req.ChainType)
	blocksResponse, err := bs.Service.GetBlocks(chainType, req.Limit, req.Height)
	if err != nil {
		return nil, err
	}

	return blocksResponse, nil
}

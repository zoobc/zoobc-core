package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	SnapshotFileServiceMainBlock struct {
		QueryExecutor   query.ExecutorInterface
		SpineBlockQuery query.BlockQueryInterface
		MainBlockQuery  query.BlockQueryInterface
		Logger          *log.Logger
		// below fields are for better code testability
		SnapshotInterval          uint32
		SnapshotGenerationTimeout int64
		SpineBlockManifestService SpineBlockManifestServiceInterface
		SpineBlockDownloadService SpineBlockDownloadServiceInterface
	}
)

func NewSnapshotFileServiceMainBlock(
	queryExecutor query.ExecutorInterface,
	mainBlockQuery, spineBlockQuery query.BlockQueryInterface,
	spineBlockManifestService SpineBlockManifestServiceInterface,
	spineBlockDownloadService SpineBlockDownloadServiceInterface,
	logger *log.Logger,
) *SnapshotFileServiceMainBlock {
	return &SnapshotFileServiceMainBlock{
		QueryExecutor:             queryExecutor,
		SpineBlockQuery:           spineBlockQuery,
		MainBlockQuery:            mainBlockQuery,
		SnapshotInterval:          constant.MainchainSnapshotInterval,
		SnapshotGenerationTimeout: constant.SnapshotGenerationTimeout,
		SpineBlockManifestService: spineBlockManifestService,
		Logger:                    logger,
		SpineBlockDownloadService: spineBlockDownloadService,
	}
}

func (ss *SnapshotService) NewSnapshotFileNewSnapshotFile(block *model.Block) (*model.SnapshotFileInfo, error) {
	return nil, nil
}

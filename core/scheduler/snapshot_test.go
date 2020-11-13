package scheduler

import (
	"crypto/sha256"
	"errors"
	"math/rand"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
	"github.com/zoobc/zoobc-core/testfixtures"
)

type (
	mockBlockStateStorageCheckSnapshotIntegrityFilled struct {
		storage.BlockStateStorage
	}
	mockSpineBlockManifestCheckSnapshotIntegrityFilled struct {
		service.SpineBlockManifestService
	}
	mockBlockSpinePublicKeysCheckSnapshotIntegrityFilled struct {
		service.BlockSpinePublicKeyService
	}
	mockSnapshotChunkUtilCheckSnapshotIntegrityFilled struct {
		util.ChunkUtil
	}
	mockNodeConfigurationCheckSnapshotIntegrityFilled struct {
		service.NodeConfigurationServiceInterface
	}
	mockFileServiceCheckSnapshotIntegrityFilled struct {
		service.FileService
	}
)

func (*mockFileServiceCheckSnapshotIntegrityFilled) GetFileNameFromBytes([]byte) string {
	return ""
}
func (*mockFileServiceCheckSnapshotIntegrityFilled) ReadFileFromDir(string, string) (b []byte, err error) {
	return nil, nil
}
func (*mockNodeConfigurationCheckSnapshotIntegrityFilled) GetHost() *model.Host {
	return &model.Host{
		Info: &model.Node{
			ID: 1234567890,
		},
	}
}

func (*mockSnapshotChunkUtilCheckSnapshotIntegrityFilled) GetShardAssignment([]byte, int, []int64, bool) (storage.ShardMap, error) {
	return storage.ShardMap{
		NodeShards: map[int64][]uint64{
			1234567890: {1, 3},
			1234567891: {3, 4},
		},
		ShardChunks: map[uint64][][]byte{
			1: {
				{1, 23, 4},
			},
		},
	}, nil
}

func (*mockBlockSpinePublicKeysCheckSnapshotIntegrityFilled) GetSpinePublicKeysByBlockHeight(
	uint32,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	return []*model.SpinePublicKey{
		{
			NodeID: 1234567890,
		},
		{
			NodeID: rand.Int63n(10),
		},
	}, nil
}

func (*mockBlockStateStorageCheckSnapshotIntegrityFilled) GetItem(_, item interface{}) error {
	assert := item.(*model.Block)
	*assert = model.Block{Height: 3000}
	return nil
}
func (*mockSpineBlockManifestCheckSnapshotIntegrityFilled) GetSpineBlockManifestsByManifestReferenceHeightRange(
	uint32, uint32,
) (manifests []*model.SpineBlockManifest, err error) {
	fullHashed, fileChunkHashes, err := testfixtures.GetFixtureForSnapshotBasicChunks(constant.SnapshotChunkSize, service.NewFileService(
		logrus.New(),
		new(codec.CborHandle),
		"./testdata",
	), mockSnapshotPayload)
	if err != nil {
		return nil, err
	}
	var fullChunkHashes []byte
	for _, chunkHaHash := range fileChunkHashes {
		fullChunkHashes = append(fullChunkHashes, chunkHaHash...)
	}

	return []*model.SpineBlockManifest{
		{
			FileChunkHashes: fullChunkHashes,
			FullFileHash:    fullHashed,
		},
	}, nil
}

func TestSnapshotScheduler_CheckChunksIntegrity(t *testing.T) {
	type fields struct {
		Logger                     *logrus.Logger
		SpineBlockManifestService  service.SpineBlockManifestServiceInterface
		FileService                service.FileServiceInterface
		SnapshotChunkUtil          util.ChunkUtilInterface
		NodeShardStorage           storage.CacheStorageInterface
		BlockStateStorage          storage.CacheStorageInterface
		BlockCoreService           service.BlockServiceInterface
		BlockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface
		NodeConfigurationService   service.NodeConfigurationServiceInterface
		FileDownloaderService      p2p.FileDownloaderInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Want:NoNeedToDownload",
			fields: fields{
				SpineBlockManifestService:  &mockSpineBlockManifestCheckSnapshotIntegrityFilled{},
				BlockStateStorage:          &mockBlockStateStorageCheckSnapshotIntegrityFilled{},
				BlockSpinePublicKeyService: &mockBlockSpinePublicKeysCheckSnapshotIntegrityFilled{},
				SnapshotChunkUtil:          &mockSnapshotChunkUtilCheckSnapshotIntegrityFilled{},
				NodeConfigurationService:   &mockNodeConfigurationCheckSnapshotIntegrityFilled{},
				FileService:                &mockFileServiceCheckSnapshotIntegrityFilled{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotScheduler{
				SpineBlockManifestService:  tt.fields.SpineBlockManifestService,
				FileService:                tt.fields.FileService,
				SnapshotChunkUtil:          tt.fields.SnapshotChunkUtil,
				NodeShardStorage:           tt.fields.NodeShardStorage,
				BlockStateStorage:          tt.fields.BlockStateStorage,
				BlockCoreService:           tt.fields.BlockCoreService,
				BlockSpinePublicKeyService: tt.fields.BlockSpinePublicKeyService,
				NodeConfigurationService:   tt.fields.NodeConfigurationService,
				FileDownloaderService:      tt.fields.FileDownloaderService,
			}
			if err := ss.CheckChunksIntegrity(); err != nil && !tt.wantErr {
				t.Errorf("CheckChunksIntegrity got err: %s", err.Error())
			}
		})
	}
}

var (
	mockSnapshotPayload = &model.SnapshotPayload{
		Blocks: []*model.Block{transaction.GetFixturesForBlock(720, 1234567890)},
	}
)

type (
	mockBlockCoreServiceDeleteUnmaintainedChunksSuccess struct {
		service.BlockService
	}
	mockSpinePublicKeyServiceDeleteUnmaintainedChunksSuccess struct {
		service.BlockSpinePublicKeyService
	}
	mockSpineBlockManifestServiceDeleteUnmaintainedChunksEmpty struct {
		service.SpineBlockManifestService
	}
	mockSpineBlockManifestServiceDeleteUnmaintainedChunksFilled struct {
		service.SpineBlockManifestService
	}
	mockNodeConfigurationServiceDeleteUnmaintainedChunksSuccess struct {
		service.NodeConfigurationService
	}

	mockNodeShardCacheStorageNotFound struct {
		storage.NodeShardCacheStorage
	}
	mockNodeShardCacheStorageSuccess struct {
		storage.NodeShardCacheStorage
	}
	mockBlockStateStorageEmpty struct {
		storage.BlockStateStorage
	}
	mockBlockStateStorageFilled struct {
		storage.BlockStateStorage
	}
	mockSnapshotChunkUtilSuccess struct {
		util.ChunkUtil
	}
)

func (*mockNodeShardCacheStorageNotFound) GetItem(interface{}, interface{}) error {
	return nil
}

func (*mockNodeShardCacheStorageSuccess) GetItem(interface{}, interface{}) error {
	return errors.New("error needed")
}

func (*mockBlockStateStorageEmpty) GetItem(interface{}, interface{}) error {
	return nil
}
func (*mockBlockStateStorageFilled) GetItem(_, item interface{}) error {
	assert := item.(*model.Block)
	*assert = model.Block{Height: 3000}
	return nil
}

func (*mockSnapshotChunkUtilSuccess) GetShardAssignment([]byte, int, []int64, bool) (storage.ShardMap, error) {
	return storage.ShardMap{
		NodeShards: map[int64][]uint64{
			1234567890: {1, 3},
			1234567891: {3, 4},
		},
		ShardChunks: map[uint64][][]byte{
			1: {
				{1, 23, 4},
			},
		},
	}, nil
}

func (*mockBlockCoreServiceDeleteUnmaintainedChunksSuccess) GetLastBlock() (*model.Block, error) {
	return &model.Block{
		Height: 100,
	}, nil
}

func (*mockSpinePublicKeyServiceDeleteUnmaintainedChunksSuccess) GetSpinePublicKeysByBlockHeight(
	uint32,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	return []*model.SpinePublicKey{
		{
			NodeID: 1234567890,
		},
		{
			NodeID: rand.Int63n(10),
		},
	}, nil
}

func (*mockSpineBlockManifestServiceDeleteUnmaintainedChunksEmpty) GetSpineBlockManifestsByManifestReferenceHeightRange(
	uint32, uint32,
) (manifests []*model.SpineBlockManifest, err error) {
	return nil, nil
}
func (*mockSpineBlockManifestServiceDeleteUnmaintainedChunksFilled) GetSpineBlockManifestsByManifestReferenceHeightRange(
	uint32, uint32,
) (manifests []*model.SpineBlockManifest, err error) {
	fullHashed, fileChunkHashes, err := testfixtures.GetFixtureForSnapshotBasicChunks(constant.SnapshotChunkSize, service.NewFileService(
		logrus.New(),
		new(codec.CborHandle),
		"./testdata",
	), mockSnapshotPayload)
	if err != nil {
		return nil, err
	}
	var fullChunkHashes []byte
	for _, chunkHaHash := range fileChunkHashes {
		fullChunkHashes = append(fullChunkHashes, chunkHaHash...)
	}

	return []*model.SpineBlockManifest{
		{
			FileChunkHashes: fullChunkHashes,
			FullFileHash:    fullHashed,
		},
	}, nil
}
func (*mockNodeConfigurationServiceDeleteUnmaintainedChunksSuccess) GetHost() *model.Host {
	return &model.Host{
		Info: &model.Node{
			ID: 1234567890,
		},
	}
}
func TestSnapshotScheduler_DeleteUnmaintainedChunks(t *testing.T) {
	type fields struct {
		Logger                     *logrus.Logger
		SpineBlockManifestService  service.SpineBlockManifestServiceInterface
		FileService                service.FileServiceInterface
		SnapshotChunkUtil          util.ChunkUtilInterface
		NodeShardStorage           storage.CacheStorageInterface
		BlockStateStorage          storage.CacheStorageInterface
		BlockCoreService           service.BlockServiceInterface
		BlockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface
		NodeConfigurationService   service.NodeConfigurationServiceInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Success:BlockHeightLessThanConstant",
			fields: fields{
				SpineBlockManifestService:  nil,
				FileService:                nil,
				SnapshotChunkUtil:          nil,
				NodeShardStorage:           nil,
				BlockStateStorage:          &mockBlockStateStorageEmpty{},
				BlockCoreService:           nil,
				BlockSpinePublicKeyService: nil,
				NodeConfigurationService:   nil,
			},
		},
		{
			name: "Success:EmptyManifests",
			fields: fields{
				SpineBlockManifestService:  &mockSpineBlockManifestServiceDeleteUnmaintainedChunksEmpty{},
				SnapshotChunkUtil:          util.NewChunkUtil(sha256.Size, &mockNodeShardCacheStorageNotFound{}, logrus.New()),
				NodeShardStorage:           &mockNodeShardCacheStorageNotFound{},
				BlockStateStorage:          &mockBlockStateStorageFilled{},
				BlockCoreService:           &mockBlockCoreServiceDeleteUnmaintainedChunksSuccess{},
				BlockSpinePublicKeyService: &mockSpinePublicKeyServiceDeleteUnmaintainedChunksSuccess{},
				NodeConfigurationService:   &mockNodeConfigurationServiceDeleteUnmaintainedChunksSuccess{},
				FileService: service.NewFileService(
					logrus.New(),
					new(codec.CborHandle),
					"./testdata",
				),
			},
		},
		{
			name: "Success:DeletingFromShard",
			fields: fields{
				SpineBlockManifestService:  &mockSpineBlockManifestServiceDeleteUnmaintainedChunksFilled{},
				SnapshotChunkUtil:          &mockSnapshotChunkUtilSuccess{},
				NodeShardStorage:           &mockNodeShardCacheStorageSuccess{},
				BlockStateStorage:          &mockBlockStateStorageFilled{},
				BlockCoreService:           &mockBlockCoreServiceDeleteUnmaintainedChunksSuccess{},
				BlockSpinePublicKeyService: &mockSpinePublicKeyServiceDeleteUnmaintainedChunksSuccess{},
				NodeConfigurationService:   &mockNodeConfigurationServiceDeleteUnmaintainedChunksSuccess{},
				FileService: service.NewFileService(
					logrus.New(),
					new(codec.CborHandle),
					"./testdata",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotScheduler{
				SpineBlockManifestService:  tt.fields.SpineBlockManifestService,
				FileService:                tt.fields.FileService,
				SnapshotChunkUtil:          tt.fields.SnapshotChunkUtil,
				NodeShardStorage:           tt.fields.NodeShardStorage,
				BlockStateStorage:          tt.fields.BlockStateStorage,
				BlockCoreService:           tt.fields.BlockCoreService,
				BlockSpinePublicKeyService: tt.fields.BlockSpinePublicKeyService,
				NodeConfigurationService:   tt.fields.NodeConfigurationService,
			}
			if err := ss.DeleteUnmaintainedChunks(); (err != nil) != tt.wantErr {
				t.Errorf("DeleteUnmaintainedChunks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

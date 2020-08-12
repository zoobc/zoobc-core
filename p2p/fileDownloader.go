package p2p

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/util"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"golang.org/x/crypto/sha3"
)

type (
	// FileDownloaderInterface snapshot logic shared across block types
	FileDownloaderInterface interface {
		DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) (*model.SnapshotFileInfo, error)
	}

	FileDownloader struct {
		FileService                service.FileServiceInterface
		P2pService                 Peer2PeerServiceInterface
		BlockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface
		BlockchainStatusService    service.BlockchainStatusServiceInterface
		ChunkUtil                  util.ChunkUtilInterface
		Logger                     *log.Logger
	}
)

func NewFileDownloader(
	p2pService Peer2PeerServiceInterface,
	fileService service.FileServiceInterface,
	blockchainStatusService service.BlockchainStatusServiceInterface,
	blockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface,
	chunkUtil util.ChunkUtilInterface,
	logger *log.Logger,
) *FileDownloader {
	return &FileDownloader{
		P2pService:                 p2pService,
		FileService:                fileService,
		BlockSpinePublicKeyService: blockSpinePublicKeyService,
		BlockchainStatusService:    blockchainStatusService,
		ChunkUtil:                  chunkUtil,
		Logger:                     logger,
	}
}

// DownloadSnapshot downloads a snapshot from the p2p network
func (ss *FileDownloader) DownloadSnapshot(
	ct chaintype.ChainType,
	spineBlockManifest *model.SpineBlockManifest,
) (*model.SnapshotFileInfo, error) {
	var (
		failedDownloadChunkNames = model.NewMapStringInt() // map instead of array to avoid duplicates
		hashSize                 = sha3.New256().Size()
		wg                       sync.WaitGroup
		validNodeRegistryIDs     = make(map[int64]bool)
		shardToDownload          [][]string
	)

	fileChunkHashes, err := ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), hashSize)
	if err != nil {
		return nil, err
	}
	if len(fileChunkHashes) == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "Failed parsing File Chunk Hashes from Spine Block Manifest")
	}
	shardMap := ss.ChunkUtil.ShardChunk(spineBlockManifest.GetFileChunkHashes(), constant.ShardBitLength)
	for _, shard := range shardMap {
		temp := make([]string, len(shard))
		for i, chunk := range shard {
			fileName := ss.FileService.GetFileNameFromHash(chunk)
			temp[i] = fileName
		}
		shardToDownload = append(shardToDownload, temp)
	}
	ss.BlockchainStatusService.SetIsDownloadingSnapshot(ct, true)
	// get valid spine public keys from height 0 to manifest reference height (last height after snapshot imported)
	validSpinePublicKeys, err := ss.BlockSpinePublicKeyService.GetValidSpinePublicKeyByBlockHeightInterval(
		0,
		spineBlockManifest.ManifestReferenceHeight,
	)

	if err != nil {
		return nil, err
	}
	// fetching nodeID of the valid node registry at the snapshot height, so we only download snapshot from those peers
	for _, key := range validSpinePublicKeys {
		validNodeRegistryIDs[key.NodeID] = true
	}
	// TODO: implement some sort of rate limiting for number of concurrent downloads (eg. by segmenting the WaitGroup)
	wg.Add(len(shardToDownload))
	for _, shard := range shardToDownload {
		go func(fileChunkHashes []string) {
			defer wg.Done()
			failed, err := ss.P2pService.DownloadFilesFromPeer(
				spineBlockManifest.FullFileHash,
				fileChunkHashes,
				validNodeRegistryIDs,
				constant.DownloadSnapshotNumberOfRetries,
			)
			if err != nil {
				ss.Logger.Error(err)
			}
			if len(failed) > 0 {
				for _, failedFileName := range failed {
					var nInt int64 = 0
					n, ok := failedDownloadChunkNames.Load(failedFileName)
					if ok {
						nInt = n + 1
					}
					failedDownloadChunkNames.Store(failedFileName, nInt)
				}
				return
			}
		}(shard)
	}
	wg.Wait()
	ss.BlockchainStatusService.SetIsDownloadingSnapshot(ct, false)

	if failedDownloadChunkNames.Count() > 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("One or more snapshot chunks failed to download (name/failed times) %v",
			failedDownloadChunkNames.GetMap()))
	}

	return &model.SnapshotFileInfo{
		SnapshotFileHash:           spineBlockManifest.GetFullFileHash(),
		FileChunksHashes:           fileChunkHashes,
		ChainType:                  ct.GetTypeInt(),
		Height:                     spineBlockManifest.ManifestReferenceHeight,
		ProcessExpirationTimestamp: spineBlockManifest.ExpirationTimestamp,
		SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
	}, nil
}

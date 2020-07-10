package p2p

import (
	"fmt"
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
		FileService             service.FileServiceInterface
		P2pService              Peer2PeerServiceInterface
		BlockchainStatusService service.BlockchainStatusServiceInterface
		Logger                  *log.Logger
	}
)

func NewFileDownloader(
	p2pService Peer2PeerServiceInterface,
	fileService service.FileServiceInterface,
	logger *log.Logger,
	blockchainStatusService service.BlockchainStatusServiceInterface,
) *FileDownloader {
	return &FileDownloader{
		P2pService:              p2pService,
		FileService:             fileService,
		Logger:                  logger,
		BlockchainStatusService: blockchainStatusService,
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
	)

	fileChunkHashes, err := ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), hashSize)
	if err != nil {
		return nil, err
	}
	if len(fileChunkHashes) == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "Failed parsing File Chunk Hashes from Spine Block Manifest")
	}

	ss.BlockchainStatusService.SetIsDownloadingSnapshot(ct, true)
	// TODO: implement some sort of rate limiting for number of concurrent downloads (eg. by segmenting the WaitGroup)
	wg.Add(len(fileChunkHashes))
	for _, fileChunkHash := range fileChunkHashes {
		go func(fileChunkHash []byte) {
			defer wg.Done()
			// TODO: for now download just one chunk per peer,
			//  but in future we could download multiple chunks at once from one peer
			fileName := ss.FileService.GetFileNameFromHash(fileChunkHash)
			failed, err := ss.P2pService.DownloadFilesFromPeer(spineBlockManifest.GetFileChunkHashes(), []string{fileName}, constant.DownloadSnapshotNumberOfRetries)
			if err != nil {
				ss.Logger.Error(err)
			}
			if len(failed) > 0 {
				var nInt int64 = 0
				n, ok := failedDownloadChunkNames.Load(fileName)
				if ok {
					nInt = n + 1
				}
				failedDownloadChunkNames.Store(fileName, nInt)
				return
			}
		}(fileChunkHash)
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

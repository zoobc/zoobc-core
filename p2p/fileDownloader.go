package p2p

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"golang.org/x/crypto/sha3"
	"sync"
)

type (
	// FileDownloaderInterface snapshot logic shared across block types
	FileDownloaderInterface interface {
		DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) error
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
func (ss *FileDownloader) DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) error {
	var (
		failedDownloadChunkNames sync.Map // map instead of array to avoid duplicates
		hashSize                 = sha3.New256().Size()
		wg                       sync.WaitGroup
	)
	fileChunkHashes, err := ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), hashSize)
	if err != nil {
		return err
	}
	if len(fileChunkHashes) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "Failed parsing File Chunk Hashes from Spine Block Manifest")
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
			failed, err := ss.P2pService.DownloadFilesFromPeer([]string{fileName}, constant.DownloadSnapshotNumberOfRetries)
			if err != nil {
				ss.Logger.Error(err)
			}
			if len(failed) > 0 {
				n, ok := failedDownloadChunkNames.Load(fileName)
				nInt := 0
				if ok {
					nInt = n.(int) + 1
				}
				failedDownloadChunkNames.Store(fileName, nInt)
				return
			}
		}(fileChunkHash)
	}
	wg.Wait()
	ss.BlockchainStatusService.SetIsDownloadingSnapshot(ct, false)

	// convert sync.Map to a regular map to check its size and print it out in case > 0
	failedDownloadChunkNamesMap := make(map[string]int)
	failedDownloadChunkNames.Range(func(k, v interface{}) bool {
		failedDownloadChunkNamesMap[k.(string)] = v.(int)
		return true
	})
	if len(failedDownloadChunkNamesMap) > 0 {
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("One or more snapshot chunks failed to download (name/failed times) %v",
			failedDownloadChunkNamesMap))
	}
	return nil
}

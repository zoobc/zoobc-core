package p2p

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
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
		failedDownloadChunkNames = make([]string, 0)
		hashSize                 = sha3.New256().Size()
		wg                       sync.WaitGroup
	)
	fileChunkHashes, err := ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), hashSize)
	if err != nil {
		return err
	}

	ss.BlockchainStatusService.SetIsDownloadingSnapshot(ct, true)
	// TODO: implement some sort of rate limiting for number of concurrent downloads (eg. by segmenting the WaitGroup)
	wg.Add(len(fileChunkHashes))
	for _, fileChunkHash := range fileChunkHashes {
		fileName, err := ss.FileService.GetFileNameFromHash(fileChunkHash)
		if err != nil {
			failedDownloadChunkNames = append(failedDownloadChunkNames, fileName)
			wg.Done()
			continue
		}
		go func(fileName string, fileChunkHash []byte) {
			defer wg.Done()
			// TODO: for now download just one chunk per peer,
			//  but in future we could download multiple chunks at once from one peer
			failed, err := ss.P2pService.DownloadFilesFromPeer([]string{fileName})
			if err != nil && failed != nil {
				failedDownloadChunkNames = append(failedDownloadChunkNames, failed...)
				// TODO: implement retry on failed snapshot chunks (eg. try download from a different peer)
				return
			}
		}(fileName, fileChunkHash)
	}
	wg.Wait()
	ss.BlockchainStatusService.SetIsDownloadingSnapshot(ct, false)

	if len(failedDownloadChunkNames) > 0 {
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("One or more snapshot chunks failed to download %v",
			failedDownloadChunkNames))
	}
	return nil
}

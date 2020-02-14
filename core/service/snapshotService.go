package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
	"os"
)

type (
	// SnapshotServiceInterface snapshot logic shared across block types
	SnapshotServiceInterface interface {
		GenerateSnapshot(block *model.Block, ct chaintype.ChainType, chunkSizeBytes int64) (*model.SnapshotFileInfo, error)
		DownloadSnapshot(*model.SnapshotFileInfo) error
		ValidateSnapshotFile(file *os.File, hash []byte) bool
		StartSnapshotListener() observer.Listener
	}

	SnapshotService struct {
		SpineBlockManifestService SpineBlockManifestServiceInterface
		SpineBlockDownloadService SpineBlockDownloadServiceInterface
		SnapshotBlockServices     map[int32]SnapshotBlockServiceInterface // map key = chaintype number (eg. mainchain = 0)
		FileDownloaderService     FileDownloaderServiceInterface
		Logger                    *log.Logger
	}
)

func NewSnapshotService(
	spineBlockManifestService SpineBlockManifestServiceInterface,
	spineBlockDownloadService SpineBlockDownloadServiceInterface,
	snapshotBlockServices map[int32]SnapshotBlockServiceInterface,
	fileDownloaderService FileDownloaderServiceInterface,
	logger *log.Logger,
) *SnapshotService {
	return &SnapshotService{
		SpineBlockManifestService: spineBlockManifestService,
		SpineBlockDownloadService: spineBlockDownloadService,
		SnapshotBlockServices:     snapshotBlockServices,
		FileDownloaderService:     fileDownloaderService,
		Logger:                    logger,
	}
}

// GenerateSnapshot compute and persist a snapshot to file
func (ss *SnapshotService) GenerateSnapshot(block *model.Block, ct chaintype.ChainType,
	snapshotChunkBytesLength int64) (*model.SnapshotFileInfo, error) {

	snapshotBlockService, ok := ss.SnapshotBlockServices[ct.GetTypeInt()]
	if !ok {
		return nil, fmt.Errorf("snapshots for chaintype %s not implemented", ct.GetName())
	}
	return snapshotBlockService.NewSnapshotFile(block, snapshotChunkBytesLength)
}

// StartSnapshotListener setup listener for snapshots generation
func (ss *SnapshotService) StartSnapshotListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args ...interface{}) {
			b := block.(*model.Block)
			ct, ok := args[0].(chaintype.ChainType)
			if !ok {
				ss.Logger.Fatalln("chaintype casting failures in StartSnapshotListener")
			}
			if ct.HasSnapshots() {
				snapshotBlockService, ok := ss.SnapshotBlockServices[ct.GetTypeInt()]
				if !ok {
					ss.Logger.Fatalf("snapshots for chaintype %s not implemented", ct.GetName())
				}
				if snapshotBlockService.IsSnapshotHeight(b.Height) {
					go func() {
						// if spine blocks is downloading, do not generate (or download from other peers) snapshots
						// don't generate snapshots until all spine blocks have been downloaded
						if !ss.SpineBlockDownloadService.IsSpineBlocksDownloadFinished() {
							ss.Logger.Infof("Snapshot at block "+
								"height %d not generated because spine blocks are still downloading",
								b.Height)
							return
						}
						// TODO: implement some sort of process management,
						//  such as controlling if there is another snapshot running before starting to compute a new one (
						//  or compute the new one and kill the one already running...)
						snapshotInfo, err := ss.GenerateSnapshot(b, ct, constant.SnapshotChunkLengthBytes)
						if err != nil {
							ss.Logger.Errorf("Snapshot at block "+
								"height %d terminated with errors %s", b.Height, err)
						}
						_, err = ss.SpineBlockManifestService.CreateSpineBlockManifest(
							snapshotInfo.SnapshotFileHash,
							snapshotInfo.Height,
							snapshotInfo.ProcessExpirationTimestamp,
							snapshotInfo.FileChunksHashes,
							ct,
							model.SpineBlockManifestType_Snapshot,
						)
						if err != nil {
							ss.Logger.Errorf("Cannot create spineBlockManifest at block "+
								"height %d. Error %s", b.Height, err)
						}
						ss.Logger.Infof("Snapshot at main block "+
							"height %d terminated successfully", b.Height)
					}()
				}
			}
		},
	}
}

// ValidateSnapshotFile TODO: implement logic
func (*SnapshotService) ValidateSnapshotFile(file *os.File, hash []byte) bool {
	return true
}

func (ss *SnapshotService) DownloadSnapshot(snapshotFileInfo *model.SnapshotFileInfo) error {
	var (
		failedDownloadChunkNames []string = make([]string, 0)
	)
	for _, fileChunkHash := range snapshotFileInfo.GetFileChunksHashes() {
		fileName, err := ss.FileDownloaderService.GetFileNameFromHash(fileChunkHash)
		if err != nil {
			return err
		}
		if err := ss.FileDownloaderService.DownloadFileByName(fileName, fileChunkHash); err != nil {
			ss.Logger.Errorf("Error Downloading snapshot file chunk. name: %s hash: %v", fileName, fileChunkHash)
			failedDownloadChunkNames = append(failedDownloadChunkNames, fileName)
		}
	}
	if len(failedDownloadChunkNames) > 0 {
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("One or more snapshot chunks failed to download %v",
			failedDownloadChunkNames))
	}
	return nil
}

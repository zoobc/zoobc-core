package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
	"math"
	"time"
)

type (
	// SnapshotServiceInterface snapshot logic shared across block types
	SnapshotServiceInterface interface {
		GenerateSnapshot(block *model.Block, ct chaintype.ChainType, chunkSizeBytes int) (*model.SnapshotFileInfo, error)
		IsSnapshotProcessing(ct chaintype.ChainType) bool
		StopSnapshotGeneration(ct chaintype.ChainType) error
		DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) error
		StartSnapshotListener() observer.Listener
	}

	SnapshotService struct {
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlockStatusServices       map[int32]BlockStatusServiceInterface
		SnapshotBlockServices     map[int32]SnapshotBlockServiceInterface // map key = chaintype number (eg. mainchain = 0)
		FileDownloaderService     FileDownloaderServiceInterface
		FileService               FileServiceInterface
		Logger                    *log.Logger
	}
)

var (
	// this map holds boolean channels to all block types that support snapshots
	stopSnapshotGeneration = make(map[int32]chan bool)
	// this map holds boolean values to all block types that support snapshots
	generatingSnapshot = make(map[int32]bool)
)

func NewSnapshotService(
	spineBlockManifestService SpineBlockManifestServiceInterface,
	blockStatusServices map[int32]BlockStatusServiceInterface,
	snapshotBlockServices map[int32]SnapshotBlockServiceInterface,
	fileDownloaderService FileDownloaderServiceInterface,
	fileService FileServiceInterface,
	logger *log.Logger,
) *SnapshotService {
	return &SnapshotService{
		SpineBlockManifestService: spineBlockManifestService,
		BlockStatusServices:       blockStatusServices,
		SnapshotBlockServices:     snapshotBlockServices,
		FileDownloaderService:     fileDownloaderService,
		FileService:               fileService,
		Logger:                    logger,
	}
}

// GenerateSnapshot compute and persist a snapshot to file
func (ss *SnapshotService) GenerateSnapshot(block *model.Block, ct chaintype.ChainType,
	snapshotChunkBytesLength int) (*model.SnapshotFileInfo, error) {
	stopSnapshotGeneration[ct.GetTypeInt()] = make(chan bool)
	for {
		select {
		case <-stopSnapshotGeneration[ct.GetTypeInt()]:
			ss.Logger.Infof("Snapshot generation for block type %s at height %d has been stopped",
				ct.GetName(), block.GetHeight())
			break
		default:
			snapshotBlockService, ok := ss.SnapshotBlockServices[ct.GetTypeInt()]
			if !ok {
				return nil, fmt.Errorf("snapshots for chaintype %s not implemented", ct.GetName())
			}
			generatingSnapshot[ct.GetTypeInt()] = true
			snapshotInfo, err := snapshotBlockService.NewSnapshotFile(block)
			generatingSnapshot[ct.GetTypeInt()] = false
			return snapshotInfo, err
		}
	}
}

// StopSnapshotGeneration stops current snapshot generation
func (ss *SnapshotService) StopSnapshotGeneration(ct chaintype.ChainType) error {
	if !ss.IsSnapshotProcessing(ct) {
		return blocker.NewBlocker(blocker.AppErr, "No snapshots running: nothing to stop")
	}
	stopSnapshotGeneration[ct.GetTypeInt()] <- true
	// TODO implement error handling for abrupt snapshot termination. for now we just wait a few seconds and return
	time.Sleep(2 * time.Second)
	return nil
}

func (*SnapshotService) IsSnapshotProcessing(ct chaintype.ChainType) bool {
	return generatingSnapshot[ct.GetTypeInt()]
}

// StartSnapshotListener setup listener for snapshots generation
func (ss *SnapshotService) StartSnapshotListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(blockI interface{}, args ...interface{}) {
			block := blockI.(*model.Block)
			ct, ok := args[0].(chaintype.ChainType)
			if !ok {
				ss.Logger.Fatalln("chaintype casting failures in StartSnapshotListener")
			}
			if ct.HasSnapshots() {
				snapshotBlockService, ok := ss.SnapshotBlockServices[ct.GetTypeInt()]
				if !ok {
					ss.Logger.Fatalf("snapshots for chaintype %s not implemented", ct.GetName())
				}
				if snapshotBlockService.IsSnapshotHeight(block.Height) {
					go func() {
						// if spine and main blocks are still downloading, after the node has started,
						// do not generate (or download from other peers) snapshots
						if !ss.BlockStatusServices[0].IsFirstDownloadFinished() && !ss.BlockStatusServices[1].
							IsFirstDownloadFinished() {
							ss.Logger.Infof("Snapshot at block "+
								"height %d not generated because spine blocks are still downloading",
								block.Height)
							return
						}
						// if there is another snapshot running before this, kill the one already running
						if ss.IsSnapshotProcessing(ct) {
							if err := ss.StopSnapshotGeneration(ct); err != nil {
								ss.Logger.Infoln(err)
							}
						}
						snapshotInfo, err := ss.GenerateSnapshot(block, ct, constant.SnapshotChunkSize)
						if err != nil {
							ss.Logger.Errorf("Snapshot at block "+
								"height %d terminated with errors %s", block.Height, err)
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
								"height %d. Error %s", block.Height, err)
						}
						ss.Logger.Infof("Snapshot at main block "+
							"height %d terminated successfully", block.Height)
					}()
				}
			}
		},
	}
}

func (ss *SnapshotService) DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) error {
	var (
		failedDownloadChunkNames = make([]string, 0)
		hashSize                 = sha3.New256().Size()
	)
	fileChunkHashes, err := ss.parseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), hashSize)
	if err != nil {
		return err
	}
	for _, fileChunkHash := range fileChunkHashes {
		fileName, err := ss.FileService.GetFileNameFromHash(fileChunkHash)
		if err != nil {
			return err
		}
		if err := ss.FileDownloaderService.DownloadFileByName(fileName, fileChunkHash); err != nil {
			ss.Logger.Infof("Error Downloading snapshot file chunk. name: %s hash: %v", fileName, fileChunkHash)
			failedDownloadChunkNames = append(failedDownloadChunkNames, fileName)
		}
	}
	// TODO: implement retry on failed snapshot chunks (from a different peer)
	if len(failedDownloadChunkNames) > 0 {
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("One or more snapshot chunks failed to download %v",
			failedDownloadChunkNames))
	}
	return nil
}

func (ss *SnapshotService) parseFileChunkHashes(fileHashes []byte, hashLength int) (fileHashesAry [][]byte, err error) {
	// math.Mod returns the reminder of len(fileHashes)/hashLength
	// we use it to check if the length of fileHashes is a multiple of the single hash's length (32 bytes for sha256)
	if len(fileHashes) < hashLength || math.Mod(float64(len(fileHashes)), float64(hashLength)) > 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "invalid file chunks hashes length")
	}
	for i := 0; i < len(fileHashes); i += hashLength {
		fileHashesAry = append(fileHashesAry, fileHashes[i:i+hashLength])
	}
	return fileHashesAry, nil
}

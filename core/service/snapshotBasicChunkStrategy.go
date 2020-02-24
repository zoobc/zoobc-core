package service

import (
	"bytes"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"path/filepath"
)

type (
	SnapshotBasicChunkStrategy struct {
		// chunk size in bytes
		ChunkSize   int
		FileService FileServiceInterface
	}
)

func NewSnapshotBasicChunkStrategy(
	chunkSize int,
	fileService FileServiceInterface,
) *SnapshotBasicChunkStrategy {
	return &SnapshotBasicChunkStrategy{
		ChunkSize:   chunkSize,
		FileService: fileService,
	}
}

func (ss *SnapshotBasicChunkStrategy) GenerateSnapshotChunks(snapshotPayload *model.SnapshotPayload, filePath string) (fullHash []byte,
	fileChunkHashes [][]byte, err error) {
	// encode the snapshot payload
	b, err := ss.FileService.EncodePayload(snapshotPayload)
	if err != nil {
		return nil, nil, err
	}

	//  snapshot payload full hash (will be used to verify data integrity when assembling downloaded snapshot chunks)
	fullHash = ss.FileService.HashPayload(b)

	fileChunks := util.SplitByteSliceByChunkSize(b, ss.ChunkSize)
	for _, fileChunk := range fileChunks {
		fileChunkHash := ss.FileService.HashPayload(fileChunk)

		fileChunkHashes = append(fileChunkHashes, fileChunkHash)

		fileName, err := ss.FileService.GetFileNameFromHash(fileChunkHash)
		if err != nil {
			return nil, nil, err
		}
		err = ss.FileService.SaveBytesToFile(filePath, fileName, fileChunk)
		if err != nil {
			// try remove saved files if saving a chunk file fails
			if err1 := ss.FileService.DeleteFilesByHash(filePath, fileChunkHashes); err1 != nil {
				return nil, nil, err1
			}
			return nil, nil, err
		}

		// make extra sure that the file created is not corrupted
		filePathName := filepath.Join(filePath, fileName)
		match, err := ss.FileService.VerifyFileHash(filePathName, fileChunkHash)
		if err != nil || !match {
			// try remove saved files if file chunk validation fails
			if err1 := ss.FileService.DeleteFilesByHash(filePath, fileChunkHashes); err1 != nil {
				return nil, nil, err1
			}
			return nil, nil, err
		}
	}
	return fullHash, fileChunkHashes, nil
}

func (ss *SnapshotBasicChunkStrategy) BuildSnapshotFromChunks(fullHash []byte, fileChunkHashes [][]byte,
	filePath string) (*model.SnapshotPayload, error) {
	var (
		snapshotPayload *model.SnapshotPayload
		buffer          = bytes.NewBuffer(make([]byte, 0))
	)

	for _, fileChunkHash := range fileChunkHashes {
		fileName, err := ss.FileService.GetFileNameFromHash(fileChunkHash)
		if err != nil {
			return nil, err
		}
		filePathName := filepath.Join(filePath, fileName)
		chunkBytes, err := ioutil.ReadFile(filePathName)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.AppErr,
				fmt.Sprintf("Cannot read snapshot file chunk from disk: %v", err))
		}
		buffer.Write(chunkBytes)
	}
	b := buffer.Bytes()
	payloadHash := sha3.Sum256(b)
	if !bytes.Equal(payloadHash[:], fullHash) {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			"Snapshot's file payload hash doesn't match with the one in database")
	}
	// decode the snapshot payload
	err := ss.FileService.DecodePayload(b, &snapshotPayload)
	if err != nil {
		return nil, err
	}
	return snapshotPayload, nil
}

package service

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
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

// GenerateSnapshotChunks generates a spliced (multiple file chunks of the same size) snapshot from a SnapshotPayload struct and returns
// encoded snapshot payload's hash and the file chunks' hashes (to be included in a spine block manifest)
func (ss *SnapshotBasicChunkStrategy) GenerateSnapshotChunks(snapshotPayload *model.SnapshotPayload, filePath string) (fullHash []byte,
	fileChunkHashes [][]byte, err error) {
	// encode the snapshot payload
	b, err := ss.FileService.EncodePayload(snapshotPayload)
	if err != nil {
		return nil, nil, err
	}

	//  snapshot payload full hash (will be used to verify data integrity when assembling downloaded snapshot chunks)
	fullHash, err = ss.FileService.HashPayload(b)
	if err != nil {
		return nil, nil, err
	}

	fileChunks := util.SplitByteSliceByChunkSize(b, ss.ChunkSize)
	for _, fileChunk := range fileChunks {
		fileChunkHash, err := ss.FileService.HashPayload(fileChunk)
		if err != nil {
			return nil, nil, err
		}

		fileChunkHashes = append(fileChunkHashes, fileChunkHash)

		fileName := ss.FileService.GetFileNameFromHash(fileChunkHash)
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
		fileBytes, err := ioutil.ReadFile(filePathName)
		if err != nil {
			return nil, nil, err
		}
		if !ss.FileService.VerifyFileChecksum(fileBytes, fileChunkHash) {
			// try remove saved files if file chunk validation fails
			err = ss.FileService.DeleteFilesByHash(filePath, fileChunkHashes)
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, blocker.NewBlocker(blocker.ValidationErr, "InvalidFileHash")
		}
	}
	return fullHash, fileChunkHashes, nil
}

// BuildSnapshotFromChunks rebuilds a whole snapshot file from its file chunks and parses the encoded file into a SnapshotPayload struct
func (ss *SnapshotBasicChunkStrategy) BuildSnapshotFromChunks(fullHash []byte, fileChunkHashes [][]byte,
	filePath string) (*model.SnapshotPayload, error) {
	var (
		snapshotPayload *model.SnapshotPayload
		buffer          = bytes.NewBuffer(make([]byte, 0))
	)

	for _, fileChunkHash := range fileChunkHashes {
		chunkBytes, err := ss.FileService.ReadFileByHash(filePath, fileChunkHash)
		if err != nil {
			return nil, err
		}
		buffer.Write(chunkBytes)
	}
	b := buffer.Bytes()
	payloadHash := sha3.Sum256(b)
	if !bytes.Equal(payloadHash[:], fullHash) {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			"Snapshot file payload hash different from the one in database")
	}
	// decode the snapshot payload
	err := ss.FileService.DecodePayload(b, &snapshotPayload)
	if err != nil {
		return nil, err
	}
	return snapshotPayload, nil
}

// DeleteFileByChunkHashes take in the concatenated file hashes (file name) and delete them.
func (ss *SnapshotBasicChunkStrategy) DeleteFileByChunkHashes(fileChunkHashes []byte, filePath string) error {
	fileChunks := util.SplitByteSliceByChunkSize(fileChunkHashes, ss.ChunkSize)
	err := ss.FileService.DeleteFilesByHash(filePath, fileChunks)
	if err != nil {
		return err
	}
	return nil
}

package service

import (
	"bytes"
	"encoding/base64"

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
func (ss *SnapshotBasicChunkStrategy) GenerateSnapshotChunks(
	snapshotPayload *model.SnapshotPayload,
) (fullHash []byte, fileChunkHashes [][]byte, err error) {
	var (
		encodedPayload []byte
		chunks         [][]byte
	)
	// encode the snapshot payload
	encodedPayload, err = ss.FileService.EncodePayload(snapshotPayload)
	if err != nil {
		return nil, nil, err
	}

	//  snapshot payload full hash (will be used to verify data integrity when assembling downloaded snapshot chunks)
	fullHash, err = ss.FileService.HashPayload(encodedPayload)
	if err != nil {
		return nil, nil, err
	}

	chunks = util.SplitByteSliceByChunkSize(encodedPayload, ss.ChunkSize)
	fileChunkHashes, err = ss.FileService.SaveSnapshotChunks(base64.URLEncoding.EncodeToString(fullHash), chunks)
	if err != nil {
		return nil, nil, err
	}

	return fullHash, fileChunkHashes, nil
}

// BuildSnapshotFromChunks rebuilds a whole snapshot file from its file chunks and parses the encoded file into a SnapshotPayload struct
func (ss *SnapshotBasicChunkStrategy) BuildSnapshotFromChunks(snapshotHash []byte, fileChunkHashes [][]byte) (*model.SnapshotPayload, error) {
	var (
		snapshotPayload *model.SnapshotPayload
		buffer          = bytes.NewBuffer(make([]byte, 0))
	)

	for _, fileChunkHash := range fileChunkHashes {
		chunk, err := ss.FileService.ReadFileFromDir(base64.URLEncoding.EncodeToString(snapshotHash), ss.FileService.GetFileNameFromHash(fileChunkHash))
		if err != nil {
			return nil, err
		}
		buffer.Write(chunk)
	}

	b := buffer.Bytes()
	payloadHash := sha3.Sum256(b)
	if !bytes.Equal(payloadHash[:], snapshotHash) {
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
func (ss *SnapshotBasicChunkStrategy) DeleteFileByChunkHashes(concatenatedFileChunks []byte) error {
	return ss.FileService.DeleteSnapshotDir(string(concatenatedFileChunks))

}

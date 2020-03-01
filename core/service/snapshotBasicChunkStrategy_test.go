package service

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	bcsSnapshotFullHash = []byte{
		189, 123, 189, 67, 77, 99, 212, 229, 139, 70, 138, 166, 32, 117, 190, 42, 156, 137, 6, 216, 156, 116, 20, 182, 211, 178,
		224, 220, 235, 28, 62, 12,
	}
	bcsSnapshotChunk1Hash = []byte{
		1, 1, 1, 249, 145, 71, 241, 88, 208, 4, 80, 132, 88, 43, 189, 93, 19, 104, 255, 61, 177, 177, 223,
		188, 144, 9, 73, 75, 6, 1, 1, 1,
	}
	bcsSnapshotChunk2Hash = []byte{
		2, 2, 2, 249, 145, 71, 241, 88, 208, 4, 80, 132, 88, 43, 189, 93, 19, 104, 255, 61, 177, 177, 223,
		188, 144, 9, 73, 75, 6, 2, 2, 2,
	}
	bcsEncodedPayload = []byte{
		130, 166, 110, 65, 99, 99, 111, 117, 110, 116, 65, 100, 100, 114, 101, 115, 115, 120, 44, 66, 67, 90, 110, 83, 102,
		113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118,
		76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 103, 66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 2, 84, 11,
		228, 0, 107, 66, 108, 111, 99, 107, 72, 101, 105, 103, 104, 116, 1, 102, 76, 97, 116, 101, 115, 116, 245, 106, 80,
		111, 112, 82, 101, 118, 101, 110, 117, 101, 26, 5, 245, 225, 0, 112, 83, 112, 101, 110, 100, 97, 98, 108, 101, 66,
		97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 2, 84, 11, 228, 0, 166, 110, 65, 99, 99, 111, 117, 110, 116, 65, 100, 100,
		114, 101, 115, 115, 120, 44, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75,
		111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 103,
		66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 23, 72, 118, 232, 0, 107, 66, 108, 111, 99, 107, 72, 101, 105, 103,
		104, 116, 1, 102, 76, 97, 116, 101, 115, 116, 245, 106, 80, 111, 112, 82, 101, 118, 101, 110, 117, 101, 26, 5, 245,
		225, 0, 112, 83, 112, 101, 110, 100, 97, 98, 108, 101, 66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 23, 72, 118,
		232, 0,
	}
	bcsEncodedPayloadChunk1 = []byte{
		130, 166, 110, 65, 99, 99, 111, 117, 110, 116, 65, 100, 100, 114, 101, 115, 115, 120, 44, 66, 67, 90, 110, 83, 102,
		113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118,
		76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 103, 66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 2, 84, 11,
		228, 0, 107, 66, 108, 111, 99, 107, 72, 101, 105, 103, 104, 116, 1, 102, 76, 97, 116, 101, 115, 116, 245, 106, 80,
		111, 112, 82, 101, 118, 101, 110, 117, 101, 26, 5, 245, 225, 0, 112, 83, 112, 101, 110, 100, 97, 98, 108, 101, 66,
		97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 2, 84, 11, 228, 0, 166, 110, 65, 99, 99, 111, 117, 110, 116, 65, 100, 100,
	}
	bcsEncodedPayloadChunk2 = []byte{
		114, 101, 115, 115, 120, 44, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75,
		111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 103,
		66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 23, 72, 118, 232, 0, 107, 66, 108, 111, 99, 107, 72, 101, 105, 103,
		104, 116, 1, 102, 76, 97, 116, 101, 115, 116, 245, 106, 80, 111, 112, 82, 101, 118, 101, 110, 117, 101, 26, 5, 245,
		225, 0, 112, 83, 112, 101, 110, 100, 97, 98, 108, 101, 66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 23, 72, 118,
		232, 0,
	}
)

type (
	bcsMockFileService struct {
		FileService
		successEncode              bool
		successGetFileNameFromHash bool
		successSaveBytesToFile     bool
		successVerifyFileChecksum  bool
		integrationTest            bool
	}
)

func (*bcsMockFileService) HashPayload(b []byte) ([]byte, error) {
	return bcsSnapshotFullHash, nil
}

func (mfs *bcsMockFileService) EncodePayload(v interface{}) (b []byte, err error) {
	b = bcsEncodedPayload
	if mfs.successEncode {
		return b, nil
	}
	return nil, errors.New("EncodedPayloadFail")
}

func (mfs *bcsMockFileService) GetFileNameFromHash(fileHash []byte) (string, error) {
	if mfs.successGetFileNameFromHash {
		return "vXu9Q01j1OWLRoqmIHW-KpyJBticdBS207Lg3OscPgyO", nil
	}
	return "", errors.New("GetFileNameFromHashFail")
}

func (mfs *bcsMockFileService) SaveBytesToFile(fileBasePath, fileName string, b []byte) error {
	if mfs.successSaveBytesToFile {
		return nil
	}
	return errors.New("SaveBytesToFileFail")
}

func (mfs *bcsMockFileService) DeleteFilesByHash(filePath string, fileHashes [][]byte) error {
	return nil
}

func (mfs *bcsMockFileService) ReadFileByHash(filePath string, fileHash []byte) ([]byte, error) {
	if bytes.Equal(fileHash, bcsSnapshotChunk1Hash) {
		return bcsEncodedPayloadChunk1, nil
	}
	if bytes.Equal(fileHash, bcsSnapshotChunk2Hash) {
		return bcsEncodedPayloadChunk2, nil
	}
	return bcsEncodedPayload, nil
}

func (mfs *bcsMockFileService) DecodePayload(b []byte, v interface{}) error {
	if mfs.integrationTest {
		realFs := NewFileService(
			log.New(),
			new(codec.CborHandle),
			"testdata/snapshots",
		)
		return realFs.DecodePayload(b, new(interface{}))
	}
	return nil
}

func (mfs *bcsMockFileService) VerifyFileChecksum(fileBytes, hash []byte) bool {
	if mfs.successVerifyFileChecksum {
		return true
	}
	return false
}

func TestSnapshotBasicChunkStrategy_GenerateSnapshotChunks(t *testing.T) {
	type fields struct {
		ChunkSize   int
		FileService FileServiceInterface
	}
	type args struct {
		snapshotPayload *model.SnapshotPayload
		filePath        string
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantFullHash        []byte
		wantFileChunkHashes [][]byte
		wantErr             bool
	}{
		{
			name: "GenerateSnapshotChunks:success-{singleChunk}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              true,
					successGetFileNameFromHash: true,
					successSaveBytesToFile:     true,
					successVerifyFileChecksum:  true,
				},
			},
			args: args{
				filePath:        "testdata/snapshots",
				snapshotPayload: &model.SnapshotPayload{},
			},
			wantFullHash: bcsSnapshotFullHash,
			wantFileChunkHashes: [][]byte{
				bcsSnapshotFullHash,
			},
		},
		{
			name: "GenerateSnapshotChunks:fail-{saveFile}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              true,
					successGetFileNameFromHash: true,
					successSaveBytesToFile:     false,
					successVerifyFileChecksum:  true,
				},
			},
			args: args{
				filePath:        "testdata/snapshots",
				snapshotPayload: &model.SnapshotPayload{},
			},
			wantErr: true,
		},
		{
			name: "GenerateSnapshotChunks:fail-{verifyFile}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              true,
					successGetFileNameFromHash: true,
					successSaveBytesToFile:     true,
					successVerifyFileChecksum:  false,
				},
			},
			args: args{
				filePath:        "testdata/snapshots",
				snapshotPayload: &model.SnapshotPayload{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotBasicChunkStrategy{
				ChunkSize:   tt.fields.ChunkSize,
				FileService: tt.fields.FileService,
			}
			gotFullHash, gotFileChunkHashes, err := ss.GenerateSnapshotChunks(tt.args.snapshotPayload, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotBasicChunkStrategy.GenerateSnapshotChunks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFullHash, tt.wantFullHash) {
				t.Errorf("SnapshotBasicChunkStrategy.GenerateSnapshotChunks() gotFullHash = %v, want %v", gotFullHash,
					tt.wantFullHash)
			}
			if !reflect.DeepEqual(gotFileChunkHashes, tt.wantFileChunkHashes) {
				t.Errorf("SnapshotBasicChunkStrategy.GenerateSnapshotChunks() gotFileChunkHashes = %v, want %v",
					gotFileChunkHashes, tt.wantFileChunkHashes)
			}
		})
	}
}

func TestSnapshotBasicChunkStrategy_BuildSnapshotFromChunks(t *testing.T) {
	type fields struct {
		ChunkSize   int
		FileService FileServiceInterface
	}
	type args struct {
		fullHash        []byte
		fileChunkHashes [][]byte
		filePath        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SnapshotPayload
		wantErr bool
	}{
		{
			name: "BuildSnapshotFromChunks:success",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
				},
			},
			args: args{
				filePath: "testdata/snapshots",
				fileChunkHashes: [][]byte{
					bcsSnapshotChunk1Hash,
					bcsSnapshotChunk2Hash,
				},
				fullHash: bcsSnapshotFullHash,
			},
		},
		{
			name: "BuildSnapshotFromChunks:fail-{invalidFileHash}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
				},
			},
			args: args{
				filePath: "testdata/snapshots",
				fileChunkHashes: [][]byte{
					bcsSnapshotChunk1Hash,
					bcsSnapshotChunk2Hash,
				},
				fullHash: bcsSnapshotChunk1Hash,
			},
			wantErr: true,
		},
		{
			name: "BuildSnapshotFromChunks:success-{decodePayload-integrationTest}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					integrationTest: true,
				},
			},
			args: args{
				filePath: "testdata/snapshots",
				fileChunkHashes: [][]byte{
					bcsSnapshotChunk1Hash,
					bcsSnapshotChunk2Hash,
				},
				fullHash: bcsSnapshotFullHash,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotBasicChunkStrategy{
				ChunkSize:   tt.fields.ChunkSize,
				FileService: tt.fields.FileService,
			}
			got, err := ss.BuildSnapshotFromChunks(tt.args.fullHash, tt.args.fileChunkHashes, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotBasicChunkStrategy.BuildSnapshotFromChunks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotBasicChunkStrategy.BuildSnapshotFromChunks() = %v, want %v", got, tt.want)
			}
		})
	}
}

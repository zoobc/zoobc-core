// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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

	fixtureFullHash = []byte{49, 138, 155, 46, 218, 231, 96, 233, 220, 107, 220, 207, 235, 36, 235, 50, 109, 69, 57, 14, 171, 74, 212, 53,
		226, 209, 255, 217, 108, 158, 103, 212}
	fixtureFileChunkHashes = [][]byte{
		{249, 222, 69, 55, 85, 199, 68, 191, 119, 182, 34, 150, 229, 182, 41, 143, 81, 31, 149, 110, 61, 113, 14, 46, 87, 75, 69, 170, 70,
			57, 78, 94},
		{141, 185, 110, 222, 56, 134, 77, 212, 139, 163, 131, 117, 102, 246, 112, 158, 154, 197, 251, 44, 117, 212, 41, 140, 99, 114, 241,
			181, 199, 108, 212, 51},
		{86, 146, 179, 101, 99, 123, 37, 127, 255, 248, 195, 79, 17, 128, 156, 194, 8, 219, 13, 36, 110, 106, 127, 200, 53, 104, 4, 92,
			203, 85, 188, 42},
		{175, 38, 59, 184, 185, 171, 156, 129, 177, 225, 63, 232, 66, 52, 186, 8, 231, 207, 103, 103, 255, 196, 229, 221, 148, 204, 174,
			206, 46, 6, 150, 157},
	}
	fixtureSnapshotPayload = &model.SnapshotPayload{
		Blocks: []*model.Block{
			{
				ID:                   123456789,
				Timestamp:            98765432210,
				BlockSignature:       []byte{3},
				CumulativeDifficulty: "1",
				PayloadLength:        1,
				TotalAmount:          1000,
				TotalFee:             0,
				TotalCoinBase:        1,
				Version:              0,
			},
			{
				ID:                   234567890,
				Timestamp:            98765432211,
				BlockSignature:       []byte{3},
				CumulativeDifficulty: "1",
				PayloadLength:        1,
				TotalAmount:          1000,
				TotalFee:             0,
				TotalCoinBase:        1,
				Version:              0,
			},
			{
				ID:                   3456789012,
				Timestamp:            98765432212,
				BlockSignature:       []byte{3},
				CumulativeDifficulty: "1",
				PayloadLength:        1,
				TotalAmount:          1000,
				TotalFee:             0,
				TotalCoinBase:        1,
				Version:              0,
			},
		},
	}
)

type (
	bcsMockFileService struct {
		FileService
		successEncode             bool
		successSaveBytesToFile    bool
		successVerifyFileChecksum bool
		integrationTest           bool
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

func (mfs *bcsMockFileService) GetFileNameFromHash(fileHash []byte) string {
	return mfs.GetFileNameFromBytes(fileHash)
}

func (mfs *bcsMockFileService) SaveBytesToFile(fileBasePath, fileName string, b []byte) error {
	if mfs.successSaveBytesToFile {
		return nil
	}
	return errors.New("SaveBytesToFileFail")
}

func (mfs *bcsMockFileService) SaveSnapshotChunks(string, [][]byte) (fileHashes [][]byte, err error) {
	if mfs.successSaveBytesToFile {
		return [][]byte{bcsSnapshotFullHash}, nil
	}
	return nil, errors.New("SaveBytesToFileFail")

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
	return mfs.successVerifyFileChecksum
}

func TestSnapshotBasicChunkStrategy_GenerateSnapshotChunks(t *testing.T) {
	type fields struct {
		ChunkSize   int
		FileService FileServiceInterface
	}
	type args struct {
		snapshotPayload *model.SnapshotPayload
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
				// ChunkSize:   10000000, // 10MB chunks
				ChunkSize:   100,
				FileService: NewFileService(log.New(), new(codec.CborHandle), "testdata/snapshots"),
			},
			args: args{
				snapshotPayload: fixtureSnapshotPayload,
			},
			wantFullHash:        fixtureFullHash,
			wantFileChunkHashes: fixtureFileChunkHashes,
		},
		{
			name: "GenerateSnapshotChunks:fail-{saveFile}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger:       log.New(),
						h:            new(codec.CborHandle),
						snapshotPath: "testdata/snapshots",
					},
					successEncode:             true,
					successSaveBytesToFile:    false,
					successVerifyFileChecksum: true,
				},
			},
			args: args{
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
			gotFullHash, gotFileChunkHashes, err := ss.GenerateSnapshotChunks(tt.args.snapshotPayload)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotBasicChunkStrategy.GenerateSnapshotChunks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFullHash, tt.wantFullHash) {
				t.Errorf("SnapshotBasicChunkStrategy.GenerateSnapshotChunks() gotFullHash = \n%v, want \n%v", gotFullHash,
					tt.wantFullHash)
			}
			if !reflect.DeepEqual(gotFileChunkHashes, tt.wantFileChunkHashes) {
				t.Errorf("SnapshotBasicChunkStrategy.GenerateSnapshotChunks() gotFileChunkHashes = \n%v, want \n%v",
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
				// ChunkSize:   10000000, // 10MB chunks
				ChunkSize:   100,
				FileService: NewFileService(log.New(), new(codec.CborHandle), "testdata/snapshots"),
			},
			args: args{
				filePath:        "testdata/snapshots",
				fileChunkHashes: fixtureFileChunkHashes,
				fullHash:        fixtureFullHash,
			},
			want: fixtureSnapshotPayload,
		},
		{
			name: "BuildSnapshotFromChunks:fail-{invalidFileHash}",
			fields: fields{
				ChunkSize: 10000000, // 10MB chunks
				FileService: &bcsMockFileService{
					FileService: FileService{
						Logger:       log.New(),
						h:            new(codec.CborHandle),
						snapshotPath: "testdata/snapshots",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotBasicChunkStrategy{
				ChunkSize:   tt.fields.ChunkSize,
				FileService: tt.fields.FileService,
			}
			got, err := ss.BuildSnapshotFromChunks(tt.args.fullHash, tt.args.fileChunkHashes)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotBasicChunkStrategy.BuildSnapshotFromChunks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotBasicChunkStrategy.BuildSnapshotFromChunks() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

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

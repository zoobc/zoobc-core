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
package util

import (
	"bytes"
	"crypto/rand"
	"errors"
	"hash"
	"io/ioutil"
	"sort"

	"github.com/zoobc/zoobc-core/common/constant"
)

// SortByteArrays sort a slices array
func SortByteArrays(src [][]byte) {
	sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i], src[j]) < 0 })
}

// ReadTransactionBytes get a slice containing the next nBytes from the buffer
func ReadTransactionBytes(buf *bytes.Buffer, nBytes int) ([]byte, error) {
	// TODO: renaming function, this function is not just use for reading bytes of transaction
	nextBytes := buf.Next(nBytes)
	if len(nextBytes) < nBytes {
		return nil, errors.New("EndOfBufferReached")
	}
	return nextBytes, nil
}

// FeePerByteTransaction use to calculate fee of each bytes transaction
func FeePerByteTransaction(feeTransaction int64, transactionBytes []byte) int64 {
	if len(transactionBytes) != 0 {
		return (feeTransaction * constant.OneFeePerByteTransaction) / int64(len(transactionBytes))
	}
	return feeTransaction * constant.OneFeePerByteTransaction
}

func VerifyFileHash(filePath string, hash []byte, hasher hash.Hash) (bool, error) {
	fc, err := ComputeFileHash(filePath, hasher)
	if err != nil {
		return false, err
	}
	if bytes.Equal(fc, hash) {
		return true, nil
	}
	return false, nil
}

func ComputeFileHash(filePath string, hasher hash.Hash) ([]byte, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	_, err = hasher.Write(b)
	if err != nil {
		return nil, err
	}
	return hasher.Sum([]byte{}), nil
}

// SplitByteSliceByChunkSize split a byte slice into multiple chunks of equal size,
// beside the last chunk which could be shorter than the others, if the original slice's length is not multiple of chunkSize
func SplitByteSliceByChunkSize(b []byte, chunkSize int) (splitSlice [][]byte) {
	for i := 0; i < len(b); i += chunkSize {
		end := i + chunkSize
		if end > len(b) {
			end = len(b)
		}
		splitSlice = append(splitSlice, b[i:end])
	}
	return splitSlice
}

// GetChecksumByte Calculate a checksum byte from a collection of bytes
// checksum 255 = 255, 256 = 0, 257 = 1 and so on.
func GetChecksumByte(bytes []byte) byte {
	n := len(bytes)
	var a byte
	for i := 0; i < n; i++ {
		a += bytes[i]
	}
	return a
}

// GenerateRandomBytes returns securely generated random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

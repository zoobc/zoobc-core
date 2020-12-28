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
	"reflect"
	"testing"
)

func TestAllMerkle(t *testing.T) {
	t.Run("functional-and-integration:success", func(t *testing.T) {
		merkle := MerkleRoot{}
		var hashes []*bytes.Buffer
		for i := 0; i < 8; i++ {
			var random = make([]byte, 32)
			_, _ = rand.Read(random)
			hashes = append(hashes, bytes.NewBuffer(random))
		}
		result, err := merkle.GenerateMerkleRoot(hashes)
		if err != nil {
			t.Errorf("error occurred when generating merkle root: %v", err)
		}
		// flatten the root and tree for database representation
		root, tree := merkle.ToBytes()
		if !reflect.DeepEqual(result.Bytes(), root) {
			t.Error("merkle root from generateMerkleRoot function differ from flatten root")
		}
		// restore tree state from flatten bytes in database
		hashTreeFromDb := merkle.FromBytes(tree, root)
		if !reflect.DeepEqual(merkle.HashTree, hashTreeFromDb) {
			t.Error("hash tree from flatten and build is different")
		}
		// verify every leaf behavior
		for index, leaf := range hashes {
			var normalizedIntermediateHashes [][]byte
			intermediateHashes := merkle.GetIntermediateHashes(leaf, int32(index))
			for _, ih := range intermediateHashes {
				normalizedIntermediateHashes = append(normalizedIntermediateHashes, ih.Bytes())
			}
			calculatedRoot, _ := merkle.GetMerkleRootFromIntermediateHashes(
				leaf.Bytes(), uint32(index), normalizedIntermediateHashes,
			)
			if !reflect.DeepEqual(calculatedRoot, result.Bytes()) {
				t.Error("calculated root differ from generated root")
			}
			flatenIntermediateHash := merkle.FlattenIntermediateHashes(normalizedIntermediateHashes)
			recoveredIntermediateHash := merkle.RestoreIntermediateHashes(flatenIntermediateHash)
			if !reflect.DeepEqual(recoveredIntermediateHash, normalizedIntermediateHashes) {
				t.Error("merkle tree from flatten bytes does not build the same tree")
			}
		}
	})
	t.Run("odd number of elements", func(t *testing.T) {
		merkle := MerkleRoot{}
		_, err := merkle.GenerateMerkleRoot([]*bytes.Buffer{
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
		})
		if err != nil {
			t.Error("any element should be handled")
		}
	})
	t.Run("non power of 2 even number of elements", func(t *testing.T) {
		merkle := MerkleRoot{}
		_, err := merkle.GenerateMerkleRoot([]*bytes.Buffer{
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
		})
		if err != nil {
			t.Error("any element should be handled")
		}
	})
}

func BenchmarkMerkleTree8(b *testing.B) {
	merkle := MerkleRoot{}
	var hashesData = []*bytes.Buffer{}

	for n := 0; n < 1000000; n++ {

		for i := 0; i < 8; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}
		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

func BenchmarkMerkleTree32(b *testing.B) {
	merkle := MerkleRoot{}
	for n := 0; n < 1000000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 32; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

func BenchmarkMerkleTree64(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 1000000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 64; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

func BenchmarkMerkleTree128(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 1000000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 128; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

func BenchmarkMerkleTree256(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 1000000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 256; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

func BenchmarkMerkleTree512(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 1000000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 512; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}
		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

func BenchmarkMerkleTree1024(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 1000000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 1024; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}
		_, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
	}
}

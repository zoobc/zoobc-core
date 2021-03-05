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
	"math"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"

	"golang.org/x/crypto/sha3"
)

type MerkleRootInterface interface {
	GenerateMerkleRoot(items []*bytes.Buffer) (*bytes.Buffer, error)
	GetMerkleRootFromIntermediateHashes(
		leaf []byte, leafIndex uint32,
		intermediateHashes [][]byte,
	) (root []byte, err error)
	GetIntermediateHashes(leafHash *bytes.Buffer, leafIndex int32) []*bytes.Buffer
	FlattenIntermediateHashes(intermediateHashes [][]byte) []byte
	RestoreIntermediateHashes(flattenIntermediateHashes []byte) [][]byte
	ToBytes() (root, tree []byte)
	FromBytes(tree, root []byte) [][]*bytes.Buffer
}

type MerkleRoot struct {
	// HashTree store the whole tree, only filled after calling `GenerateMerkleRoot`
	HashTree [][]*bytes.Buffer
}

func NewMerkleRoot() MerkleRootInterface {
	return &MerkleRoot{}
}

// GenerateMerkleRoot generate the root of merkle and build the tree in MerkleRoot.HashTree
// return only the root
func (mr *MerkleRoot) GenerateMerkleRoot(items []*bytes.Buffer) (*bytes.Buffer, error) {
	if len(items) == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "LeafOfMerkleRequired")
	}
	treeLevelLength := math.Log2(float64(len(items)))
	if treeLevelLength != math.Floor(treeLevelLength) {
		// find `n` of lacking element and append until condition fulfilled
		nearestBottom := math.Floor(treeLevelLength)
		targetElementLength := math.Pow(2, nearestBottom+1)
		neededElements := int(targetElementLength) - len(items)
		duplicateLastElement := items[len(items)-1]
		for i := 0; i < neededElements; i++ {
			items = append(items, duplicateLastElement)
		}
		treeLevelLength = nearestBottom + 1 // added another level with duplicated elements
	}
	treeLevelLength++ // extra level for the root
	mr.HashTree = make([][]*bytes.Buffer, int(treeLevelLength))
	mr.HashTree[0] = items
	result := mr.merkle(items)
	return result, nil
}

// merkle take slice of leaf node hashes and recursively build the merkle root
func (mr *MerkleRoot) merkle(items []*bytes.Buffer) *bytes.Buffer {
	itemLength := len(items)
	if itemLength == 1 {
		return items[0]
	}
	return mr.hash(
		mr.merkle(items[:itemLength/2]), mr.merkle(items[itemLength/2:]),
		int32(math.Log2(float64(itemLength))),
	)
}

// hash function take the 2 data to be hashed for building merkle tree
func (mr *MerkleRoot) hash(a, b *bytes.Buffer, level int32) *bytes.Buffer {
	digest := sha3.New256()
	_, _ = digest.Write(a.Bytes())
	_, _ = digest.Write(b.Bytes())
	res := bytes.NewBuffer(digest.Sum([]byte{}))
	mr.HashTree[level] = append(mr.HashTree[level], res)
	return res
}

// GetMerkleRootFromIntermediateHashes hash the root to every intermediate hashes in order until it returns the
// merkle root hash
func (*MerkleRoot) GetMerkleRootFromIntermediateHashes(
	leaf []byte, leafIndex uint32,
	intermediateHashes [][]byte,
) (root []byte, err error) {
	digest := sha3.New256()
	lastHash := leaf
	for _, nh := range intermediateHashes {
		digest.Reset()
		if (leafIndex+1)%2 == 0 {
			// right
			_, err = digest.Write(nh)
			if err != nil {
				return nil, err
			}
			_, err = digest.Write(lastHash)
			if err != nil {
				return nil, err
			}
		} else {
			// left
			_, err = digest.Write(lastHash)
			if err != nil {
				return nil, err
			}
			_, err = digest.Write(nh)
			if err != nil {
				return nil, err
			}
		}
		lastHash = digest.Sum([]byte{})
		leafIndex = uint32(math.Ceil(float64(leafIndex+1)/2)) - 1
	}
	return lastHash, nil
}

// GetIntermediateHashes crawl the hashes that are needed to verify the `leafHash`
// leafIndex is index of the leaf node passed, it should be stored to avoid `n` complexity just for finding level 0
// node hash
func (mr *MerkleRoot) GetIntermediateHashes(leafHash *bytes.Buffer, leafIndex int32) []*bytes.Buffer {
	var (
		lastParentHashIndex int
		necessaryHashes     []*bytes.Buffer
	)
	for j := 0; j < len(mr.HashTree)-1; j++ {
		if j == 0 {
			if bytes.Equal(leafHash.Bytes(), mr.HashTree[j][leafIndex].Bytes()) {
				if (leafIndex+1)%2 == 0 {
					necessaryHashes = append(necessaryHashes, mr.HashTree[j][leafIndex-1])
				} else {
					necessaryHashes = append(necessaryHashes, mr.HashTree[j][leafIndex+1])
				}
				lastParentHashIndex = int(math.Ceil(float64(leafIndex+1)/2)) - 1
				continue
			}
		} else {
			if (lastParentHashIndex+1)%2 == 0 {
				necessaryHashes = append(necessaryHashes, mr.HashTree[j][lastParentHashIndex-1])
			} else {
				necessaryHashes = append(necessaryHashes, mr.HashTree[j][lastParentHashIndex+1])
			}
			lastParentHashIndex = int(math.Ceil(float64(lastParentHashIndex+1)/2)) - 1
		}

	}
	return necessaryHashes
}

// IntermediateHashToByte flatten intermediate hashes bytes
func (*MerkleRoot) FlattenIntermediateHashes(intermediateHashes [][]byte) []byte {
	var result []byte
	for _, ih := range intermediateHashes {
		result = append(result, ih...)
	}
	return result
}

func (*MerkleRoot) RestoreIntermediateHashes(flattenIntermediateHashes []byte) [][]byte {
	var (
		result [][]byte
	)
	intermediateHashesSize := len(flattenIntermediateHashes) / constant.ReceiptHashSize
	for i := 0; i < intermediateHashesSize; i++ {
		result = append(result, flattenIntermediateHashes[i*constant.ReceiptHashSize:(i+1)*constant.ReceiptHashSize])
	}
	return result
}

// ToBytes build []byte from HashTree which is a [][]*bytes.Buffer
func (mr *MerkleRoot) ToBytes() (root, tree []byte) {
	var (
		r, t *bytes.Buffer
	)
	t = bytes.NewBuffer([]byte{})
	r = bytes.NewBuffer([]byte{})

	for k, buffer := range mr.HashTree {
		if k+1 == len(mr.HashTree) {
			r.Write(buffer[0].Bytes()) // write root
		} else {
			for _, nestBuf := range buffer {
				t.Write(nestBuf.Bytes())
			}
		}
	}
	return r.Bytes(), t.Bytes()
}

// FromBytes build []byte to [][]*bytes.Buffer tree representation for easier validation
func (mr *MerkleRoot) FromBytes(tree, root []byte) [][]*bytes.Buffer {
	var hashTree [][]*bytes.Buffer
	// 2n-1 of the tree
	treeLevelZeroLength := ((len(tree) / constant.ReceiptHashSize) + 2) / 2
	var offset int
	for treeLevelZeroLength != 1 {
		var tempHashes []*bytes.Buffer
		limit := offset + treeLevelZeroLength
		for i := offset; i < limit; i++ {
			bytesOffset := i * constant.ReceiptHashSize
			bytesLimit := bytesOffset + constant.ReceiptHashSize
			tempHashes = append(tempHashes, bytes.NewBuffer(tree[bytesOffset:bytesLimit]))
		}
		offset += treeLevelZeroLength
		treeLevelZeroLength /= 2
		hashTree = append(hashTree, tempHashes)
	}
	hashTree = append(hashTree, []*bytes.Buffer{
		bytes.NewBuffer(root),
	})
	return hashTree
}

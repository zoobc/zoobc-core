package util

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/blocker"
	"golang.org/x/crypto/sha3"
	"math"
	"reflect"
)

type MerkleRoot struct {
	// HashTree store the whole tree, only filled after calling `GenerateMerkleRoot`
	HashTree [][]*bytes.Buffer
}

// GenerateMerkleRoot generate the root of merkle and build the tree in MerkleRoot.HashTree
// return only the root
func (mr *MerkleRoot) GenerateMerkleRoot(items []*bytes.Buffer) (*bytes.Buffer, error) {
	treeLevelLength := math.Log2(float64(len(items))) + 1
	if treeLevelLength != float64(int64(treeLevelLength)) {
		return nil, blocker.NewBlocker(
			blocker.ValidationErr,
			"wrong element length, it should be power of two",
		)
	}
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
	return mr.hash(mr.merkle(items[:itemLength/2]), mr.merkle(items[itemLength/2:]),
		int32(math.Log2(float64(itemLength))))
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

// GetNecessaryHashes crawl the hashes that are needed to verify the `leafHash`
// leafIndex is index of the leaf node passed, it should be stored to avoid `n` complexity just for finding level 0
// node hash
func (mr *MerkleRoot) GetNecessaryHashes(leafHash *bytes.Buffer, leafIndex int32) []*bytes.Buffer {
	var (
		lastParentHashIndex int
		necessaryHashes     []*bytes.Buffer
	)
	for j := 0; j < len(mr.HashTree)-1; j++ {
		if j == 0 {
			if reflect.DeepEqual(leafHash.Bytes(), mr.HashTree[j][leafIndex].Bytes()) {
				if (leafIndex+1)%2 == 0 {
					necessaryHashes = append(necessaryHashes, mr.HashTree[j][leafIndex-1])
				} else {
					necessaryHashes = append(necessaryHashes, mr.HashTree[j][leafIndex+1])
				}
				lastParentHashIndex = int(math.Ceil(float64(leafIndex) / 2))
				continue
			}
		} else {
			if (lastParentHashIndex+1)%2 == 0 {
				necessaryHashes = append(necessaryHashes, mr.HashTree[j][lastParentHashIndex-1])
			} else {
				necessaryHashes = append(necessaryHashes, mr.HashTree[j][lastParentHashIndex+1])
			}
			lastParentHashIndex = int(math.Ceil(float64(lastParentHashIndex) / 2))
		}

	}
	return necessaryHashes
}

// VerifyLeaf take a leaf hash and the merkle root to verify if the leaf hash, hashed with every hash
// in the necessaryHashes will match the merkle root or not.
func (*MerkleRoot) VerifyLeaf(leaf, root *bytes.Buffer, necessaryHashes []*bytes.Buffer) bool {
	digest := sha3.New256()
	lastHash := leaf.Bytes()
	for _, nh := range necessaryHashes {
		digest.Reset()
		_, _ = digest.Write(lastHash)
		_, _ = digest.Write(nh.Bytes())
		lastHash = digest.Sum([]byte{})
	}
	return reflect.DeepEqual(lastHash, root.Bytes())
}

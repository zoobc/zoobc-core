// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"math"

// 	"github.com/zoobc/zoobc-core/common/blocker"
// 	"golang.org/x/crypto/sha3"
// )

// func main() {
// 	merkleRoot := &MerkleRoot{}
// 	var (
// 		hashedMainBlock []*bytes.Buffer
// 	)
// 	hashedMainBlock = append(hashedMainBlock, bytes.NewBuffer([]byte{227, 188, 216, 182, 94, 120, 80, 100, 253, 57, 243, 67, 23, 249, 54, 19, 40, 208, 11, 106, 129, 160, 183, 142, 47, 133, 210, 91, 145, 186, 170, 140}))
// 	_, err := merkleRoot.GenerateMerkleRoot(hashedMainBlock)
// 	if err != nil {
// 		fmt.Println("Error", err)
// 	}
// 	mRoot, mTree := merkleRoot.ToBytes()
// 	fmt.Println("Hello, playground, ")
// 	fmt.Println(mRoot)
// 	fmt.Println(mTree)
// }

// type MerkleRoot struct {
// 	// HashTree store the whole tree, only filled after calling `GenerateMerkleRoot`
// 	HashTree [][]*bytes.Buffer
// }

// // GenerateMerkleRoot generate the root of merkle and build the tree in MerkleRoot.HashTree
// // return only the root
// func (mr *MerkleRoot) GenerateMerkleRoot(items []*bytes.Buffer) (*bytes.Buffer, error) {
// 	if len(items) == 0 {
// 		return nil, blocker.NewBlocker(blocker.ValidationErr, "LeafOfMerkleRequired")
// 	}
// 	treeLevelLength := math.Log2(float64(len(items)))
// 	if treeLevelLength != math.Floor(treeLevelLength) {
// 		// find `n` of lacking element and append until condition fulfilled
// 		nearestBottom := math.Floor(treeLevelLength)
// 		targetElementLength := math.Pow(2, nearestBottom+1)
// 		neededElements := int(targetElementLength) - len(items)
// 		duplicateLastElement := items[len(items)-1]
// 		for i := 0; i < neededElements; i++ {
// 			items = append(items, duplicateLastElement)
// 		}
// 		treeLevelLength = nearestBottom + 1 // added another level with duplicated elements
// 	}
// 	treeLevelLength++ // extra level for the root
// 	mr.HashTree = make([][]*bytes.Buffer, int(treeLevelLength))
// 	mr.HashTree[0] = items
// 	result := mr.merkle(items)
// 	return result, nil
// }

// // merkle take slice of leaf node hashes and recursively build the merkle root
// func (mr *MerkleRoot) merkle(items []*bytes.Buffer) *bytes.Buffer {
// 	itemLength := len(items)
// 	if itemLength == 1 {
// 		return items[0]
// 	}
// 	return mr.hash(
// 		mr.merkle(items[:itemLength/2]), mr.merkle(items[itemLength/2:]),
// 		int32(math.Log2(float64(itemLength))),
// 	)
// }

// // hash function take the 2 data to be hashed for building merkle tree
// func (mr *MerkleRoot) hash(a, b *bytes.Buffer, level int32) *bytes.Buffer {
// 	digest := sha3.New256()
// 	_, _ = digest.Write(a.Bytes())
// 	_, _ = digest.Write(b.Bytes())
// 	res := bytes.NewBuffer(digest.Sum([]byte{}))
// 	mr.HashTree[level] = append(mr.HashTree[level], res)
// 	return res
// }

// // ToBytes build []byte from HashTree which is a [][]*bytes.Buffer
// func (mr *MerkleRoot) ToBytes() (root, tree []byte) {
// 	var (
// 		r, t *bytes.Buffer
// 	)
// 	t = bytes.NewBuffer([]byte{})
// 	r = bytes.NewBuffer([]byte{})

// 	for k, buffer := range mr.HashTree {
// 		if k+1 == len(mr.HashTree) {
// 			r.Write(buffer[0].Bytes()) // write root
// 		} else {
// 			for _, nestBuf := range buffer {
// 				t.Write(nestBuf.Bytes())
// 			}
// 		}
// 	}
// 	return r.Bytes(), t.Bytes()
// }

package main

import (
	"bytes"
	"encoding/hex"

	"github.com/zoobc/zoobc-core/common/accounttype"
)

func main() {
	hexPubKey, _ := hex.DecodeString("000000002f714c0ee3b1469902bd56f1fd444c32e0fdeae880eb7718c34fc4694a4aed55")
	accType, _ := accounttype.ParseBytesToAccountType(bytes.NewBuffer(hexPubKey))
	encodedAddress, _ := accType.GetEncodedAddress()
	println(encodedAddress)
}

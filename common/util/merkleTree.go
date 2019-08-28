package util

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type opt struct {
	dataPiece string
	Hash      []byte
}

type intermed struct {
	Hash []byte
}

type node struct {
	Hash []byte
}

type root struct {
	Hash []byte
}

func markleTree(c []opt) (leafHash, intermedHash, nodeHash, rootHash []byte) {
	var intermeds []intermed
	var nodes []node
	var roots []root
	var rootsOdd []root
	var pLength, n int

	n = len(c)
	k := n % 2

	for g := 0; g < 4; g++ {

		switch g {
		case 0:
			pLength = n
		case 1:
			pLength = n / 2
		case 2:
			pLength = (n / 2) / 2
		case 3:
			pLength = ((n / 2) / 2) / 2
		}

		var x1, y1, v1 int
		for i := 0; i < pLength; i++ {
			if i != 0 {
				x1 = v1 + 1
				y1 = x1 + 1
			} else {
				x1 = 0
				y1 = x1 + 1
			}

			dataBuffer := sha256.New()
			switch g {
			case 0:
				_, err := dataBuffer.Write([]byte(c[i].dataPiece))
				if err != nil {
					fmt.Printf("failed on leaf \n")
				}
				c[i].Hash = dataBuffer.Sum(nil)

			case 1:
				_, err := dataBuffer.Write(append(c[x1].Hash, c[y1].Hash...))
				if err != nil {
					fmt.Printf("failed on intermed \n")
				}
				intermeds = append(intermeds, intermed{dataBuffer.Sum(nil)})

			case 2:
				_, err := dataBuffer.Write(append(intermeds[x1].Hash, intermeds[y1].Hash...))
				if err != nil {
					fmt.Printf("failed on node \n")
				}
				nodes = append(nodes, node{dataBuffer.Sum(nil)})

			case 3:
				_, err := dataBuffer.Write(append(nodes[x1].Hash, nodes[y1].Hash...))
				if err != nil {
					fmt.Printf("failed on root \n")
				}
				roots = append(roots, root{dataBuffer.Sum(nil)})
				if k != 0 {
					_, err := dataBuffer.Write(append(roots[0].Hash, c[len(c)-1].dataPiece...))
					if err != nil {
						fmt.Printf("failed on root odd \n")
					}
					rootsOdd = append(rootsOdd, root{dataBuffer.Sum(nil)})
				}

			}

			v1 = x1 + 1
		}
	}
	if k != 0 {

		return c[1].Hash, intermeds[1].Hash, nodes[1].Hash, rootsOdd[0].Hash
	}
	return c[1].Hash, intermeds[1].Hash, nodes[1].Hash, roots[0].Hash
}

// VerifyTree validates the hashes
func VerifyTree(content string, leafHash, intermedHash, nodeHash, merkleRoot []byte) (bool, error) {

	hLeaf := sha256.New()
	_, err := hLeaf.Write([]byte(content))
	if err != nil {
		fmt.Printf("failed on leaf content \n")
	}
	contentHash := hLeaf.Sum(nil)

	hIntermed := sha256.New()
	_, err = hIntermed.Write(append(contentHash, leafHash...))
	if err != nil {
		fmt.Printf("failed on intermed \n")
	}
	intermedHashLeft := hIntermed.Sum(nil)

	hNode := sha256.New()
	_, err = hNode.Write(append(intermedHashLeft, intermedHash...))
	if err != nil {
		fmt.Printf("failed on node \n")
	}
	nodeHashLeft := hNode.Sum(nil)

	hRoot := sha256.New()
	_, err = hRoot.Write(append(nodeHashLeft, nodeHash...))
	if err != nil {
		fmt.Printf("failed on root \n")
	}
	rootHash := hRoot.Sum(nil)

	if bytes.Equal(merkleRoot, rootHash) {
		return true, nil
	}
	return false, nil
}

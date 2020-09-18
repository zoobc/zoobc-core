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

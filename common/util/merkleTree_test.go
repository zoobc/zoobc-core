package util

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestAllMerkle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		merkle := MerkleRoot{}
		hashes := []*bytes.Buffer{
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			bytes.NewBuffer([]byte{8, 7, 6, 5, 4, 3, 2, 1}),
			bytes.NewBuffer([]byte{1, 2, 3, 4}),
			bytes.NewBuffer([]byte{4, 3, 2, 1}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6}),
			bytes.NewBuffer([]byte{6, 5, 4, 3, 2, 1}),
			bytes.NewBuffer([]byte{1, 1, 2, 2, 3, 3, 4, 4}),
			bytes.NewBuffer([]byte{4, 4, 3, 3, 2, 2, 1, 1}),
		}
		result, err := merkle.GenerateMerkleRoot(hashes)
		if err != nil {
			t.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}), 0)
		verRes := merkle.VerifyLeaf(bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}), result, nH)
		if !verRes {
			t.Errorf("nh: %v\nresult: %v\nverres: %v", nH, base64.StdEncoding.EncodeToString(result.Bytes()), verRes)
		}
	})
	t.Run("fail:validation", func(t *testing.T) {
		merkle := MerkleRoot{}
		_, err := merkle.GenerateMerkleRoot([]*bytes.Buffer{
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
		})
		if err == nil {
			t.Error("1 element should return error")
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

func BenchmarkMerkleTreeValidation8(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 8; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
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

func BenchmarkMerkleTreeValidation32(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 32; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
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

func BenchmarkMerkleTreeValidation64(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 64; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
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

func BenchmarkMerkleTreeValidation128(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 128; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
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

func BenchmarkMerkleTreeValidation256(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {

		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 256; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}
		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
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

func BenchmarkMerkleTreeValidation512(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {
		var hashesData = []*bytes.Buffer{}
		for i := 0; i < 512; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
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

func BenchmarkMerkleTreeValidation1024(b *testing.B) {
	merkle := MerkleRoot{}

	for n := 0; n < 50000; n++ {
		var hashesData = []*bytes.Buffer{}

		for i := 0; i < 1024; i++ {
			dataRand := make([]byte, 32)
			_, err := rand.Read(dataRand)
			if err != nil {
				b.Errorf("error occurred random func ")
			}
			hashesData = append(hashesData, bytes.NewBuffer(dataRand))
		}

		result, err := merkle.GenerateMerkleRoot(hashesData)
		if err != nil {
			b.Errorf("error occurred when generating merkle root: %v", err)
		}
		nH := merkle.GetIntermediateHashes(hashesData[0], 0)
		merkle.VerifyLeaf(bytes.NewBuffer(hashesData[0].Bytes()), result, nH)
	}
}

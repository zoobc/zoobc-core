package util

import (
	"bytes"
	"testing"
)

var c = []opt{
	{
		dataPiece: "3e55352e649cb04ce701b6046d27163cf849a9fe73396ff072036ad7d0186ee3",
	},
	{
		dataPiece: "28789ecbad8bfe7fb0eeabebbfd6bb96f1eea48ac63f8e80a7d039b52d79f85a",
	},
	{
		dataPiece: "463a2de540ea69dfe8bfc9201eec511e4fcaeb5dafbbc6167c773371d206febb",
	},
	{
		dataPiece: "f1aa5cf0f1f28e8bd46d31b45f48cb3dbb405cffb8533a3647210982618a8e57",
	},
	{
		dataPiece: "db4b0f506b85c867e995c91ec9ac3a77cae374289208452a58838152dc3502cf",
	},
	{
		dataPiece: "2d46429f02dc81e2187031af9070c21570efc26b4a5b127c04774f5a6acc043f",
	},
	{
		dataPiece: "0ee15feb5d7b018d43d3cf8c918f388da3bddb5bd3d2c71e920e2cd144e0aa31",
	},
	{
		dataPiece: "73f3573e063a936e7fd3f410fe36c5a52bac1a7b113155b8c6b332626d41aa61",
	},
}

var expectedHash = []byte{149, 74, 143, 23, 72, 195, 94, 212, 140, 54, 97, 194, 43, 162, 124, 125, 214,
	225, 249, 167, 124, 188, 40, 189, 156, 163, 117, 68, 122, 133, 236, 132}

type fData struct {
	content      string
	leafHash     []byte
	intermedHash []byte
	nodeHash     []byte
	expectedHash []byte
}

var vData = &fData{
	content: "3e55352e649cb04ce701b6046d27163cf849a9fe73396ff072036ad7d0186ee3",
	leafHash: []byte{16, 29, 205, 124, 108, 80, 67, 82, 149, 167, 94, 250, 86, 150,
		5, 163, 28, 70, 166, 8, 217, 21, 49, 78, 186, 52, 119, 175, 128, 57, 232, 185},
	intermedHash: []byte{42, 253, 250, 99, 180, 185, 41, 97, 80, 134, 35, 38, 237, 18,
		96, 42, 118, 164, 51, 31, 195, 4, 178, 127, 209, 147, 216, 23, 174, 30, 32, 17},
	nodeHash: []byte{143, 55, 89, 239, 75, 150, 215, 28, 202, 189, 16, 251, 113, 219, 87, 183,
		164, 29, 101, 108, 252, 234, 47, 128, 226, 247, 25, 5, 39, 77, 233, 170},
	expectedHash: []byte{149, 74, 143, 23, 72, 195, 94, 212, 140, 54, 97, 194, 43, 162, 124, 125, 214,
		225, 249, 167, 124, 188, 40, 189, 156, 163, 117, 68, 122, 133, 236, 132},
}

func Test_MerkleRoot(t *testing.T) {

	_, _, _, x := markleTree(c)
	if !bytes.Equal(x, expectedHash) {
		t.Errorf("error: expected hash equal to %v got %v", expectedHash, x)
	}

}

func BenchmarkMerkleTree(b *testing.B) {
	// run the merkle tree bench
	for n := 0; n < 50000000; n++ {
		markleTree(c)
	}
}

func Test_VerifyMerkle(t *testing.T) {

	z, err := VerifyTree(vData.content, vData.leafHash, vData.intermedHash, vData.nodeHash, vData.expectedHash)
	if err != nil {
		t.Error("error: unexpected error:  ", err)
	}
	if z != true {
		t.Errorf("error: expected  %v got %v", true, z)
	}

}

func BenchmarkVerifyMerkleTree(b *testing.B) {
	// run verification the merkle tree bench
	for n := 0; n < 10000000; n++ {
		_, err := VerifyTree(vData.content, vData.leafHash, vData.intermedHash, vData.nodeHash, vData.expectedHash)
		if err != nil {
			b.Error("error: unexpected error:  ", err)
		}
	}
}

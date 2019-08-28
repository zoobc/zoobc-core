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

var expectedHash = []byte{76, 190, 133, 199, 105, 131, 156, 92, 19, 148, 223, 173, 129,
	47, 62, 141, 22, 180, 249, 64, 105, 246, 110, 120, 57, 206, 233, 181, 91, 177, 251, 158}

type fData struct {
	content      string
	leafHash     []byte
	intermedHash []byte
	nodeHash     []byte
	expectedHash []byte
}

var vData = &fData{
	content: "3e55352e649cb04ce701b6046d27163cf849a9fe73396ff072036ad7d0186ee3",
	leafHash: []byte{203, 33, 98, 252, 48, 232, 182, 33, 10, 220, 153, 63, 148, 150,
		155, 61, 48, 229, 124, 232, 43, 174, 194, 62, 101, 215, 48, 235, 202, 195, 222, 57},
	intermedHash: []byte{177, 247, 10, 191, 42, 44, 93, 250, 236, 151, 196, 144, 127, 107,
		206, 140, 106, 22, 87, 207, 96, 0, 45, 202, 74, 40, 55, 40, 221, 151, 113, 202},
	nodeHash: []byte{96, 160, 130, 169, 133, 176, 42, 86, 179, 124, 138, 146, 82, 150,
		166, 151, 201, 187, 88, 235, 131, 135, 59, 180, 28, 159, 119, 164, 93, 45, 181, 179},
	expectedHash: []byte{76, 190, 133, 199, 105, 131, 156, 92, 19, 148, 223, 173, 129,
		47, 62, 141, 22, 180, 249, 64, 105, 246, 110, 120, 57, 206, 233, 181, 91, 177, 251, 158},
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

package util

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestAllMerkle(t *testing.T) {
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
	result := merkle.GenerateMerkleRoot(hashes)
	nH := merkle.GetNecessaryHashes(bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}), 0)
	verRes := merkle.VerifyLeaf(bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8}), result, nH)
	if !verRes {
		t.Errorf("nh: %v\nresult: %v\nverres: %v", nH, base64.StdEncoding.EncodeToString(result.Bytes()), verRes)
	}
}

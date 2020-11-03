package model

import (
	"encoding/json"

	"github.com/zoobc/lib/address"
)

// NodeKey
type NodeKeyFromFile struct {
	ID        int64
	PublicKey string
	Seed      string
}

// MarshalJSON for NullString
func (m NodeKey) MarshalJSON() ([]byte, error) {
	publicKeyStr, err := address.EncodeZbcID(PrefixConstant_ZNK.String(), m.PublicKey)
	if err != nil {
		return nil, err
	}
	return json.Marshal(NodeKeyFromFile{
		ID:        m.ID,
		PublicKey: publicKeyStr,
		Seed:      m.Seed,
	})
}

// UnmarshalJSON for NullString
func (m *NodeKey) UnmarshalJSON(b []byte) error {
	var (
		result = &NodeKeyFromFile{}
		pubKey = make([]byte, 32)
	)
	err := json.Unmarshal(b, &result)
	if err != nil {
		return err
	}
	err = address.DecodeZbcID(result.PublicKey, pubKey)
	if err != nil {
		return err
	}
	*m = NodeKey{
		ID:        result.ID,
		PublicKey: pubKey,
		Seed:      result.Seed,
	}
	return nil
}

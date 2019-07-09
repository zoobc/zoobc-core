package p2p

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

// PeerService represent peer service
type PeerService struct {
	Peer      *model.Peer
	ChainType contract.ChainType
}

// NewPeerService to get instance of singleton peer service
func NewPeerService(chaintypeNumber int32) *PeerService {
	NewChainType := chaintype.GetChainType(chaintypeNumber)
	return &PeerService{
		ChainType: NewChainType,
	}
}

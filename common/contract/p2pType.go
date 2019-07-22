package contract

import "github.com/zoobc/zoobc-core/common/model"

// P2PType is interface for P2p instance
type P2PType interface {
	InitService(myAddress string, port uint32, wellknownPeers []string) (P2PType, error)
	StartP2P(chaintype ChainType)
	GetHostInstance() *model.Host
}

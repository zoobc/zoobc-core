package contract

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
)

// P2PType is interface for P2p instance
type P2PType interface {
	InitService(myAddress string, port uint32, wellknownPeers []string, obsr *observer.Observer) (P2PType, error)
	StartP2P()
	GetHostInstance() *model.Host
	SendBlockListener() observer.Listener
	SendTransactionListener() observer.Listener
}

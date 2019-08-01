package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/observer"
)

// InitP2P to initialize p2p strategy will used
// TODO: Add Switcer Interface
func InitP2P(myAddress string, port uint32, wellknownPeers []string, p2pType contract.P2PType, obsr *observer.Observer) contract.P2PType {
	p2pService, err := p2pType.InitService(myAddress, port, wellknownPeers, obsr)
	if err != nil {
		log.Fatalf("Faild to initialize P2P service\nError : %v\n", err)
	}
	return p2pService
}

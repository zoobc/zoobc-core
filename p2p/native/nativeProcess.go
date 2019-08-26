package native

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/native/service"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// startServer to run p2p service as server
func startServer(
	blockServices map[int32]coreService.BlockServiceInterface,
	obsr *observer.Observer,
	nodeSecretPhrase string) {
	port := hostServiceInstance.Host.GetInfo().GetPort()
	listener := nativeUtil.ServerListener(int(port))
	go func() {
		_ = service.NewServerService(blockServices, obsr, nodeSecretPhrase).StartListening(listener)
	}()
}

// resolvePeersThread to periodically try get response from peers in UnresolvedPeer list
func resolvePeersThread() {
	go hostServiceInstance.ResolvePeers()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hostServiceInstance.ResolvePeers()
			go hostServiceInstance.UpdateResolvedPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// getMorePeersThread to periodically request more peers from another node in Peers list
func getMorePeersThread() {
	go hostServiceInstance.GetMorePeersHandler()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hostServiceInstance.GetMorePeersHandler()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// updateBlacklistedStatus to periodically check blacklisting time of black listed peer,
// every 60sec if there are blacklisted peers to unblacklist
func updateBlacklistedStatus() {
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint64(time.Now().Unix())
				for _, p := range hostServiceInstance.Host.GetBlacklistedPeers() {
					if p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+constant.BlacklistingPeriod <= curTime {
						hostServiceInstance.Host.KnownPeers[nativeUtil.GetFullAddressPeer(p)] = hostServiceInstance.PeerUnblacklist(p)
					}
				}
				break
			case <-sigs:
				ticker.Stop()
				return
			}
		}
	}()
}

/* 	========================================
 *	Handler to send data
 *	========================================
 */

// sendBlock send block to the list peer
func sendBlock(block *model.Block) {
	peers := hostServiceInstance.GetResolvedPeers()
	for _, peer := range peers {
		go func() {
			receipt, err := service.NewPeerServiceClient().SendBlock(peer, block)
			if err != nil {
				log.Warnf("sendBlockHandler Error accord %v\n", err)
			}
			fmt.Printf("receipt %v\n", receipt)
		}()
	}
}

// sendTransaction send transaction to the list peer
func sendTransactionBytes(transactionBytes []byte) {
	peers := hostServiceInstance.GetResolvedPeers()
	for _, peer := range peers {
		go func() {
			_, err := service.NewPeerServiceClient().SendTransaction(peer, transactionBytes)
			if err != nil {
				log.Warnf("sendTransactionBytesHandler Error accord %v\n", err)
			}
		}()
	}
}

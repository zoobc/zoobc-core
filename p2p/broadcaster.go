package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/p2p/rpcClient"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

type (
	// p2p broadcaster
	BroadcasterInterface interface {
		SendTransactionBytes(transactionBytes []byte, peers map[string]*model.Peer)
		SendBlock(block *model.Block, peers map[string]*model.Peer)
		SendMyPeers(srcPeer *model.Node, destPeer *model.Peer, peers []*model.Node)
	}

	Broadcaster struct {
		PeerServiceClient rpcClient.PeerServiceClientInterface
		QueryExecutor     query.ExecutorInterface
		ReceiptQuery      query.ReceiptQueryInterface
	}
)

func NewBroadcaster(
	peerServiceClient rpcClient.PeerServiceClientInterface,
	queryExecutor query.ExecutorInterface,
	receiptQuery query.ReceiptQueryInterface,
) *Broadcaster {
	return &Broadcaster{
		PeerServiceClient: peerServiceClient,
		QueryExecutor:     queryExecutor,
		ReceiptQuery:      receiptQuery,
	}
}

// sendBlock send block to the list peer provided
func (bc *Broadcaster) SendBlock(block *model.Block, peers map[string]*model.Peer) {
	for _, peer := range peers {
		go func() {
			receipt, err := bc.PeerServiceClient.SendBlock(peer, block)
			if err != nil {
				log.Warnf("sendBlockHandler Error accord %v\n", err)
				return
			}
			insertReceiptQ, insertReceiptArg := bc.ReceiptQuery.InsertReceipt(receipt)
			res, err := bc.QueryExecutor.ExecuteStatement(insertReceiptQ, insertReceiptArg...)
			if err != nil {
				log.Warnf("fail to save receipt")
			}
			log.Infof("receipt saved - result: %v\n", res)
		}()
	}
}

// sendTransaction send transaction to the list peer provided
func (bc *Broadcaster) SendTransactionBytes(transactionBytes []byte, peers map[string]*model.Peer) {
	for _, peer := range peers {
		go func() {
			_, err := bc.PeerServiceClient.SendTransaction(peer, transactionBytes)
			if err != nil {
				log.Warnf("sendTransactionBytesHandler Error accord %v\n", err)
			}
		}()
	}
}

// SendMyPeers sends resolved peers of a host including the address of the host itself
func (bc *Broadcaster) SendMyPeers(srcPeer *model.Node, destPeer *model.Peer, peers []*model.Node) {
	var myPeersInfo []*model.Node
	myPeersInfo = append(myPeersInfo, srcPeer)
	for _, peer := range peers {
		myPeersInfo = append(myPeersInfo, peer)
	}

	_, err := bc.PeerServiceClient.SendPeers(destPeer, myPeersInfo)
	if err != nil {
		log.Printf("failed to send the host peers to %s: %v", p2pUtil.GetFullAddressPeer(destPeer), err)
	}
}

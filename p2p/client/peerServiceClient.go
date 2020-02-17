package client

import (
	"context"
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/service"
	coreService "github.com/zoobc/zoobc-core/core/service"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type (
	// PeerServiceClientInterface acts as interface for PeerServiceClient
	PeerServiceClientInterface interface {
		GetPeerInfo(destPeer *model.Peer) (*model.Node, error)
		GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error)
		SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error)
		SendBlock(
			destPeer *model.Peer,
			block *model.Block,
			chainType chaintype.ChainType,
		) error
		SendTransaction(
			destPeer *model.Peer,
			transactionBytes []byte,
			chainType chaintype.ChainType,
		) error
		SendBlockTransactions(
			destPeer *model.Peer,
			transactionsBytes [][]byte,
			chainType chaintype.ChainType,
		) error
		RequestBlockTransactions(
			destPeer *model.Peer,
			transactonIDs []int64,
			chainType chaintype.ChainType,
			blockID int64,
		) error
		GetCumulativeDifficulty(*model.Peer, chaintype.ChainType) (*model.GetCumulativeDifficultyResponse, error)
		GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, lastBlockID,
			astMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error)
		GetNextBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, blockID int64, limit uint32) (*model.BlockIdsResponse, error)
		GetNextBlocks(destPeer *model.Peer, chaintype chaintype.ChainType, blockIds []int64, blockID int64) (*model.BlocksData, error)
		// connection managements
		DeleteConnection(destPeer *model.Peer) error
		GetConnection(destPeer *model.Peer) (*grpc.ClientConn, error)
	}
	// PeerServiceClient represent peer service
	PeerServiceClient struct {
		Dialer              Dialer
		Logger              *log.Logger
		QueryExecutor       query.ExecutorInterface
		NodeReceiptQuery    query.NodeReceiptQueryInterface
		BatchReceiptQuery   query.BatchReceiptQueryInterface
		MerkleTreeQuery     query.MerkleTreeQueryInterface
		ReceiptService      coreService.ReceiptServiceInterface
		NodePublicKey       []byte
		Host                *model.Host
		PeerConnections     map[string]*grpc.ClientConn
		PeerConnectionsLock sync.RWMutex
	}
	// Dialer represent peer service
	Dialer func(destinationPeer *model.Peer) (*grpc.ClientConn, error)
)

// NewPeerServiceClient to get instance of singleton peer service, this should only be instantiated from main.go
func NewPeerServiceClient(
	queryExecutor query.ExecutorInterface,
	nodeReceiptQuery query.NodeReceiptQueryInterface,
	nodePublicKey []byte,
	batchReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	receiptService coreService.ReceiptServiceInterface,
	host *model.Host,
	logger *log.Logger,
) PeerServiceClientInterface {
	// set to current struct log
	return &PeerServiceClient{
		Dialer: func(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
			conn, err := grpc.Dial(
				p2pUtil.GetFullAddressPeer(destinationPeer),
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(interceptor.NewClientInterceptor(
					logger,
					map[codes.Code]string{
						codes.Unavailable:     "indicates the destination service is currently unavailable",
						codes.InvalidArgument: "indicates the argument request is invalid",
						codes.Unauthenticated: "indicates the request is unauthenticated",
					},
				)),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		QueryExecutor:     queryExecutor,
		NodeReceiptQuery:  nodeReceiptQuery,
		BatchReceiptQuery: batchReceiptQuery,
		MerkleTreeQuery:   merkleTreeQuery,
		ReceiptService:    receiptService,
		NodePublicKey:     nodePublicKey,
		Logger:            logger,
		Host:              host,
		PeerConnections:   make(map[string]*grpc.ClientConn),
	}
}

// saveNewConnection cache the connection to peer to keep an open connection, this avoid the overhead of open/close
// connection on every request
func (psc *PeerServiceClient) saveNewConnection(destPeer *model.Peer) (*grpc.ClientConn, error) {
	psc.PeerConnectionsLock.Lock()
	defer psc.PeerConnectionsLock.Unlock()
	connection, err := psc.Dialer(destPeer)
	if err != nil {
		return nil, err
	}
	// psc.PeerConnections[p2pUtil.GetFullAddressPeer(destPeer)] = connection
	return connection, nil
}

// DeleteConnection delete the cached connection in psc.PeerConnections
func (psc *PeerServiceClient) DeleteConnection(destPeer *model.Peer) error {
	psc.PeerConnectionsLock.Lock()
	defer psc.PeerConnectionsLock.Unlock()
	connection := psc.PeerConnections[p2pUtil.GetFullAddressPeer(destPeer)]
	if connection == nil {
		return nil
	}
	err := connection.Close()
	if err != nil {
		return err
	}
	delete(psc.PeerConnections, p2pUtil.GetFullAddressPeer(destPeer))
	return nil
}

func (psc *PeerServiceClient) GetConnection(destPeer *model.Peer) (*grpc.ClientConn, error) {
	var (
		exist *grpc.ClientConn
		err   error
	)
	// psc.PeerConnectionsLock.RLock()
	// exist = psc.PeerConnections[p2pUtil.GetFullAddressPeer(destPeer)]
	// psc.PeerConnectionsLock.RUnlock()
	// if exist == nil {
	// 	exist, err = psc.saveNewConnection(destPeer)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	exist, err = psc.saveNewConnection(destPeer)
	if err != nil {
		return nil, err
	}
	// add a copy to avoid pointer delete
	return exist, nil
}

// setDefaultMetadata use to set default metadata.
// It will use in validation request
func (psc *PeerServiceClient) setDefaultMetadata() map[string]string {
	return map[string]string{p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(psc.Host.GetInfo())}
}

// getDefaultContext use to get default context with deadline & default metadata
func (psc *PeerServiceClient) getDefaultContext(requestTimeOut time.Duration) (context.Context, context.CancelFunc) {
	if requestTimeOut == 0 {
		requestTimeOut = math.MaxInt64
	}
	var (
		header                      = metadata.New(psc.setDefaultMetadata())
		clientDeadline              = time.Now().Add(requestTimeOut)
		ctxWithDeadline, cancelFunc = context.WithDeadline(context.Background(), clientDeadline)
	)
	return metadata.NewOutgoingContext(ctxWithDeadline, header), cancelFunc
}

// GetPeerInfo to get Peer info
func (psc *PeerServiceClient) GetPeerInfo(destPeer *model.Peer) (*model.Node, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoClient)

	// add a copy to avoid pointer delete
	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(10 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(
		ctx,
		&model.GetPeerInfoRequest{
			Version: "v1,.0.1",
		},
	)
	if err != nil {
		return nil, err
	}
	return res, err
}

// GetMorePeers to collect more peers available
func (psc *PeerServiceClient) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetMorePeersClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetMorePeersClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(10 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(ctx, &model.Empty{})
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendPeers sends set of peers to other node (to populate the network)
func (psc *PeerServiceClient) SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendPeersClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendPeersClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(10 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	res, err := p2pClient.SendPeers(ctx, &model.SendPeersRequest{
		Peers: peersInfo,
	})
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendBlock send block to selected peer, got Receipt
func (psc *PeerServiceClient) SendBlock(
	destPeer *model.Peer,
	block *model.Block,
	chainType chaintype.ChainType,
) error {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendBlockClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendBlockClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
	defer connection.Close()
	var (
		response       *model.SendBlockResponse
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(25 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	response, err = p2pClient.SendBlock(ctx, &model.SendBlockRequest{
		SenderPublicKey: psc.NodePublicKey,
		Block:           block,
		ChainType:       chainType.GetTypeInt(),
	})
	if err != nil {
		return err
	}
	if response == nil || response.BatchReceipt == nil {
		return err
	}
	// validate receipt before storing
	err = psc.ReceiptService.ValidateReceipt(response.BatchReceipt)
	if err != nil {
		return err
	}
	err = psc.storeReceipt(response.BatchReceipt)
	return err
}

// SendTransaction send transaction to selected peer
func (psc *PeerServiceClient) SendTransaction(
	destPeer *model.Peer,
	transactionBytes []byte,
	chainType chaintype.ChainType,
) error {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendTransactionClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendTransactionClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
	defer connection.Close()
	var (
		response       *model.SendTransactionResponse
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(20 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	response, err = p2pClient.SendTransaction(ctx, &model.SendTransactionRequest{
		SenderPublicKey:  psc.NodePublicKey,
		TransactionBytes: transactionBytes,
		ChainType:        chainType.GetTypeInt(),
	})
	if err != nil {
		return err
	}
	if response == nil || response.BatchReceipt == nil {
		return nil
	}
	err = psc.ReceiptService.ValidateReceipt(response.BatchReceipt)
	if err != nil {
		return err
	}
	err = psc.storeReceipt(response.BatchReceipt)
	return err
}

// SendBlockTransactions sends transactions required by a block requested by the peer
func (psc *PeerServiceClient) SendBlockTransactions(
	destPeer *model.Peer,
	transactionsBytes [][]byte,
	chainType chaintype.ChainType,
) error {
	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
	defer connection.Close()
	var (
		response       *model.SendBlockTransactionsResponse
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(20 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	response, err = p2pClient.SendBlockTransactions(ctx, &model.SendBlockTransactionsRequest{
		SenderPublicKey:   psc.NodePublicKey,
		TransactionsBytes: transactionsBytes,
		ChainType:         chainType.GetTypeInt(),
	})
	if err != nil {
		return err
	}
	if response == nil || response.BatchReceipts == nil || len(response.BatchReceipts) == 0 {
		return nil
	}

	for _, batchReceipt := range response.BatchReceipts {
		// continue even though some receipts are failing
		_ = psc.ReceiptService.ValidateReceipt(batchReceipt)
		_ = psc.storeReceipt(batchReceipt)
	}
	return err
}

func (psc *PeerServiceClient) RequestBlockTransactions(
	destPeer *model.Peer,
	transactonIDs []int64,
	chainType chaintype.ChainType,
	blockID int64,
) error {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pRequestBlockTransactionsClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pRequestBlockTransactionsClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(20 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	_, err = p2pClient.RequestBlockTransactions(ctx, &model.RequestBlockTransactionsRequest{
		TransactionIDs: transactonIDs,
		ChainType:      chainType.GetTypeInt(),
		BlockID:        blockID,
	})
	if err != nil {
		return err
	}
	return nil
}

// GetCumulativeDifficulty request the cumulative difficulty status of a node
func (psc *PeerServiceClient) GetCumulativeDifficulty(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetCumulativeDifficultyClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetCumulativeDifficultyClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(15 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	res, err := p2pClient.GetCumulativeDifficulty(ctx, &model.GetCumulativeDifficultyRequest{
		ChainType: chaintype.GetTypeInt(),
	})
	if err != nil {
		psc.Logger.Infof("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetCommonMilestoneBlockIDs request the blockIds that may act as milestone block
func (psc *PeerServiceClient) GetCommonMilestoneBlockIDs(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
	lastBlockID, lastMilestoneBlockID int64,
) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetCommonMilestoneBlockIDsClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetCommonMilestoneBlockIDsClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(15 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	res, err := p2pClient.GetCommonMilestoneBlockIDs(ctx, &model.GetCommonMilestoneBlockIdsRequest{
		ChainType:            chaintype.GetTypeInt(),
		LastBlockID:          lastBlockID,
		LastMilestoneBlockID: lastMilestoneBlockID,
	})
	if err != nil {
		psc.Logger.Infof("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlockIDs request the blockIds of the next blocks requested
func (psc *PeerServiceClient) GetNextBlockIDs(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
	blockID int64,
	limit uint32,
) (*model.BlockIdsResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetNextBlockIDsClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetNextBlockIDsClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(15 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	res, err := p2pClient.GetNextBlockIDs(ctx, &model.GetNextBlockIdsRequest{
		ChainType: chaintype.GetTypeInt(),
		BlockId:   blockID,
		Limit:     limit,
	})
	if err != nil {
		psc.Logger.Warnf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlocks request the next blocks matching the array of blockIds
func (psc *PeerServiceClient) GetNextBlocks(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
	blockIds []int64,
	blockID int64,
) (*model.BlocksData, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetNextBlocksClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetNextBlocksClient)

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(15 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	res, err := p2pClient.GetNextBlocks(
		ctx,
		&model.GetNextBlocksRequest{
			ChainType: chaintype.GetTypeInt(),
			BlockId:   blockID,
			BlockIds:  blockIds,
		},
	)
	if err != nil {
		return nil, err
	}
	return res, err
}

// storeReceipt function will decide to storing receipt into node_receipt or batch_receipt
// and will generate _merkle_root_
func (psc *PeerServiceClient) storeReceipt(batchReceipt *model.BatchReceipt) error {
	var (
		err error
	)

	psc.Logger.Info("Insert Batch Receipt")
	insertBatchReceiptQ, argsInsertBatchReceiptQ := psc.BatchReceiptQuery.InsertBatchReceipt(batchReceipt)
	_, err = psc.QueryExecutor.ExecuteStatement(insertBatchReceiptQ, argsInsertBatchReceiptQ...)
	if err != nil {
		return err
	}

	monitoring.IncrementReceiptCounter()
	return nil
}

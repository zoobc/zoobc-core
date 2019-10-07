package client

import (
	"bytes"
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	//  PeerServiceClientInterface acts as interface for PeerServiceClient
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
		GetCumulativeDifficulty(*model.Peer, chaintype.ChainType) (*model.GetCumulativeDifficultyResponse, error)
		GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, lastBlockID,
			astMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error)
		GetNextBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, blockID int64, limit uint32) (*model.BlockIdsResponse, error)
		GetNextBlocks(destPeer *model.Peer, chaintype chaintype.ChainType, blockIds []int64, blockID int64) (*model.BlocksData, error)
	}
	// PeerService represent peer service
	PeerServiceClient struct {
		Dialer            Dialer
		Logger            *log.Logger
		QueryExecutor     query.ExecutorInterface
		ReceiptQuery      query.ReceiptQueryInterface
		BatchReceiptQuery query.BatchReceiptQueryInterface
		MerkleTreeQuery   query.MerkleTreeQueryInterface
		NodePublicKey     []byte
		Host              *model.Host
	}
)

// PeerService represent peer service
type Dialer func(destinationPeer *model.Peer) (*grpc.ClientConn, error)

// ClientPeerService to get instance of singleton peer service, this should only be instantiated from main.go
func NewPeerServiceClient(
	queryExecutor query.ExecutorInterface,
	receiptQuery query.ReceiptQueryInterface,
	nodePublicKey []byte,
	batchReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	host *model.Host,
) PeerServiceClientInterface {
	logLevels := viper.GetStringSlice("logLevels")
	apiLogger, _ := util.InitLogger(".log/", "debugP2PClient.log", logLevels)
	// set to current struct log
	return &PeerServiceClient{
		Dialer: func(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
			conn, err := grpc.Dial(
				p2pUtil.GetFullAddressPeer(destinationPeer),
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(interceptor.NewClientInterceptor(apiLogger)),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		QueryExecutor:     queryExecutor,
		ReceiptQuery:      receiptQuery,
		BatchReceiptQuery: batchReceiptQuery,
		MerkleTreeQuery:   merkleTreeQuery,
		NodePublicKey:     nodePublicKey,
		Logger:            apiLogger,
		Host:              host,
	}
}

func (psc *PeerServiceClient) setDefaultMetadata() map[string]string {
	return map[string]string{p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(psc.Host.GetInfo())}
}

// GetPeerInfo to get Peer info
func (psc *PeerServiceClient) GetPeerInfo(destPeer *model.Peer) (*model.Node, error) {
	var (
		connection, _ = psc.Dialer(destPeer)
		p2pClient     = service.NewP2PCommunicationClient(connection)
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
	)
	defer connection.Close()

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(
		ctx,
		&model.GetPeerInfoRequest{
			Version: "v1,.0.1",
		})
	if err != nil {
		return nil, err
	}
	return res, err
}

// GetMorePeers to collect more peers available
func (psc *PeerServiceClient) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	var (
		connection, _ = psc.Dialer(destPeer)
		p2pClient     = service.NewP2PCommunicationClient(connection)
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
	)
	defer connection.Close()

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(ctx, &model.Empty{})
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendPeers sends set of peers to other node (to populate the network)
func (psc *PeerServiceClient) SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	var (
		connection, _ = psc.Dialer(destPeer)
		p2pClient     = service.NewP2PCommunicationClient(connection)
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
	)
	defer connection.Close()
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
	var (
		err           error
		response      *model.SendBlockResponse
		connection, _ = psc.Dialer(destPeer)
		p2pClient     = service.NewP2PCommunicationClient(connection)
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
	)
	defer connection.Close()
	response, err = p2pClient.SendBlock(ctx, &model.SendBlockRequest{
		SenderPublicKey: psc.NodePublicKey,
		Block:           block,
		ChainType:       chainType.GetTypeInt(),
	})
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
	var (
		err           error
		response      *model.SendTransactionResponse
		connection, _ = psc.Dialer(destPeer)
		p2pClient     = service.NewP2PCommunicationClient(connection)
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
	)
	defer connection.Close()

	response, err = p2pClient.SendTransaction(ctx, &model.SendTransactionRequest{
		SenderPublicKey:  psc.NodePublicKey,
		TransactionBytes: transactionBytes,
		ChainType:        chainType.GetTypeInt(),
	})
	if err != nil {
		return err
	}
	err = psc.storeReceipt(response.BatchReceipt)
	return err
}

// GetCumulativeDifficulty request the cumulative difficulty status of a node
func (psc PeerServiceClient) GetCumulativeDifficulty(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	var (
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
		connection, _ = grpc.Dial(
			p2pUtil.GetFullAddressPeer(destPeer),
			grpc.WithInsecure(),
			// grpc.WithUnaryInterceptor(),
		)
		p2pClient = service.NewP2PCommunicationClient(connection)
	)
	defer connection.Close()

	res, err := p2pClient.GetCumulativeDifficulty(ctx, &model.GetCumulativeDifficultyRequest{
		ChainType: chaintype.GetTypeInt(),
	})
	if err != nil {
		psc.Logger.Warnf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetCommonMilestoneBlockIDs request the blockIds that may act as milestone block
func (psc PeerServiceClient) GetCommonMilestoneBlockIDs(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
	lastBlockID, lastMilestoneBlockID int64,
) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	var (
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
		connection, _ = grpc.Dial(
			p2pUtil.GetFullAddressPeer(destPeer),
			grpc.WithInsecure(),
			// grpc.WithUnaryInterceptor(),
		)
		p2pClient = service.NewP2PCommunicationClient(connection)
	)
	defer connection.Close()

	res, err := p2pClient.GetCommonMilestoneBlockIDs(ctx, &model.GetCommonMilestoneBlockIdsRequest{
		ChainType:            chaintype.GetTypeInt(),
		LastBlockID:          lastBlockID,
		LastMilestoneBlockID: lastMilestoneBlockID,
	})
	if err != nil {
		psc.Logger.Warnf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlockIDs request the blockIds of the next blocks requested
func (psc PeerServiceClient) GetNextBlockIDs(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
	blockID int64,
	limit uint32,
) (*model.BlockIdsResponse, error) {
	var (
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
		connection, _ = grpc.Dial(
			p2pUtil.GetFullAddressPeer(destPeer),
			grpc.WithInsecure(),
			// grpc.WithUnaryInterceptor(),
		)
		p2pClient = service.NewP2PCommunicationClient(connection)
	)
	defer connection.Close()

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
func (psc PeerServiceClient) GetNextBlocks(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
	blockIds []int64,
	blockID int64,
) (*model.BlocksData, error) {
	var (
		header        = metadata.New(psc.setDefaultMetadata())
		ctx           = metadata.NewOutgoingContext(context.Background(), header)
		connection, _ = grpc.Dial(
			p2pUtil.GetFullAddressPeer(destPeer),
			grpc.WithInsecure(),
			// grpc.WithUnaryInterceptor(),
		)
		p2pClient = service.NewP2PCommunicationClient(connection)
	)
	defer connection.Close()

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
		err            error
		count          uint32
		queries        [][]interface{}
		batchReceipts  []*model.BatchReceipt
		receipt        *model.Receipt
		hashedReceipts []*bytes.Buffer
		merkleRoot     util.MerkleRoot
	)

	psc.Logger.Info("Insert Batch Receipt")
	insertBatchReceiptQ, argsInsertBatchReceiptQ := psc.BatchReceiptQuery.InsertBatchReceipt(batchReceipt)
	_, err = psc.QueryExecutor.ExecuteStatement(insertBatchReceiptQ, argsInsertBatchReceiptQ...)
	if err != nil {
		return err
	}

	countBatchReceiptQ := query.GetTotalRecordOfSelect(
		psc.BatchReceiptQuery.GetBatchReceipts(constant.ReceiptBatchMaximum, 0),
	)
	err = psc.QueryExecutor.ExecuteSelectRow(countBatchReceiptQ).Scan(&count)
	if err != nil {
		return err
	}
	psc.Logger.Info("Count Batch Receipts: ", count)

	if count >= constant.ReceiptBatchMaximum {
		psc.Logger.Info("Start Store Batch To Receipt: ", count)
		getBatchReceiptsQ := psc.BatchReceiptQuery.GetBatchReceipts(constant.ReceiptBatchMaximum, 0)
		rows, err := psc.QueryExecutor.ExecuteSelect(getBatchReceiptsQ, false)
		if err != nil {
			return err
		}
		defer rows.Close()

		queries = make([][]interface{}, (constant.ReceiptBatchMaximum*2)+1)
		batchReceipts = psc.BatchReceiptQuery.BuildModel(batchReceipts, rows)

		for _, b := range batchReceipts {
			hashedReceipts = append(
				hashedReceipts,
				bytes.NewBuffer(util.GetSignedBatchReceiptBytes(b)),
			)
		}
		_, err = merkleRoot.GenerateMerkleRoot(hashedReceipts)
		if err != nil {
			return err
		}
		rootMerkle, treeMerkle := merkleRoot.ToBytes()

		for k, r := range batchReceipts {

			var (
				br       = r
				rmrIndex = uint32(k)
			)

			receipt = &model.Receipt{
				BatchReceipt: br,
				RMR:          rootMerkle,
				RMRIndex:     rmrIndex,
			}
			insertReceiptQ, insertReceiptArgs := psc.ReceiptQuery.InsertReceipt(receipt)
			queries[k] = append([]interface{}{insertReceiptQ}, insertReceiptArgs...)
			removeBatchReceiptQ, removeBatchReceiptArgs := psc.BatchReceiptQuery.RemoveBatchReceipt(br.DatumType, br.DatumHash)
			queries[(constant.ReceiptBatchMaximum)+uint32(k)] = append([]interface{}{removeBatchReceiptQ}, removeBatchReceiptArgs...)
		}

		insertMerkleTreeQ, insertMerkleTreeArgs := psc.MerkleTreeQuery.InsertMerkleTree(rootMerkle, treeMerkle)
		queries[len(queries)-1] = append([]interface{}{insertMerkleTreeQ}, insertMerkleTreeArgs...)

		err = psc.QueryExecutor.BeginTx()
		if err != nil {
			return err
		}
		err = psc.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			return err
		}
		err = psc.QueryExecutor.CommitTx()
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

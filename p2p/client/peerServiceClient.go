// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package client

import (
	"context"
	"fmt"
	"github.com/zoobc/zoobc-core/common/feedbacksystem"
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type (
	// PeerServiceClientInterface acts as interface for PeerServiceClient
	PeerServiceClientInterface interface {
		GetNodeAddressesInfo(destPeer *model.Peer, nodeRegistrations []*model.NodeRegistration) (*model.GetNodeAddressesInfoResponse, error)
		SendNodeAddressInfo(destPeer *model.Peer, nodeAddressInfos []*model.NodeAddressInfo) (*model.Empty, error)
		GetNodeProofOfOrigin(destPeer *model.Peer) (*model.ProofOfOrigin, error)
		GetPeerInfo(destPeer *model.Peer) (*model.GetPeerInfoResponse, error)
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
			transactionIDs []int64,
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
		RequestDownloadFile(destPeer *model.Peer, snapshotHash []byte, fileChunkNames []string) (*model.FileDownloadResponse, error)
	}
	// PeerServiceClient represent peer service
	PeerServiceClient struct {
		Dialer                   Dialer
		Logger                   *log.Logger
		QueryExecutor            query.ExecutorInterface
		NodeReceiptQuery         query.BatchReceiptQueryInterface
		MerkleTreeQuery          query.MerkleTreeQueryInterface
		ReceiptService           coreService.ReceiptServiceInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodePublicKey            []byte
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerConnections          map[string]*grpc.ClientConn
		PeerConnectionsLock      sync.RWMutex
		NodeAuthValidation       auth.NodeAuthValidationInterface
		FeedbackStrategy         feedbacksystem.FeedbackStrategyInterface
	}
	// Dialer represent peer service
	Dialer func(destinationPeer *model.Peer) (*grpc.ClientConn, error)
)

// NewPeerServiceClient to get instance of singleton peer service, this should only be instantiated from main.go
func NewPeerServiceClient(
	queryExecutor query.ExecutorInterface,
	nodeReceiptQuery query.BatchReceiptQueryInterface,
	nodePublicKey []byte,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	receiptService coreService.ReceiptServiceInterface,
	nodeConfigurationService coreService.NodeConfigurationServiceInterface,
	nodeAuthValidation auth.NodeAuthValidationInterface,
	feedbackStrategy feedbacksystem.FeedbackStrategyInterface,
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
		QueryExecutor:            queryExecutor,
		NodeReceiptQuery:         nodeReceiptQuery,
		MerkleTreeQuery:          merkleTreeQuery,
		ReceiptService:           receiptService,
		NodeRegistrationService:  nodeRegistrationService,
		NodePublicKey:            nodePublicKey,
		Logger:                   logger,
		NodeConfigurationService: nodeConfigurationService,
		PeerConnections:          make(map[string]*grpc.ClientConn),
		NodeAuthValidation:       nodeAuthValidation,
		FeedbackStrategy:         feedbackStrategy,
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
	psc.PeerConnections[p2pUtil.GetFullAddressPeer(destPeer)] = connection
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
	psc.PeerConnectionsLock.RLock()
	exist = psc.PeerConnections[p2pUtil.GetFullAddressPeer(destPeer)]
	psc.PeerConnectionsLock.RUnlock()
	if exist == nil {
		exist, err = psc.saveNewConnection(destPeer)
		if err != nil {
			return nil, err
		}
	}
	// add a copy to avoid pointer delete
	return exist, nil
}

// setDefaultMetadata use to set default metadata.
// It will use in validation request
func (psc *PeerServiceClient) setDefaultMetadata() map[string]string {
	return map[string]string{
		p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(psc.NodeConfigurationService.GetHost().GetInfo()),
		"version":                         psc.NodeConfigurationService.GetHost().GetInfo().GetVersion(),
		"codename":                        psc.NodeConfigurationService.GetHost().GetInfo().GetCodeName(),
	}
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

// GetNodeAddressesInfo to get a list of node addresses from a peer
func (psc *PeerServiceClient) GetNodeAddressesInfo(
	destPeer *model.Peer,
	nodeRegistrations []*model.NodeRegistration,
) (*model.GetNodeAddressesInfoResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	// add a copy to avoid pointer delete
	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(2 * time.Second)
		nodeIDs        = make([]int64, len(nodeRegistrations))
	)
	defer func() {
		cancelReq()
	}()

	for i, nr := range nodeRegistrations {
		nodeIDs[i] = nr.NodeID
	}

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetNodeAddressesInfo(
		ctx,
		&model.GetNodeAddressesInfoRequest{
			NodeIDs: nodeIDs,
		},
	)
	if err != nil {
		return nil, err
	}
	if res.NodeAddressesInfo == nil {
		return nil, blocker.NewBlocker(blocker.P2PPeerError, fmt.Sprintf(
			"GetNodeAddressesInfo client: peer %s:%d returned an empty node address list",
			destPeer.GetInfo().Address, destPeer.GetInfo().Port))
	}
	monitoring.IncrementGetAddressInfoTableFromPeer()

	return res, err
}

// GetPeerInfo to get Peer info
func (psc *PeerServiceClient) GetPeerInfo(destPeer *model.Peer) (*model.GetPeerInfoResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	// add a copy to avoid pointer delete
	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
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
		&model.GetPeerInfoRequest{},
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
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
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

// GetNodeProofOfOrigin get a cryptographic prove of a node authenticity and origin
func (psc *PeerServiceClient) GetNodeProofOfOrigin(
	destPeer *model.Peer,
) (*model.ProofOfOrigin, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetNodeProofOfOwnershipInfoClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetNodeProofOfOwnershipInfoClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	if destPeer.Info.GetID() == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf(
			"Cannot get proof of origin from an unregistered node: %s:%d",
			destPeer.GetInfo().Address, destPeer.GetInfo().Port))
	}

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(10 * time.Second)
	)
	defer func() {
		cancelReq()
	}()

	// generate the otp
	challenge, err := util.GenerateRandomBytes(64)
	if err != nil {
		return nil, err
	}

	// send the challenge
	res, err := p2pClient.GetNodeProofOfOrigin(ctx, &model.GetNodeProofOfOriginRequest{
		ChallengeMessage: challenge,
		Timestamp:        time.Now().Unix() + constant.ProofOfOriginExpirationOffset,
	})
	if err != nil {
		return nil, err
	}

	// validate response: message signature = challenge+timestamp
	nr, err := psc.NodeRegistrationService.GetNodeRegistrationByNodeID(destPeer.Info.GetID())
	if err != nil {
		return nil, err
	}
	err = psc.NodeAuthValidation.ValidateProofOfOrigin(res, nr.GetNodePublicKey(), challenge)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SendNodeAddressInfo sends a nodeAddressInfo to other node (to populate the network)
func (psc *PeerServiceClient) SendNodeAddressInfo(destPeer *model.Peer, nodeAddressInfos []*model.NodeAddressInfo) (*model.Empty, error) {

	if len(nodeAddressInfos) == 0 {
		return &model.Empty{}, nil
	}

	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendNodeAddressInfoClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendNodeAddressInfoClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(2 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	res, err := p2pClient.SendNodeAddressInfo(ctx, &model.SendNodeAddressInfoRequest{
		NodeAddressInfoMessage: nodeAddressInfos,
	})
	if err != nil {
		return nil, err
	}
	monitoring.IncrementSendAddressInfoToPeer()
	return res, err
}

// SendPeers sends set of peers to other node (to populate the network)
func (psc *PeerServiceClient) SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendPeersClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendPeersClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
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
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
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
	if response == nil || response.GetReceipt() == nil {
		if response.GetReceipt() == nil {
			psc.Logger.Infof("NO RECEIPT FROM %s:%d", destPeer.Info.Address, destPeer.Info.Port)
		}
		return err
	}

	// validate receipt before storing
	err = psc.ReceiptService.CheckDuplication(psc.NodePublicKey, response.GetReceipt().GetDatumHash())
	if err != nil {
		return err
	}
	err = psc.ReceiptService.ValidateReceipt(response.GetReceipt())
	if err != nil {
		return err
	}

	return psc.ReceiptService.StoreReceipt(
		response.GetReceipt(),
		response.GetReceipt().GetSenderPublicKey(),
		&chaintype.MainChain{},
	)
}

// SendTransaction send transaction to selected peer
func (psc *PeerServiceClient) SendTransaction(
	destPeer *model.Peer,
	transactionBytes []byte,
	chainType chaintype.ChainType,
) error {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendTransactionClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendTransactionClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
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
	if response == nil || response.GetReceipt() == nil {
		return nil
	}

	err = psc.ReceiptService.CheckDuplication(psc.NodePublicKey, response.GetReceipt().GetDatumHash())
	if err != nil {
		return err
	}
	err = psc.ReceiptService.ValidateReceipt(response.GetReceipt())
	if err != nil {
		return err
	}
	return psc.ReceiptService.StoreReceipt(
		response.GetReceipt(),
		response.GetReceipt().GetSenderPublicKey(),
		&chaintype.MainChain{},
	)
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
	var (
		response       *model.SendBlockTransactionsResponse
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(20 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendTransactionClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendTransactionClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	response, err = p2pClient.SendBlockTransactions(ctx, &model.SendBlockTransactionsRequest{
		SenderPublicKey:   psc.NodePublicKey,
		TransactionsBytes: transactionsBytes,
		ChainType:         chainType.GetTypeInt(),
	})
	if err != nil {
		return err
	}
	if response == nil || response.GetReceipts() == nil || len(response.GetReceipts()) == 0 {
		return nil
	}

	// continue even though some receipts are failing
	for _, receipt := range response.GetReceipts() {

		err = psc.ReceiptService.CheckDuplication(psc.NodePublicKey, receipt.GetDatumHash())
		if err != nil {
			psc.Logger.Warnf("[SendBlockTransactions:CheckDuplication] - %s", err.Error())
			continue
		}
		err = psc.ReceiptService.ValidateReceipt(receipt)
		if err != nil {
			psc.Logger.Warnf("[SendBlockTransactions:ValidateReceipt] - %s", err.Error())
			continue
		}
		if e := psc.ReceiptService.StoreReceipt(
			receipt,
			receipt.GetSenderPublicKey(),
			&chaintype.MainChain{},
		); e != nil {
			psc.Logger.Warnf("SendBlockTransactions: %s", e.Error())
		}
	}
	return err
}

func (psc *PeerServiceClient) RequestBlockTransactions(
	destPeer *model.Peer,
	transactionIDs []int64,
	chainType chaintype.ChainType,
	blockID int64,
) error {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pRequestBlockTransactionsClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pRequestBlockTransactionsClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return err
	}
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(20 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	_, err = p2pClient.RequestBlockTransactions(ctx, &model.RequestBlockTransactionsRequest{
		TransactionIDs: transactionIDs,
		ChainType:      chainType.GetTypeInt(),
		BlockID:        blockID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (psc *PeerServiceClient) RequestDownloadFile(
	destPeer *model.Peer,
	snapshotHash []byte,
	fileChunkNames []string,
) (*model.FileDownloadResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pRequestFileDownloadClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pRequestFileDownloadClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
	var (
		p2pClient      = service.NewP2PCommunicationClient(connection)
		ctx, cancelReq = psc.getDefaultContext(20 * time.Second)
	)
	defer func() {
		cancelReq()
	}()
	res, err := p2pClient.RequestFileDownload(ctx, &model.FileDownloadRequest{
		SnapshotHash:   snapshotHash,
		FileChunkNames: fileChunkNames,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetCumulativeDifficulty request the cumulative difficulty status of a node
func (psc *PeerServiceClient) GetCumulativeDifficulty(
	destPeer *model.Peer,
	chaintype chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetCumulativeDifficultyClient)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetCumulativeDifficultyClient)
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
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
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
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
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}
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
	psc.FeedbackStrategy.IncrementVarCount("P2POutgoingRequests")
	defer psc.FeedbackStrategy.DecrementVarCount("P2POutgoingRequests")

	connection, err := psc.GetConnection(destPeer)
	if err != nil {
		return nil, err
	}

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

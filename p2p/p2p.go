package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/handler"
	p2pService "github.com/zoobc/zoobc-core/p2p/service"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type (
	Peer2PeerServiceInterface interface {
		StartP2P(
			myAddress string,
			ownerAccountAddress string,
			peerPort uint32,
			nodeSecretPhrase string,
			queryExecutor query.ExecutorInterface,
			blockServices map[int32]coreService.BlockServiceInterface,
			mempoolServices map[int32]coreService.MempoolServiceInterface,
			fileService coreService.FileServiceInterface,
			nodeRegistrationService coreService.NodeRegistrationServiceInterface,
			nodeConfigurationService coreService.NodeConfigurationServiceInterface,
			nodeAddressInfoService coreService.NodeAddressInfoServiceInterface,
			observer *observer.Observer,
		)
		// exposed api list
		GetHostInfo() *model.Host
		GetResolvedPeers() map[string]*model.Peer
		GetUnresolvedPeers() map[string]*model.Peer
		GetPriorityPeers() map[string]*model.Peer

		// event listener that relate to p2p communication
		SendBlockListener() observer.Listener
		SendTransactionListener() observer.Listener
		RequestBlockTransactionsListener() observer.Listener
		SendBlockTransactionsListener() observer.Listener

		// internal p2p methods
		DownloadFilesFromPeer(fileChunksNames []string, retryCount uint32) (failed []string, err error)
	}
	Peer2PeerService struct {
		PeerExplorer             strategy.PeerExplorerStrategyInterface
		PeerServiceClient        client.PeerServiceClientInterface
		Logger                   *log.Logger
		TransactionUtil          transaction.UtilInterface
		FileService              coreService.FileServiceInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
	}
)

// NewP2PService to initialize peer to peer service wrapper
func NewP2PService(
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	logger *log.Logger,
	transactionUtil transaction.UtilInterface,
	fileService coreService.FileServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	nodeConfigurationService coreService.NodeConfigurationServiceInterface,
) (Peer2PeerServiceInterface, error) {
	return &Peer2PeerService{
		PeerServiceClient:        peerServiceClient,
		Logger:                   logger,
		PeerExplorer:             peerExplorer,
		TransactionUtil:          transactionUtil,
		FileService:              fileService,
		NodeRegistrationService:  nodeRegistrationService,
		NodeConfigurationService: nodeConfigurationService,
	}, nil
}

// StartP2P initiate all p2p dependencies and run all p2p thread service
func (s *Peer2PeerService) StartP2P(
	myAddress, ownerAccountAddress string,
	peerPort uint32,
	// nodeSecretPhrase needed in on-going feature by @iltoga
	nodeSecretPhrase string,
	queryExecutor query.ExecutorInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	mempoolServices map[int32]coreService.MempoolServiceInterface,
	fileService coreService.FileServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	nodeConfigurationService coreService.NodeConfigurationServiceInterface,
	nodeAddressInfoService coreService.NodeAddressInfoServiceInterface,
	observer *observer.Observer,
) {
	// peer to peer service layer | under p2p handler
	p2pServerService := p2pService.NewP2PServerService(
		nodeRegistrationService,
		fileService,
		nodeConfigurationService,
		nodeAddressInfoService,
		s.PeerExplorer,
		blockServices,
		mempoolServices,
		nodeSecretPhrase,
		observer,
	)
	// start listening on peer port
	go func() { // register handlers and listening to incoming p2p request
		var (
			grpcServer = grpc.NewServer(
				grpc.UnaryInterceptor(interceptor.NewServerInterceptor(
					s.Logger,
					ownerAccountAddress,
					map[codes.Code]string{
						codes.Unavailable:     "indicates the destination service is currently unavailable",
						codes.InvalidArgument: "indicates the argument request is invalid",
						codes.Unauthenticated: "indicates the request is unauthenticated",
					},
				)),
			)
		)

		service.RegisterP2PCommunicationServer(grpcServer, handler.NewP2PServerHandler(
			p2pServerService,
		))
		if err := grpcServer.Serve(p2pUtil.ServerListener(int(s.NodeConfigurationService.GetHost().GetInfo().GetPort()))); err != nil {
			s.Logger.Fatal(err.Error())
		}
	}()
	go s.PeerExplorer.Start()
}

// GetHostInfo exposed the p2p host information to the client
func (s *Peer2PeerService) GetHostInfo() *model.Host {
	return s.NodeConfigurationService.GetHost()
}

// GetResolvedPeers exposed current node resolved peer list
func (s *Peer2PeerService) GetResolvedPeers() map[string]*model.Peer {
	return s.PeerExplorer.GetResolvedPeers()
}

// GetUnresolvedPeers exposed current node unresolved peer list.
func (s *Peer2PeerService) GetUnresolvedPeers() map[string]*model.Peer {
	return s.PeerExplorer.GetUnresolvedPeers()
}

// GetPriorityPeers exposed current node priority peer list.
func (s *Peer2PeerService) GetPriorityPeers() map[string]*model.Peer {
	return s.PeerExplorer.GetPriorityPeers()
}

// SendBlockListener setup listener for send block to the list peer
func (s *Peer2PeerService) SendBlockListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args ...interface{}) {
			var (
				b         *model.Block
				chainType chaintype.ChainType
				ok        bool
			)
			b, ok = block.(*model.Block)
			if !ok {
				s.Logger.Fatalln("Block casting failures in SendBlockListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in SendBlockListener")
			}

			peers := s.PeerExplorer.GetResolvedPeers()
			for _, peer := range peers {
				go func(p *model.Peer) {
					if err := s.PeerServiceClient.SendBlock(p, b, chainType); err != nil {
						s.Logger.Errorf("SendBlockListener: %s", err)
					}
				}(peer)
			}
		},
	}
}

// SendTransactionListener setup listener for transaction to the list peer
func (s *Peer2PeerService) SendTransactionListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionBytes interface{}, args ...interface{}) {
			var (
				t         []byte
				chainType chaintype.ChainType
				ok        bool
			)
			t, ok = transactionBytes.([]byte)
			if !ok {
				s.Logger.Fatalln("transactionBytes casting failures in SendTransactionListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in SendTransactionListener")
			}
			peers := s.PeerExplorer.GetResolvedPeers()
			for _, peer := range peers {
				go func(p *model.Peer) {
					_ = s.PeerServiceClient.SendTransaction(p, t, chainType)

				}(peer)
			}
		},
	}
}

func (s *Peer2PeerService) RequestBlockTransactionsListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionIDs interface{}, args ...interface{}) {
			var (
				txIDs     = transactionIDs.([]int64)
				peer      *model.Peer
				chainType chaintype.ChainType
				blockID   int64
				ok        bool
			)

			// check number of arguments before casting the argument type
			if len(args) < 3 {
				s.Logger.Fatalln("number of needed arguments too few in RequestBlockTransactionsListener")
				return
			}

			blockID, ok = args[0].(int64)
			if !ok {
				s.Logger.Fatalln("blockID casting failures in RequestBlockTransactionsListener")
			}

			chainType, ok = args[1].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in RequestBlockTransactionsListener")
			}

			peer, ok = args[2].(*model.Peer)
			if !ok {
				s.Logger.Fatalln("peer casting failures in RequestBlockTransactionsListener")
			}

			go func(p *model.Peer) {
				_ = s.PeerServiceClient.RequestBlockTransactions(p, txIDs, chainType, blockID)
			}(peer)
		},
	}
}

func (s *Peer2PeerService) SendBlockTransactionsListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionsInterface interface{}, args ...interface{}) {
			var (
				txsBytes  [][]byte
				txs       []*model.Transaction
				chainType chaintype.ChainType
				peer      *model.Peer
				ok        bool
			)

			txs, ok = transactionsInterface.([]*model.Transaction)
			if !ok {
				s.Logger.Fatalln("Transaction casting failures in SendBlockTransactionsListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in SendBlockTransactionsListener")
			}

			peer, ok = args[1].(*model.Peer)
			if !ok {
				s.Logger.Fatalln("Peer casting failures in SendBlockTransactionsListener")
			}

			for _, tx := range txs {
				txByte, err := s.TransactionUtil.GetTransactionBytes(tx, true)
				if err != nil {
					continue
				}
				txsBytes = append(txsBytes, txByte)
			}
			go func(p *model.Peer) {
				_ = s.PeerServiceClient.SendBlockTransactions(p, txsBytes, chainType)
			}(peer)
		},
	}
}

// DownloadFilesFromPeer download a file from a random peer
func (s *Peer2PeerService) DownloadFilesFromPeer(fileChunksNames []string, maxRetryCount uint32) ([]string, error) {
	var (
		peer          *model.Peer
		resolvedPeers = s.PeerExplorer.GetResolvedPeers()
		peerKey       string
		retryCount    uint32
	)
	// Retry downloading from different peers until all chunks are downloaded or retry limit is reached
	if len(resolvedPeers) < 1 {
		return nil, blocker.NewBlocker(blocker.P2PPeerError, "no resolved peer can be found")
	}
	// convert the slice to a map to make it easier to find elements in it
	fileChunkNamesMap := make(map[string]string)
	for _, s := range fileChunksNames {
		fileChunkNamesMap[s] = s
	}
	fileChunksToDownload := fileChunksNames
	r := util.GetFastRandomSeed()
	for retryCount < maxRetryCount+1 {
		retryCount++

		// randomly select one of the resolved peers to download files from
		// (no need for secure random here. we just want to get a quick pseudo random index)
		randomIdx := int(util.GetFastRandom(r, len(resolvedPeers)))
		if randomIdx != 0 {
			randomIdx %= len(resolvedPeers)
		}
		idx := 0
		for peerKey, peer = range resolvedPeers {
			if idx == randomIdx {
				// remove selected peer from map to avoid selecting it again
				delete(resolvedPeers, peerKey)
				break
			}
			idx++
		}

		// download the files
		fileDownloadResponse, err := s.PeerServiceClient.RequestDownloadFile(peer, fileChunksToDownload)
		if err != nil {
			log.Warnf("error download: %v\nchunks: %v\npeer: %v\n", err, fileChunksToDownload, peer)
			return nil, err
		}

		// check first that all chunks returned are valid
		skipFilesFromPeer := false
		for _, fileChunk := range fileDownloadResponse.GetFileChunks() {
			fileChunkComputedName := s.FileService.GetFileNameFromBytes(fileChunk)
			if _, ok := fileChunkNamesMap[fileChunkComputedName]; !ok {
				s.Logger.Errorf("peer returned an invalid file chunk: %s", fileChunkComputedName)
				skipFilesFromPeer = true
				break
			}
		}
		// never trust a peer that returns wrong data, just skip all files downloaded from it
		if skipFilesFromPeer {
			continue
		}

		// save downloaded chunks to storage as soon as possible to avoid keeping in memory large arrays
		for _, fileChunk := range fileDownloadResponse.GetFileChunks() {
			fileChunkComputedName := s.FileService.GetFileNameFromBytes(fileChunk)
			err = s.FileService.SaveBytesToFile(s.FileService.GetDownloadPath(), fileChunkComputedName, fileChunk)
			if err != nil {
				s.Logger.Errorf("failed saving file to storage: %s", err)
				return nil, err
			}
		}

		// set next files to download = previous files that failed to download
		fileChunksToDownload = fileDownloadResponse.GetFailed()
		// break download loop either if all files have been successfully downloaded or there are no more peers to connect to
		if len(fileChunksToDownload) == 0 || len(resolvedPeers) == 0 {
			if len(fileChunksToDownload) > 0 && len(resolvedPeers) == 0 {
				s.Logger.Debug("no more resolved peers to download files from. Already tried them all!")
			}
			break
		}
	}

	return fileChunksToDownload, nil
}

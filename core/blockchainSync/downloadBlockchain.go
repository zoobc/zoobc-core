package blockchainSync

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/constant"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

func (bss *BlockchainSyncService) Start(runNext chan bool) {
	if bss.ChainType == nil {
		panic("no chaintype")
	}
	if bss.P2pService == nil {
		panic("no p2p service defined")
	}
	bss.GetMoreBlocksThread(runNext)
}

func (bss *BlockchainSyncService) GetMoreBlocksThread(runNext chan bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case download := <-runNext:
			if download {
				go bss.getMoreBlocks(runNext)
			}
		case <-sigs:
			return
		}
	}
}

func (bss BlockchainSyncService) getMoreBlocks(runNext chan bool) {
	log.Info("Get more blocks...")
	// notify observer about start of blockchain download of this specific chain

	lastBlock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		panic(fmt.Sprintf("failed to start getMoreBlocks go routine: %v", err))
	}
	if lastBlock == nil {
		panic("There is no genesis block found")
	}
	initialHeight := lastBlock.Height
	for bss.NeedGetMoreBlocks {
		// observers.BlockNotifier().Notify(observers.BLOCK_DOWNLOADING, nil, bss.Chaintype)
		currentLastBlock, err := bss.BlockService.GetLastBlock()
		if err != nil {
			log.Error("failed to get the current last block")
			continue
		}
		currentHeight := currentLastBlock.Height
		err = bss.getPeerBlockchainInfo()
		if err != nil {
			log.Warnf("\nfailed to getPeerBlockchainInfo: %v\n\n", err)
		}
		afterDownloadLastBlock, err := bss.BlockService.GetLastBlock()
		if err != nil {
			log.Error("failed to get the last block state after block download")
			continue
		}
		heightAfterDownload := afterDownloadLastBlock.Height
		if currentHeight > 0 && currentHeight == heightAfterDownload {
			bss.IsDownloading = false
			log.Printf("Finished %s blockchain download: %d blocks pulled", bss.ChainType.GetName(), heightAfterDownload-initialHeight)
			// observers.BlockNotifier().Notify(observers.BLOCK_DOWNLOAD_FINISH, heightAfterDownload, bs.Chaintype)
			break
		}
		break
	}

	// bs.RestorePrunableData()

	// TODO: Handle interruption and other exceptions
	time.Sleep(constant.GetMoreBlocksDelay * time.Second)
	runNext <- true
}

func (bss *BlockchainSyncService) getPeerBlockchainInfo() error {
	bss.PeerHasMore = true
	peer := bss.P2pService.GetAnyResolvedPeer()
	if peer == nil {
		return errors.New("no connected peer can be found")
	}
	peerCumulativeDifficultyResponse, err := bss.P2pService.GetCumulativeDifficulty(peer, bss.ChainType)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get Cumulative Difficulty of peer %v: %v", peer.Info.Address, err))
	}

	peerCumulativeDifficulty, _ := new(big.Int).SetString(peerCumulativeDifficultyResponse.CumulativeDifficulty, 10)
	peerHeight := peerCumulativeDifficultyResponse.Height

	lastBlock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	lastBlockCumulativeDifficulty, _ := new(big.Int).SetString(lastBlock.CumulativeDifficulty, 10)
	lastBlockHeight := lastBlock.Height
	lastBlockID := lastBlock.ID

	if peerCumulativeDifficulty.Cmp(lastBlockCumulativeDifficulty) <= 0 {
		return errors.New("peer's cumulative difficulty is lower/same with the current node's")
	}

	// this is to set the status of download blockchain process
	if peerHeight > 0 {
		bss.LastBlockchainFeeder = peer
		bss.LastBlockchainFeederHeight = peerHeight
	}

	commonMilestoneBlockID := bss.ChainType.GetGenesisBlockID()
	if lastBlockID != commonMilestoneBlockID {
		var err error
		commonMilestoneBlockID, err = bss.getPeerCommonBlockID(peer)
		if err != nil {
			return err
		}
	}

	chainBlockIds := bss.getBlockIdsAfterCommon(peer, commonMilestoneBlockID)
	if len(chainBlockIds) < 2 || !bss.PeerHasMore {
		return errors.New("the peer does not have more updated chain")
	}

	commonBlockID := chainBlockIds[0]
	commonBlock, err := bss.BlockService.GetBlockByID(commonBlockID)
	if err != nil {
		log.Warnf("common block %v not found, milestone %v", commonBlockID, commonMilestoneBlockID)
		return err
	}
	if commonBlock == nil || lastBlockHeight-commonBlock.GetHeight() >= 720 {
		return errors.New("invalid common block")
	}

	if !bss.IsDownloading && bss.LastBlockchainFeederHeight-commonBlock.GetHeight() > 10 {
		log.Println("Blockchain download in progress")
		bss.IsDownloading = true
	}

	bss.BlockService.ChainWriteLock()
	defer bss.BlockService.ChainWriteUnlock()

	bss.downloadFromPeer(peer, commonBlock, chainBlockIds)

	// TODO: analyze the importance of this mechanism
	bss.confirmBlockchainState(peer, commonMilestoneBlockID)
	newLastBlock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return err
	}

	if lastBlockID == newLastBlock.ID {
		log.Println("Did not accept peers's blocks, back to our own fork")
	}
	return nil
}

func (bss *BlockchainSyncService) confirmBlockchainState(peer *model.Peer, commonMilestoneBlockID int64) error {
	confirmations := int32(0)
	// counting the confirmations of the common block received with other peers he knows
	for _, peerToCheck := range bss.P2pService.GetResolvedPeers() {
		if confirmations >= constant.DefaultNumberOfForkConfirmations {
			break
		}

		// if the host found other peer with better difficulty
		otherPeerChainBlockIds := bss.getBlockIdsAfterCommon(peer, commonMilestoneBlockID)
		currentLastBlock, err := bss.BlockService.GetLastBlock()
		if err != nil {
			return err
		}
		currentLastBlockCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
		if otherPeerChainBlockIds[0] == currentLastBlock.ID {
			confirmations++
			continue
		}
		otherPeerCommonBlock, err := bss.BlockService.GetBlockByID(otherPeerChainBlockIds[0])
		if err != nil {
			return err
		}
		if currentLastBlock.Height-otherPeerCommonBlock.Height >= 720 {
			continue
		}

		otherPeerCumulativeDifficultyResponse, err := bss.P2pService.GetCumulativeDifficulty(peerToCheck, bss.ChainType)
		if err != nil || otherPeerCumulativeDifficultyResponse.CumulativeDifficulty == "" {
			continue
		}

		otherPeerCumulativeDifficulty, _ := new(big.Int).SetString(otherPeerCumulativeDifficultyResponse.CumulativeDifficulty, 10)
		if otherPeerCumulativeDifficulty.Cmp(currentLastBlockCumulativeDifficulty) <= 0 {
			continue
		}

		log.Println("Found a peer with better difficulty")
		bss.downloadFromPeer(peerToCheck, otherPeerCommonBlock, otherPeerChainBlockIds)
	}
	log.Println("Got ", confirmations, " confirmations")
	return nil
}

func (bss *BlockchainSyncService) downloadFromPeer(feederPeer *model.Peer, commonBlock *model.Block, chainBlockIds []int64) error {
	var peersTobeDeactivated []*model.Peer
	segSize := constant.BlockDownloadSegSize

	stop := uint32(len(chainBlockIds) - 1)

	var peersSlice []*model.Peer
	for _, peer := range bss.P2pService.GetResolvedPeers() {
		peersSlice = append(peersSlice, peer)
	}

	if len(peersSlice) < 1 {
		return errors.New("the host does not have resolved peers")
	}

	nextPeerIdx := int(commonUtil.GetSecureRandom()) % len(peersSlice)
	peerUsed := feederPeer
	blocksSegments := [][]*model.Block{}

	for start := uint32(0); start < stop; start = start + segSize {
		if start != uint32(0) {
			peerUsed = peersSlice[nextPeerIdx]
			nextPeerIdx = nextPeerIdx + 1
			if nextPeerIdx >= len(peersSlice) {
				nextPeerIdx = 0
			}
		}

		// TODO: apply retry mechanism
		startTime := time.Now()
		nextBlocks, err := bss.getNextBlocks(constant.BlockDownloadSegSize, peerUsed, chainBlockIds, start, commonUtil.MinUint32(start+segSize, stop))
		if err != nil {
			log.Warn(err)
			return err
		}
		elapsedTime := time.Since(startTime)
		if elapsedTime > constant.MaxResponseTime {
			peersTobeDeactivated = append(peersTobeDeactivated, peerUsed)
		}

		if len(nextBlocks) < 1 {
			log.Warnf("disconnecting with peer %v for not responding correctly in getting the next blocks\n", peerUsed.Info.Address)
			peersTobeDeactivated = append(peersTobeDeactivated, peerUsed)
			continue
		}
		blocksSegments = append(blocksSegments, nextBlocks)
	}

	blocksToBeProcessed := []*model.Block{}
	for _, blockSegment := range blocksSegments {
		for i := 0; i < len(blockSegment); i++ {
			if coreUtil.IsBlockIDExist(chainBlockIds, blockSegment[i].ID) {
				blocksToBeProcessed = append(blocksToBeProcessed, blockSegment[i])
			}
		}
	}

	for _, peer := range peersTobeDeactivated {
		bss.P2pService.DisconnectPeer(peer)
	}

	forkBlocks := []*model.Block{}
	for _, block := range blocksToBeProcessed {
		if block.Height == 0 {
			continue
		}
		lastBlock, _ := bss.BlockService.GetLastBlock()
		previousBlockID := coreUtil.GetBlockIDFromHash(block.PreviousBlockHash)
		if lastBlock.ID == previousBlockID {
			err := bss.BlockService.PushBlock(lastBlock, block, false)
			if err != nil {
				// TODO: analyze the mechanism of blacklisting peer here
				// bss.P2pService.Blacklist(peer)
				log.Warnln("failed to push block from peer:", err)
			}
		} else {
			forkBlocks = append(forkBlocks, block)
		}
	}

	if len(forkBlocks) > 0 {
		// log.Println("processing fork blocks %v", forkBlocks)
		// ProcessFork(forkBlocks,commonBlock)
	}
	return nil
}

func (bss *BlockchainSyncService) getPeerCommonBlockID(peer *model.Peer) (int64, error) {
	lastMilestoneBlockID := int64(0)
	lastBlock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		log.Warnf("failed to get blockchain last block: %v\n", err)
		return 0, err
	}
	lastBlockID := lastBlock.ID
	for {
		commonMilestoneBlockIDResponse, err := bss.P2pService.GetCommonMilestoneBlockIDs(peer, bss.ChainType, lastBlockID, lastMilestoneBlockID)
		if err != nil {
			log.Warnf("failed to get common milestone from the peer: %v\n", err)
			bss.P2pService.DisconnectPeer(peer)
			return 0, err
		}
		for _, blockID := range commonMilestoneBlockIDResponse.BlockIds {
			_, err := bss.BlockService.GetBlockByID(blockID)
			if err == nil {
				return blockID, nil
			}
			lastMilestoneBlockID = blockID
		}
	}
	return 0, nil
}

func (bss *BlockchainSyncService) getBlockIdsAfterCommon(peer *model.Peer, commonMilestoneBlockID int64) []int64 {
	blockIds, err := bss.P2pService.GetNextBlockIDs(peer, bss.ChainType, commonMilestoneBlockID, constant.PeerGetBlocksLimit)
	if err != nil {
		return []int64{}
	}

	newBlockIDIdx := 0
	for idx, blockID := range blockIds.BlockIds {
		_, err := bss.BlockService.GetBlockByID(blockID)
		// mark the new block ID starting where it is not found
		if err != nil {
			break
		}
		newBlockIDIdx = idx
	}
	if newBlockIDIdx >= len(blockIds.BlockIds) {
		return []int64{}
	}
	return blockIds.BlockIds[newBlockIDIdx:]
}

func (bss *BlockchainSyncService) getNextBlocks(maxNextBlocks uint32, peerUsed *model.Peer, blockIds []int64, start, stop uint32) ([]*model.Block, error) {
	blocks := []*model.Block{}
	nextBlocksResponse, err := bss.P2pService.GetNextBlocks(peerUsed, bss.ChainType, blockIds[start:stop], blockIds[start])
	nextBlocks := nextBlocksResponse.NextBlocks
	nextBlocksLength := uint32(len(nextBlocks))
	if nextBlocksLength > maxNextBlocks {
		return nil, fmt.Errorf("too many blocks returned (%d blocks), possibly a rogue peer %v", nextBlocksLength, peerUsed.Info.Address)
	}
	if nextBlocks == nil || err != nil || nextBlocksLength == 0 {
		return nil, err
	}
	if len(nextBlocks) > 0 {
		return nextBlocks, nil
	}
	return blocks, nil
}

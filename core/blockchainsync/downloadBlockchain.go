package blockchainsync

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"

	"github.com/zoobc/zoobc-core/common/constant"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockchainDownloadInterface interface {
		SetIsDownloading(newValue bool)
		IsDownloadFinish(currentLastBlock *model.Block) bool
		GetPeerBlockchainInfo() (*PeerBlockchainInfo, error)
		DownloadFromPeer(feederPeer *model.Peer, chainBlockIds []int64, commonBlock *model.Block) (*PeerForkInfo, error)
		ConfirmWithPeer(peerToCheck *model.Peer, commonMilestoneBlockID int64) ([]int64, error)
	}
	BlockchainDownloader struct {
		IsDownloading bool // only for status
		PeerHasMore   bool
		ChainType     chaintype.ChainType

		BlockService      service.BlockServiceInterface
		PeerServiceClient client.PeerServiceClientInterface
		PeerExplorer      strategy.PeerExplorerStrategyInterface
	}

	PeerBlockchainInfo struct {
		Peer                   *model.Peer
		ChainBlockIds          []int64
		CommonBlock            *model.Block
		CommonMilestoneBlockID int64
	}

	PeerForkInfo struct {
		ForkBlocks []*model.Block
		FeederPeer *model.Peer
	}
)

func (bd *BlockchainDownloader) IsDownloadFinish(currentLastBlock *model.Block) bool {
	currentHeight := currentLastBlock.Height
	currentCumulativeDifficulty := currentLastBlock.CumulativeDifficulty
	afterDownloadLastBlock, err := bd.BlockService.GetLastBlock()
	if err != nil {
		log.Warnf("failed to get the last block state after block download")
		return false
	}
	heightAfterDownload := afterDownloadLastBlock.Height
	cumulativeDifficultyAfterDownload := afterDownloadLastBlock.CumulativeDifficulty
	if currentHeight > 0 && currentHeight == heightAfterDownload && currentCumulativeDifficulty == cumulativeDifficultyAfterDownload {
		return true
	}
	return false
}

func (bd *BlockchainDownloader) SetIsDownloading(newValue bool) {
	bd.IsDownloading = newValue
}

func (bd *BlockchainDownloader) GetPeerBlockchainInfo() (*PeerBlockchainInfo, error) {
	bd.PeerHasMore = true
	peer := bd.PeerExplorer.GetAnyResolvedPeer()
	if peer == nil {
		return nil, errors.New("no connected peer can be found")
	}
	peerCumulativeDifficultyResponse, err := bd.PeerServiceClient.GetCumulativeDifficulty(peer, bd.ChainType)
	if err != nil {
		return nil, fmt.Errorf("failed to get Cumulative Difficulty of peer %v: %v", peer.Info.Address, err)
	}

	peerCumulativeDifficulty, _ := new(big.Int).SetString(peerCumulativeDifficultyResponse.CumulativeDifficulty, 10)
	peerHeight := peerCumulativeDifficultyResponse.Height

	lastBlock, err := bd.BlockService.GetLastBlock()
	if err != nil {
		return nil, err
	}
	lastBlockCumulativeDifficulty, _ := new(big.Int).SetString(lastBlock.CumulativeDifficulty, 10)
	lastBlockHeight := lastBlock.Height
	lastBlockID := lastBlock.ID

	if peerCumulativeDifficulty.Cmp(lastBlockCumulativeDifficulty) <= 0 {
		return nil, errors.New("peer's cumulative difficulty is lower/same with the current node's")
	}

	commonMilestoneBlockID := bd.ChainType.GetGenesisBlockID()
	if lastBlockID != commonMilestoneBlockID {
		var err error
		commonMilestoneBlockID, err = bd.getPeerCommonBlockID(peer)
		if err != nil {
			return nil, err
		}
	}

	chainBlockIds := bd.getBlockIdsAfterCommon(peer, commonMilestoneBlockID)
	if len(chainBlockIds) < 2 || !bd.PeerHasMore {
		return nil, errors.New("the peer does not have more updated chain")
	}

	commonBlockID := chainBlockIds[0]
	commonBlock, err := bd.BlockService.GetBlockByID(commonBlockID)
	if err != nil {
		log.Warnf("common block %v not found, milestone block id: %v", commonBlockID, commonMilestoneBlockID)
		return nil, err
	}
	if commonBlock == nil || lastBlockHeight-commonBlock.GetHeight() >= constant.MinRollbackBlocks {
		return nil, errors.New("invalid common block")
	}

	if !bd.IsDownloading && peerHeight-commonBlock.GetHeight() > 10 {
		log.Println("Blockchain download in progress")
		bd.IsDownloading = true
	}

	return &PeerBlockchainInfo{
		Peer:                   peer,
		ChainBlockIds:          chainBlockIds,
		CommonBlock:            commonBlock,
		CommonMilestoneBlockID: commonMilestoneBlockID,
	}, nil
}

// ConfirmWithPeer confirms the state of our blockchain with other peer
// returns (otherPeerChainBlockIds: []int64, error)
// if otherPeerChainBlockIds has member, it means that there are blocks to download from the peer
func (bd *BlockchainDownloader) ConfirmWithPeer(peerToCheck *model.Peer, commonMilestoneBlockID int64) ([]int64, error) {
	// if the host found other peer with better difficulty
	otherPeerChainBlockIds := bd.getBlockIdsAfterCommon(peerToCheck, commonMilestoneBlockID)
	currentLastBlock, err := bd.BlockService.GetLastBlock()
	if err != nil {
		return []int64{}, err
	}
	currentLastBlockCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
	if len(otherPeerChainBlockIds) < 1 || otherPeerChainBlockIds[0] == currentLastBlock.ID {
		return []int64{}, nil
	}
	otherPeerCommonBlock, err := bd.BlockService.GetBlockByID(otherPeerChainBlockIds[0])
	if err != nil {
		return []int64{}, err
	}
	if currentLastBlock.Height-otherPeerCommonBlock.Height >= constant.MinRollbackBlocks {
		return []int64{}, blocker.NewBlocker(blocker.ChainValidationErr,
			fmt.Sprintf("Peer %s common block differs by more than %d blocks compared to our blockchain", peerToCheck.GetInfo().Address,
				constant.MinRollbackBlocks))
	}

	otherPeerCumulativeDifficultyResponse, err := bd.PeerServiceClient.GetCumulativeDifficulty(peerToCheck, bd.ChainType)
	if err != nil || otherPeerCumulativeDifficultyResponse.CumulativeDifficulty == "" {
		return []int64{}, blocker.NewBlocker(blocker.ChainValidationErr, fmt.Sprintf("error in peer %s cumulative difficulty",
			peerToCheck.GetInfo().Address))
	}

	otherPeerCumulativeDifficulty, _ := new(big.Int).SetString(otherPeerCumulativeDifficultyResponse.CumulativeDifficulty, 10)
	if otherPeerCumulativeDifficulty.Cmp(currentLastBlockCumulativeDifficulty) <= 0 {
		return []int64{}, blocker.NewBlocker(blocker.ChainValidationErr, fmt.Sprintf("peer's cumulative difficulty %s is lower than ours",
			peerToCheck.GetInfo().Address))
	}

	return otherPeerChainBlockIds, nil
}

func (bd *BlockchainDownloader) DownloadFromPeer(feederPeer *model.Peer, chainBlockIds []int64,
	commonBlock *model.Block) (*PeerForkInfo, error) {
	var peersTobeDeactivated []*model.Peer
	segSize := constant.BlockDownloadSegSize

	stop := uint32(len(chainBlockIds))

	var peersSlice []*model.Peer
	for _, peer := range bd.PeerExplorer.GetResolvedPeers() {
		peersSlice = append(peersSlice, peer)
	}

	if len(peersSlice) < 1 {
		return nil, errors.New("the host does not have resolved peers")
	}

	nextPeerIdx := int(commonUtil.GetSecureRandom()) % len(peersSlice)
	peerUsed := feederPeer
	blocksSegments := [][]*model.Block{}

	for start := uint32(0); start < stop; start += segSize {
		if start != uint32(0) {
			peerUsed = peersSlice[nextPeerIdx]
			nextPeerIdx++
			if nextPeerIdx >= len(peersSlice) {
				nextPeerIdx = 0
			}
		}

		// TODO: apply retry mechanism
		startTime := time.Now()
		nextBlocks, err := bd.getNextBlocks(constant.BlockDownloadSegSize, peerUsed, chainBlockIds,
			start, commonUtil.MinUint32(start+segSize, stop))
		if err != nil {
			log.Warn(err)
			return nil, err
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
		bd.PeerExplorer.DisconnectPeer(peer)
	}

	forkBlocks := []*model.Block{}
	for _, block := range blocksToBeProcessed {
		if block.Height == 0 {
			continue
		}
		lastBlock, err := bd.BlockService.GetLastBlock()
		if err != nil {
			return nil, err
		}
		previousBlockID := coreUtil.GetBlockIDFromHash(block.PreviousBlockHash)
		if lastBlock.ID == previousBlockID {
			err := bd.BlockService.PushBlock(lastBlock, block, false)
			if err != nil {
				// TODO: analyze the mechanism of blacklisting peer here
				// bd.P2pService.Blacklist(peer)
				log.Warnln("failed to push block from peer:", err)
			}
		} else {
			forkBlocks = append(forkBlocks, block)

		}
	}

	return &PeerForkInfo{
		ForkBlocks: forkBlocks,
		FeederPeer: feederPeer,
	}, nil

}

func (bd *BlockchainDownloader) getPeerCommonBlockID(peer *model.Peer) (int64, error) {
	lastMilestoneBlockID := int64(0)
	lastBlock, err := bd.BlockService.GetLastBlock()
	if err != nil {
		log.Errorf("failed to get blockchain last block: %v\n", err)
		return 0, err
	}
	lastBlockID := lastBlock.ID
	for {
		commonMilestoneBlockIDResponse, err := bd.PeerServiceClient.GetCommonMilestoneBlockIDs(
			peer, bd.ChainType, lastBlockID, lastMilestoneBlockID,
		)
		if err != nil {
			log.Warnf("failed to get common milestone from the peer: %v\n", err)
			bd.PeerExplorer.DisconnectPeer(peer)
			return 0, err
		}

		if commonMilestoneBlockIDResponse.Last {
			bd.PeerHasMore = false
		}

		for _, blockID := range commonMilestoneBlockIDResponse.BlockIds {
			_, err := bd.BlockService.GetBlockByID(blockID)
			if err == nil {
				return blockID, nil
			}
			errCasted := err.(blocker.Blocker)
			if errCasted.Type != blocker.BlockNotFoundErr {
				return 0, err
			}
			lastMilestoneBlockID = blockID
		}
	}
}

func (bd *BlockchainDownloader) getBlockIdsAfterCommon(peer *model.Peer, commonMilestoneBlockID int64) []int64 {
	blockIds, err := bd.PeerServiceClient.GetNextBlockIDs(peer, bd.ChainType, commonMilestoneBlockID, constant.PeerGetBlocksLimit)
	if err != nil {
		return []int64{}
	}

	newBlockIDIdx := 0
	for idx, blockID := range blockIds.BlockIds {
		_, err := bd.BlockService.GetBlockByID(blockID)
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

func (bd *BlockchainDownloader) getNextBlocks(maxNextBlocks uint32, peerUsed *model.Peer,
	blockIds []int64, start, stop uint32) ([]*model.Block, error) {
	var blocks []*model.Block
	nextBlocksResponse, err := bd.PeerServiceClient.GetNextBlocks(peerUsed, bd.ChainType, blockIds[start:stop], blockIds[start])
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

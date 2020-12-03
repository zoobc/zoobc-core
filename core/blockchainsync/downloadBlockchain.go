package blockchainsync

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

type (
	BlockchainDownloadInterface interface {
		IsDownloadFinish(currentLastBlock *model.Block) bool
		GetPeerBlockchainInfo() (*PeerBlockchainInfo, error)
		DownloadFromPeer(feederPeer *model.Peer, chainBlockIds []int64, commonBlock *model.Block) (*PeerForkInfo, error)
		ConfirmWithPeer(peerToCheck *model.Peer, commonMilestoneBlockID int64) ([]int64, error)
	}
	BlockchainDownloader struct {
		PeerHasMore             bool
		ChainType               chaintype.ChainType
		BlockService            service.BlockServiceInterface
		PeerServiceClient       client.PeerServiceClientInterface
		PeerExplorer            strategy.PeerExplorerStrategyInterface
		Logger                  *log.Logger
		BlockchainStatusService service.BlockchainStatusServiceInterface
		firstDownloadCounter    int32
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

func NewBlockchainDownloader(
	blockService service.BlockServiceInterface,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	logger *log.Logger,
	blockchainStatusService service.BlockchainStatusServiceInterface,
) *BlockchainDownloader {
	return &BlockchainDownloader{
		ChainType:               blockService.GetChainType(),
		BlockService:            blockService,
		PeerServiceClient:       peerServiceClient,
		PeerExplorer:            peerExplorer,
		Logger:                  logger,
		BlockchainStatusService: blockchainStatusService,
	}
}

func (bd *BlockchainDownloader) IsDownloadFinish(currentLastBlock *model.Block) bool {
	currentHeight := currentLastBlock.Height
	currentCumulativeDifficulty := currentLastBlock.CumulativeDifficulty
	afterDownloadLastBlock, err := bd.BlockService.GetLastBlock()
	if err != nil {
		bd.Logger.Warnf("failed to get the last block state after block download: %v\n", err)
		return false
	}
	heightAfterDownload := afterDownloadLastBlock.Height
	cumulativeDifficultyAfterDownload := afterDownloadLastBlock.CumulativeDifficulty
	if currentHeight == heightAfterDownload && currentCumulativeDifficulty == cumulativeDifficultyAfterDownload {
		if currentHeight == 0 {
			bd.firstDownloadCounter++
			if bd.firstDownloadCounter >= constant.MaxResolvedPeers {
				bd.firstDownloadCounter = 0
				return true
			}
		}
		bd.firstDownloadCounter = 0
		return true
	}
	return false
}

func (bd *BlockchainDownloader) GetPeerBlockchainInfo() (*PeerBlockchainInfo, error) {
	var (
		err                              error
		peerCumulativeDifficultyResponse *model.GetCumulativeDifficultyResponse
		lastBlock, commonBlock           *model.Block
	)

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 30)
	bd.PeerHasMore = true
	peer := bd.PeerExplorer.GetAnyResolvedPeer()
	if peer == nil {
		return nil, blocker.NewBlocker(blocker.P2PPeerError, "no connected peer can be found")
	}
	peerCumulativeDifficultyResponse, err = bd.PeerServiceClient.GetCumulativeDifficulty(peer, bd.ChainType)
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 31)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 32)
		return &PeerBlockchainInfo{
				Peer:        peer,
				CommonBlock: commonBlock,
			}, blocker.NewBlocker(blocker.AppErr,
				fmt.Sprintf("failed to get Cumulative Difficulty of peer %v: %v", peer.Info.Address, err))
	}
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 33)
	peerCumulativeDifficulty, _ := new(big.Int).SetString(peerCumulativeDifficultyResponse.CumulativeDifficulty, 10)
	peerHeight := peerCumulativeDifficultyResponse.Height
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 34)
	lastBlock, err = bd.BlockService.GetLastBlock()
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 35)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 36)
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	lastBlockCumulativeDifficulty, _ := new(big.Int).SetString(lastBlock.CumulativeDifficulty, 10)
	lastBlockHeight := lastBlock.Height
	lastBlockID := lastBlock.ID

	if peerCumulativeDifficulty == nil || lastBlockCumulativeDifficulty == nil ||
		peerCumulativeDifficulty.Cmp(lastBlockCumulativeDifficulty) <= 0 {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 37)
		return &PeerBlockchainInfo{
				Peer:        peer,
				CommonBlock: commonBlock,
			}, blocker.NewBlocker(blocker.ChainValidationErr,
				fmt.Sprintf(
					"cumulative difficulty is lower/same with the current node's. Own: %s/%d, Peer: %s/%d",
					lastBlock.CumulativeDifficulty, lastBlock.Height,
					peerCumulativeDifficultyResponse.CumulativeDifficulty,
					peerCumulativeDifficultyResponse.Height,
				),
			)
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 38)
	commonMilestoneBlockID := bd.ChainType.GetGenesisBlockID()
	if lastBlockID != commonMilestoneBlockID {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 39)
		commonMilestoneBlockID, err = bd.getPeerCommonBlockID(peer)
		if err != nil {
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 40)
			return nil, err
		}
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 41)
	chainBlockIds := bd.getBlockIdsAfterCommon(peer, commonMilestoneBlockID)
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 42)
	if int32(len(chainBlockIds)) < constant.MinimumPeersBlocksToDownload || !bd.PeerHasMore {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 43)
		return &PeerBlockchainInfo{
			Peer:        peer,
			CommonBlock: commonBlock,
		}, blocker.NewBlocker(blocker.ChainValidationErr, "the peer does not have more updated chain")
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 44)
	commonBlockID := chainBlockIds[0]
	commonBlock, err = bd.BlockService.GetBlockByID(commonBlockID, false)
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 45)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 46)
		return &PeerBlockchainInfo{
				Peer:        peer,
				CommonBlock: commonBlock,
			}, blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("common block %v not found, milestone block id: %v",
				commonBlockID, commonMilestoneBlockID))
	}
	if commonBlock == nil || lastBlockHeight-commonBlock.GetHeight() >= constant.MinRollbackBlocks {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 47)
		return &PeerBlockchainInfo{
			Peer:        peer,
			CommonBlock: commonBlock,
		}, blocker.NewBlocker(blocker.AppErr, "invalid common block")
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 48)
	if !bd.BlockchainStatusService.IsDownloading(bd.ChainType) && peerHeight-commonBlock.GetHeight() > 10 {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 49)
		bd.Logger.Info("Blockchain download in progress")
		bd.BlockchainStatusService.SetIsDownloading(bd.ChainType, true)
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
	var (
		err                                    error
		currentLastBlock, otherPeerCommonBlock *model.Block
		otherPeerCumulativeDifficultyResponse  *model.GetCumulativeDifficultyResponse
	)

	// if the host found other peer with better difficulty
	otherPeerChainBlockIds := bd.getBlockIdsAfterCommon(peerToCheck, commonMilestoneBlockID)
	currentLastBlock, err = bd.BlockService.GetLastBlock()
	if err != nil {
		return []int64{}, err
	}
	currentLastBlockCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
	if len(otherPeerChainBlockIds) < 1 || otherPeerChainBlockIds[0] == currentLastBlock.ID {
		return []int64{}, nil
	}
	otherPeerCommonBlock, err = bd.BlockService.GetBlockByID(otherPeerChainBlockIds[0], false)
	if err != nil {
		return []int64{}, err
	}
	if currentLastBlock.Height-otherPeerCommonBlock.Height >= constant.MinRollbackBlocks {
		return []int64{}, blocker.NewBlocker(blocker.ChainValidationErr,
			fmt.Sprintf("Peer %s common block differs by more than %d blocks compared to our blockchain", peerToCheck.GetInfo().Address,
				constant.MinRollbackBlocks))
	}

	otherPeerCumulativeDifficultyResponse, err = bd.PeerServiceClient.GetCumulativeDifficulty(peerToCheck, bd.ChainType)
	if err != nil || otherPeerCumulativeDifficultyResponse.CumulativeDifficulty == "" {
		return []int64{}, blocker.NewBlocker(blocker.ChainValidationErr, fmt.Sprintf("error in peer %s cumulative difficulty",
			peerToCheck.GetInfo().Address))
	}

	otherPeerCumulativeDifficulty, _ := new(big.Int).SetString(otherPeerCumulativeDifficultyResponse.CumulativeDifficulty, 10)
	if otherPeerCumulativeDifficulty.Cmp(currentLastBlockCumulativeDifficulty) <= 0 {
		return []int64{}, blocker.NewBlocker(blocker.ChainValidationErr, fmt.Sprintf("peer's cumulative difficulty %s:%v is lower than ours",
			peerToCheck.GetInfo().Address, peerToCheck.GetInfo().Port))
	}

	return otherPeerChainBlockIds, nil
}

func (bd *BlockchainDownloader) DownloadFromPeer(feederPeer *model.Peer, chainBlockIds []int64,
	commonBlock *model.Block) (*PeerForkInfo, error) {
	var (
		peersTobeDeactivated   []*model.Peer
		peersSlice             []*model.Peer
		forkBlocks             []*model.Block
		segSize                = constant.BlockDownloadSegSize
		stop                   = uint32(len(chainBlockIds))
		numberOfErrorsInACycle int
	)
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 50)

	for _, peer := range bd.PeerExplorer.GetResolvedPeers() {
		peersSlice = append(peersSlice, peer)
	}

	if len(peersSlice) < 1 {
		return nil, errors.New("the host does not have resolved peers")
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 51)
	initialPeerIdx := int(commonUtil.GetSecureRandom()) % len(peersSlice)
	nextPeerIdx := initialPeerIdx
	peerUsed := feederPeer
	blocksSegments := [][]*model.Block{}

	for start := uint32(0); start < stop; {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 52)
		if start != uint32(0) {
			peerUsed = peersSlice[nextPeerIdx]
			nextPeerIdx++
			if nextPeerIdx >= len(peersSlice) {
				nextPeerIdx = 0
			}
			if nextPeerIdx == initialPeerIdx {
				numberOfErrorsInACycle = 0
			}
		}

		// TODO: apply retry mechanism
		startTime := time.Now()
		nextBlocks, err := bd.getNextBlocks(constant.BlockDownloadSegSize, peerUsed, chainBlockIds,
			start, commonUtil.MinUint32(start+segSize, stop))
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 53)
		if err != nil || len(nextBlocks) == 0 {
			// counting the error in a cycle
			numberOfErrorsInACycle++
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 54)
			if numberOfErrorsInACycle >= (len(peersSlice)/3)*2 {
				monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 55)
				return nil, blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf(
					"invalid blockchain downloaded from the feeder %v",
					peerUsed,
				))
			}
			continue
		}

		elapsedTime := time.Since(startTime)
		if elapsedTime > constant.MaxResponseTime {
			peersTobeDeactivated = append(peersTobeDeactivated, peerUsed)
		}
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 56)
		if len(nextBlocks) < 1 || uint32(len(nextBlocks)) > segSize {
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 57)
			bd.Logger.Warnf("disconnecting with peer %v for not responding correctly in getting the next blocks\n", peerUsed.Info.Address)
			peersTobeDeactivated = append(peersTobeDeactivated, peerUsed)
			continue
		}
		blocksSegments = append(blocksSegments, nextBlocks)
		start += uint32(len(nextBlocks))
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 58)
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 59)
	var blocksToBeProcessed []*model.Block
	for _, blockSegment := range blocksSegments {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 60)
		for i := 0; i < len(blockSegment); i++ {
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 61)
			if coreUtil.IsBlockIDExist(chainBlockIds, blockSegment[i].ID) {
				monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 62)
				blocksToBeProcessed = append(blocksToBeProcessed, blockSegment[i])
			}
		}
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 63)
	for _, peer := range peersTobeDeactivated {
		bd.PeerExplorer.DisconnectPeer(peer)
	}
	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 64)

	for idx, block := range blocksToBeProcessed {
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 65)
		if block.Height == 0 {
			continue
		}
		lastBlock, err := bd.BlockService.GetLastBlock()
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 66)
		if err != nil {
			return nil, err
		}
		if block.ID == lastBlock.ID && block.Height == lastBlock.Height {
			continue
		}
		previousBlockID := coreUtil.GetBlockIDFromHash(block.PreviousBlockHash)
		monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 67)
		if lastBlock.ID == previousBlockID {
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 68)
			err := bd.BlockService.ValidateBlock(block, lastBlock)
			if err != nil {
				monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 69)
				blockerUsed := blocker.ValidateMainBlockErr
				if chaintype.IsSpineChain(bd.ChainType) {
					blockerUsed = blocker.ValidateSpineBlockErr
				}
				bd.Logger.Warnf(
					"[download blockchain] failed to verify block %v from peer: %s\nwith previous: %v\nvalidateBlock fail: %v\n",
					block.ID, err.Error(), lastBlock.ID, blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()),
				)
				blacklistErr := bd.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
				if blacklistErr != nil {
					monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 70)
					bd.Logger.Errorf("Failed to add blacklist: %v\n", blacklistErr)
				}
				return &PeerForkInfo{
					FeederPeer: feederPeer,
				}, err
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 71)
			err = bd.BlockService.PushBlock(lastBlock, block, false, true)
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 72)
			if err != nil {
				blacklistErr := bd.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
				if blacklistErr != nil {
					bd.Logger.Errorf("Failed to add blacklist: %v\n", blacklistErr)
				}
				blockerUsed := blocker.PushMainBlockErr
				if chaintype.IsSpineChain(bd.ChainType) {
					blockerUsed = blocker.PushSpineBlockErr
				}
				bd.Logger.Warn(
					"[DownloadBlockchain] failed to push block from peer:",
					blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()),
				)
				return &PeerForkInfo{
					FeederPeer: feederPeer,
				}, err
			}
		} else {
			monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 73)
			forkBlocks = blocksToBeProcessed[idx:]
			break
		}
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(bd.ChainType, 74)
	return &PeerForkInfo{
		ForkBlocks: forkBlocks,
		FeederPeer: feederPeer,
	}, nil

}

func (bd *BlockchainDownloader) getPeerCommonBlockID(peer *model.Peer) (int64, error) {
	var (
		lastMilestoneBlockID int64
		trialCounter         uint32
		// to avoid processing duplicated block IDs
		commonMilestoneTemp = make(map[int64]bool)
	)
	lastBlock, err := bd.BlockService.GetLastBlock()
	if err != nil {
		bd.Logger.Infof("failed to get blockchain last block: %v\n", err)
		return 0, err
	}

	lastBlockID := lastBlock.ID
	for {
		if trialCounter >= constant.MaxCommonMilestoneRequestTrial {
			err := bd.PeerExplorer.PeerBlacklist(peer, "different blockchain fork")
			if err != nil {
				bd.Logger.Errorf("Failed to add blacklist: %v\n", err)
			}
			return 0, err
		}
		trialCounter++
		commonMilestoneBlockIDResponse, err := bd.PeerServiceClient.GetCommonMilestoneBlockIDs(
			peer, bd.ChainType, lastBlockID, lastMilestoneBlockID,
		)
		if err != nil {
			bd.Logger.Infof("failed to get common milestone from the peer: %v\n", err)
			bd.PeerExplorer.DisconnectPeer(peer)
			return 0, err
		}

		if commonMilestoneBlockIDResponse.Last {
			bd.PeerHasMore = false
		}

		for _, blockID := range commonMilestoneBlockIDResponse.BlockIds {
			if commonMilestoneTemp[blockID] {
				continue
			}
			_, err := bd.BlockService.GetBlockByID(blockID, false)
			if err == nil {
				return blockID, nil
			}
			errCasted := err.(blocker.Blocker)
			if errCasted.Type != blocker.BlockNotFoundErr {
				return 0, err
			}
			lastMilestoneBlockID = blockID
			commonMilestoneTemp[blockID] = true
		}

		// if block is not found and it's indicated as genesis
		if len(commonMilestoneTemp) == 1 {
			for blockID := range commonMilestoneTemp {
				if blockID == lastMilestoneBlockID {
					return 0, err
				}
			}
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
		_, err := bd.BlockService.GetBlockByID(blockID, false)
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
	if err != nil {
		return nil, err
	}
	nextBlocks := nextBlocksResponse.NextBlocks
	nextBlocksLength := uint32(len(nextBlocks))
	if nextBlocksLength > maxNextBlocks {
		return nil, fmt.Errorf("too many blocks returned (%d blocks), possibly a rogue peer %v", nextBlocksLength, peerUsed.Info.Address)
	}
	if len(nextBlocks) > 0 {
		return nextBlocks, nil
	}
	return blocks, nil
}

package service

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	util3 "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
	util2 "github.com/zoobc/zoobc-core/p2p/util"
)

type (
	// PublishedReceiptServiceInterface act as interface for processing the published receipt data
	PublishedReceiptServiceInterface interface {
		ProcessPublishedReceipts(previousBlock, block *model.Block) (int, error)
	}

	PublishedReceiptService struct {
		PublishedReceiptQuery        query.PublishedReceiptQueryInterface
		BlockQuery                   query.BlockQueryInterface
		ReceiptUtil                  util.ReceiptUtilInterface
		PublishedReceiptUtil         util.PublishedReceiptUtilInterface
		ReceiptService               ReceiptServiceInterface
		QueryExecutor                query.ExecutorInterface
		TransactionCoreService       TransactionCoreServiceInterface
		ScrambleNodeService          ScrambleNodeServiceInterface
		NodeRegistrationService      NodeRegistrationServiceInterface
		NodeConfigurationService     NodeConfigurationServiceInterface
		ProvedReceiptReminderStorage storage.CacheStackStorageInterface
		BlocksStorage                storage.CacheStackStorageInterface
	}
)

func NewPublishedReceiptService(
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	blockQuery query.BlockQueryInterface,
	receiptUtil util.ReceiptUtilInterface,
	publishedReceiptUtil util.PublishedReceiptUtilInterface,
	receiptService ReceiptServiceInterface,
	queryExecutor query.ExecutorInterface,
	transactionCoreService TransactionCoreServiceInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	nodeConfigurationService NodeConfigurationServiceInterface,
	provedReceiptReminderStorage storage.CacheStackStorageInterface,
	blockStorage storage.CacheStackStorageInterface,
) *PublishedReceiptService {
	return &PublishedReceiptService{
		PublishedReceiptQuery:        publishedReceiptQuery,
		BlockQuery:                   blockQuery,
		ReceiptUtil:                  receiptUtil,
		PublishedReceiptUtil:         publishedReceiptUtil,
		ReceiptService:               receiptService,
		QueryExecutor:                queryExecutor,
		TransactionCoreService:       transactionCoreService,
		ScrambleNodeService:          scrambleNodeService,
		NodeRegistrationService:      nodeRegistrationService,
		NodeConfigurationService:     nodeConfigurationService,
		ProvedReceiptReminderStorage: provedReceiptReminderStorage,
		BlocksStorage:                blockStorage,
	}
}

// ProcessPublishedReceipts takes published receipts in a block and validate
// them, this function will run in a db transaction to ensure
// queryExecutor.Begin() is called before calling this function.
func (ps *PublishedReceiptService) ProcessPublishedReceipts(previousBlock, block *model.Block) (int, error) {
	var (
		linkedCount int
		err         error
	)
	if block.GetHeight() < constant.MaxReceiptBatchCacheRound {
		return linkedCount, err
	}
	scrambleAtReceiptHeight, err := ps.ScrambleNodeService.GetScrambleNodesByHeight(block.Height - constant.MaxReceiptBatchCacheRound)
	if err != nil {
		return linkedCount, err
	}
	blocksmithNodeRegistry, err := ps.NodeRegistrationService.GetNodeRegistrationByNodePublicKey(block.GetBlocksmithPublicKey())
	if err != nil {
		return linkedCount, err
	}
	blocksmithPriority, blocksmithSortedPriority, err := util2.GetPriorityPeersByNodeID(
		blocksmithNodeRegistry.GetNodeID(),
		scrambleAtReceiptHeight,
	)
	if err != nil {
		return linkedCount, err
	}
	hostID, err := ps.NodeConfigurationService.GetHostID()
	if err != nil {
		fmt.Printf("non-critical-error: %v", err)
	}
	hostPublicKey := ps.NodeConfigurationService.GetNodePublicKey()
	for index, rc := range block.GetFreeReceipts() {
		// validate sender and recipient of receipt
		rcCopy := *rc
		err = ps.ReceiptService.ValidateReceipt(rc.GetReceipt())
		if err != nil {
			return 0, err
		}

		// check if block.Blocksmith has me as priority peer
		if _, ok := blocksmithPriority[fmt.Sprintf("%d", hostID)]; ok {
			var provedReceiptReminder storage.ProvedReceiptReminderObject
			if rc.GetReceipt().GetRMRLinked() != nil &&
				bytes.Equal(rc.GetReceipt().GetRecipientPublicKey(), hostPublicKey) {
				// insert empty bytes as merkle tree to indicate that node was in priority but not having its receipt published
				provedReceiptReminder = storage.ProvedReceiptReminderObject{
					MerkleRoot:           rc.GetReceipt().GetRMRLinked(),
					ReferenceBlockHash:   rc.GetReceipt().GetReferenceBlockHash(),
					ReferenceBlockHeight: rc.GetReceipt().GetReferenceBlockHeight(),
				}
			} else {
				// insert empty bytes as merkle tree to indicate that node was in priority but not having its receipt published
				provedReceiptReminder = storage.ProvedReceiptReminderObject{
					MerkleRoot: make([]byte, 0),
				}
			}
			err := ps.ProvedReceiptReminderStorage.Push(provedReceiptReminder)
			if err != nil {
				return linkedCount, err
			}
		}
		// store in database
		// assign index and height, index is the order of the receipt in the block,
		// it's different with receiptIndex which is used to validate merkle root.
		rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)
		err := ps.PublishedReceiptUtil.InsertPublishedReceipt(&rcCopy, true)
		if err != nil {
			return 0, err
		}
	}
	rng := crypto.NewRandomNumberGenerator()
	rng.Reset(constant.BlocksmithSelectionProvedReceiptSeedPrefix, block.GetBlockSeed())
	for index, rc := range block.GetProvedReceipts() {
		// generate random number (consensus safe) as to which receipt to pick
		rdNumItemIndex := rng.Next()
		leafRandomNumber := rng.Next()
		rcCopy := *rc
		if ps.ReceiptService.IsProvedReceiptEmpty(rc) {
			// node doesn't publish receipt for this slot, skipping
			fmt.Printf("empty proved at index: %d\n", index)
			continue
		}
		fmt.Printf("NON-empty proved at index: %d\n", index)

		// validation...
		// fetch block+txs at provedReceiptRO height
		blockAtHeight, err := util3.GetBlockByHeightUseBlocksCache(
			rc.GetBatchReferenceBlockHeight()-1,
			ps.QueryExecutor,
			ps.BlockQuery,
			ps.BlocksStorage,
		)
		if err != nil {
			return linkedCount, err
		}
		txsAtHeight, err := ps.TransactionCoreService.GetTransactionsByBlockHeight(blockAtHeight.Height)
		if err != nil {
			return linkedCount, err
		}
		itemIndex := rng.ConvertRandomNumberToIndex(rdNumItemIndex, int64(len(txsAtHeight)+1))
		// pick the right data hash of the receipt based on `itemIndex` value
		var (
			itemHash []byte
		)
		if itemIndex == 0 {
			itemHash = blockAtHeight.BlockHash
		} else {
			itemHash = txsAtHeight[itemIndex-1].TransactionHash
		}
		if !bytes.Equal(rc.GetReceipt().GetDatumHash(), itemHash) {
			// node does not publish the expected receipt, stop receipt validation, block has invalid proved receipt
			return 0, blocker.NewBlocker(blocker.ValidationErr, "ProcessPublishReceipt:InvalidReceiptHashPublished")
		}
		scrambleAtHeight, err := ps.ScrambleNodeService.GetScrambleNodesByHeight(rc.GetBatchReferenceBlockHeight())
		if err != nil {
			return 0, blocker.NewBlocker(
				blocker.AppErr,
				fmt.Sprintf("ProcessPublishReceipt:GetScrambleNodesByHeight-%v", err),
			)
		}
		_, sortedPriorityAtHeight, err := util2.GetPriorityPeersByNodeID(blocksmithNodeRegistry.GetNodeID(), scrambleAtHeight)
		if err != nil {
			fmt.Printf("%v", err)
			return 0, blocker.NewBlocker(
				blocker.AppErr,
				fmt.Sprintf("ProcessPublishReceipt:GetPriorityPeersByNodeID-%v", err),
			)
		}
		recipientIndex := rng.ConvertRandomNumberToIndex(leafRandomNumber, int64(len(sortedPriorityAtHeight)))
		if int(recipientIndex) >= len(blocksmithSortedPriority) {
			return 0, blocker.NewBlocker(
				blocker.ValidationErr,
				fmt.Sprintf("ProcessPublishReceipt:InvalidReceiptRecipient-IndexOutOfRange-index=%d-priorityPeerLength=%d",
					recipientIndex,
					len(blocksmithSortedPriority),
				),
			)
		}
		// check if receipt come from expected recipient based on `recipientIndex`
		if !bytes.Equal(
			rc.GetReceipt().GetRecipientPublicKey(),
			blocksmithSortedPriority[recipientIndex].GetInfo().GetPublicKey(),
		) {
			// looking for which receipt included, -1 means no matching recipient
			var getIndex = -1
			for i, peer := range blocksmithSortedPriority {
				if bytes.Equal(peer.GetInfo().GetPublicKey(), rc.GetReceipt().GetRecipientPublicKey()) {
					getIndex = i
					break
				}
			}
			return 0, blocker.NewBlocker(
				blocker.ValidationErr,
				fmt.Sprintf("ProcessPublishReceipt:InvalidReceiptRecipient-expect:index=%d-pk=%s-get:index=%d:pk=%s",
					recipientIndex, hex.EncodeToString(blocksmithSortedPriority[recipientIndex].GetInfo().GetPublicKey()),
					getIndex, hex.EncodeToString(rc.GetReceipt().GetRecipientPublicKey()),
				),
			)
		}
		// calculate rc+intermediateHash merkle root, and validate if is in `published_receipt.rmr_linked`
		calculatedMerkleRoot, err := ps.ReceiptService.GetMerkleRootFromReceiptIntermediateHash(rc.GetReceipt(), rc.GetIntermediateHashes())
		// check if calculated merkle root in `published_receipt.rmr_linked`
		_, err = ps.PublishedReceiptUtil.GetPublishedReceiptByLinkedRMR(calculatedMerkleRoot)
		if err != nil {
			return 0, err
		}
		// store in database
		// assign index and height, index is the order of the receipt in the block,
		// it's different with receiptIndex which is used to validate merkle root.
		rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)
		err = ps.PublishedReceiptUtil.InsertPublishedReceipt(&rcCopy, true)
		if err != nil {
			return 0, err
		}
	}
	return linkedCount, nil
}

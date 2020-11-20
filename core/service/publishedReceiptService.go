package service

import (
	"bytes"
	"fmt"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/core/util"
	util2 "github.com/zoobc/zoobc-core/p2p/util"
)

type (
	// PublishedReceiptServiceInterface act as interface for processing the published receipt data
	PublishedReceiptServiceInterface interface {
		ProcessPublishedReceipts(block *model.Block) (int, error)
	}

	PublishedReceiptService struct {
		PublishedReceiptQuery        query.PublishedReceiptQueryInterface
		ReceiptUtil                  util.ReceiptUtilInterface
		PublishedReceiptUtil         util.PublishedReceiptUtilInterface
		ReceiptService               ReceiptServiceInterface
		QueryExecutor                query.ExecutorInterface
		ScrambleNodeService          ScrambleNodeServiceInterface
		NodeRegistrationService      NodeRegistrationServiceInterface
		NodeConfigurationService     NodeConfigurationServiceInterface
		ProvedReceiptReminderStorage storage.CacheStorageInterface
	}
)

func NewPublishedReceiptService(
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil util.ReceiptUtilInterface,
	publishedReceiptUtil util.PublishedReceiptUtilInterface,
	receiptService ReceiptServiceInterface,
	queryExecutor query.ExecutorInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	nodeConfigurationService NodeConfigurationServiceInterface,
	provedReceiptReminderStorage storage.CacheStorageInterface,
) *PublishedReceiptService {
	return &PublishedReceiptService{
		PublishedReceiptQuery:        publishedReceiptQuery,
		ReceiptUtil:                  receiptUtil,
		PublishedReceiptUtil:         publishedReceiptUtil,
		ReceiptService:               receiptService,
		QueryExecutor:                queryExecutor,
		ScrambleNodeService:          scrambleNodeService,
		NodeRegistrationService:      nodeRegistrationService,
		NodeConfigurationService:     nodeConfigurationService,
		ProvedReceiptReminderStorage: provedReceiptReminderStorage,
	}
}

// ProcessPublishedReceipts takes published receipts in a block and validate
// them, this function will run in a db transaction so ensure
// queryExecutor.Begin() is called before calling this function.
func (ps *PublishedReceiptService) ProcessPublishedReceipts(block *model.Block) (int, error) {
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
	blocksmithPriority, err := util2.GetPriorityPeersByNodeID(blocksmithNodeRegistry.GetNodeID(), scrambleAtReceiptHeight)
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
			if bytes.Equal(rc.GetReceipt().GetRecipientPublicKey(), hostPublicKey) {
				// insert empty bytes as merkle tree to indicate that node was in priority but not having its receipt published
				provedReceiptReminder = storage.ProvedReceiptReminderObject{
					MerkleRoot: rc.GetReceipt().RMRLinked,
				}
			} else {
				// insert empty bytes as merkle tree to indicate that node was in priority but not having its receipt published
				provedReceiptReminder = storage.ProvedReceiptReminderObject{
					MerkleRoot: make([]byte, 0),
				}
			}
			err := ps.ProvedReceiptReminderStorage.SetItem(block.Height, provedReceiptReminder)
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

		rcCopy := *rc
		if ps.ReceiptService.IsProvedReceiptEmpty(rc) {
			continue
		}
		// validation...
		// todo
		// store in database
		// assign index and height, index is the order of the receipt in the block,
		// it's different with receiptIndex which is used to validate merkle root.
		rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)
		err := ps.PublishedReceiptUtil.InsertPublishedReceipt(&rcCopy, true)
		if err != nil {
			return 0, err
		}
	}
	return linkedCount, nil
}

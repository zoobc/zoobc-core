package blockchainsync

import (
	"bytes"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

type (
	ForkingProcessorInterface interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error
	}
	ForkingProcessor struct {
		ChainType             chaintype.ChainType
		BlockService          service.BlockServiceInterface
		QueryExecutor         query.ExecutorInterface
		ActionTypeSwitcher    transaction.TypeActionSwitcher
		MempoolService        service.MempoolServiceInterface
		KVExecutor            kvdb.KVExecutorInterface
		Logger                *log.Logger
		PeerExplorer          strategy.PeerExplorerStrategyInterface
		TransactionUtil       transaction.UtilInterface
		TransactionCorService service.TransactionCoreServiceInterface
	}
)

// ProcessFork processes the forked blocks
func (fp *ForkingProcessor) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error {
	var (
		lastBlockBeforeProcess, lastBlock, currentLastBlock *model.Block
		myPoppedOffBlocks, peerPoppedOffBlocks              []*model.Block
		pushedForkBlocks                                    int
		err                                                 error
	)
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 80)

	lastBlockBeforeProcess, err = fp.BlockService.GetLastBlock()
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 81)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 82)
		return err
	}
	beforeApplyCumulativeDifficulty := lastBlockBeforeProcess.CumulativeDifficulty
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 83)
	myPoppedOffBlocks, err = fp.BlockService.PopOffToBlock(commonBlock)
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 84)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 85)
		return err
	}

	lastBlock, err = fp.BlockService.GetLastBlock()
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 86)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 87)
		return err
	}

	if lastBlock.ID == commonBlock.ID {
		// rebuilding the chain
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 88)
		for _, block := range forkBlocks {
			if block.ID == lastBlock.ID {
				continue
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 89)
			lastBlock, err = fp.BlockService.GetLastBlock()
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 90)
			if err != nil {
				return err
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 91)
			lastBlockHash, err := commonUtil.GetBlockHash(lastBlock, fp.ChainType)
			if err != nil {
				return err
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 92)
			if bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 93)
				err := fp.BlockService.ValidateBlock(block, lastBlock)
				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 94)
				if err != nil {
					monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 95)
					blacklistErr := fp.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
					if blacklistErr != nil {
						fp.Logger.Errorf("Failed to add blacklist: %v\n", blacklistErr)
					}
					blockerUsed := blocker.ValidateMainBlockErr
					if chaintype.IsSpineChain(fp.ChainType) {
						blockerUsed = blocker.ValidateSpineBlockErr
					}
					fp.Logger.Warnf("[ProcessFork] failed to verify block %v from peer %v: %s\nwith previous: %v\ndownloadBlockchain validateBlock fail: %v\n",
						block.ID, p2pUtil.GetFullAddressPeer(feederPeer), err, lastBlock.ID, blocker.NewBlocker(blockerUsed, err.Error(), block, lastBlock))
					break
				}
				err = fp.BlockService.PushBlock(lastBlock, block, false, true)
				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 96)
				if err != nil {
					blacklistErr := fp.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
					if blacklistErr != nil {
						fp.Logger.Errorf("Failed to add blacklist: %v\n", blacklistErr)
					}
					blockerUsed := blocker.PushMainBlockErr
					if chaintype.IsSpineChain(fp.ChainType) {
						blockerUsed = blocker.PushSpineBlockErr
					}
					fp.Logger.Warnf("\n\n[ProcessFork] PushBlock of fork blocks err %v\n\n", blocker.NewBlocker(blockerUsed, err.Error(), block, lastBlock))
					break
				}

				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 97)
				pushedForkBlocks++
			}
		}
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 98)
	currentLastBlock, err = fp.BlockService.GetLastBlock()
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 99)
	if err != nil {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 100)
		return err
	}
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 101)
	currentCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
	cumulativeDifficultyOriginalBefore, _ := new(big.Int).SetString(beforeApplyCumulativeDifficulty, 10)

	// if after applying the fork blocks the cumulative difficulty is still less than current one
	// only take the transactions to be processed, but later will get back to our own fork
	if pushedForkBlocks > 0 && currentCumulativeDifficulty.Cmp(cumulativeDifficultyOriginalBefore) < 0 {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 102)
		peerPoppedOffBlocks, err = fp.BlockService.PopOffToBlock(commonBlock)
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 103)
		if err != nil {
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 104)
			return err
		}
		pushedForkBlocks = 0
		for _, block := range peerPoppedOffBlocks {
			_ = fp.ProcessLater(block.Transactions)
		}
	}

	// if no fork blocks successfully applied, go back to our fork
	// other wise, just take the transactions of our popped blocks to be processed later
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 105)
	if pushedForkBlocks == 0 {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 106)
		fp.Logger.Println("Did not accept any blocks from peer, pushing back my blocks")
		for _, block := range myPoppedOffBlocks {
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 107)
			lastBlock, err = fp.BlockService.GetLastBlock()
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 108)
			if err != nil {
				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 109)
				return err
			}
			err = fp.BlockService.ValidateBlock(block, lastBlock)
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 110)
			if err != nil {
				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 111)
				blacklistErr := fp.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
				if blacklistErr != nil {
					fp.Logger.Errorf("Failed to add blacklist: %v\n", blacklistErr)
				}
				blockerUsed := blocker.ValidateMainBlockErr
				if chaintype.IsSpineChain(fp.ChainType) {
					blockerUsed = blocker.ValidateSpineBlockErr
				}
				fp.Logger.Warnf("[pushing back own block] failed to verify block %v from peer: %s\n with previous: %v\nvalidateBlock fail: %v\n",
					block.ID, err.Error(), lastBlock.ID, blocker.NewBlocker(blockerUsed, err.Error(), block, lastBlock))
				return err
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 112)
			err = fp.BlockService.PushBlock(lastBlock, block, false, true)
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 113)
			if err != nil {
				monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 114)
				blockerUsed := blocker.PushMainBlockErr
				if chaintype.IsSpineChain(fp.ChainType) {
					blockerUsed = blocker.PushSpineBlockErr
				}
				fp.Logger.Warnf("\n\nPushBlock of fork blocks err %v\n\n", blocker.NewBlocker(blockerUsed, err.Error(), block, lastBlock))
				return blocker.NewBlocker(blocker.BlockErr, "Popped off block no longer acceptable")
			}
		}
	} else {
		monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 115)
		for _, block := range myPoppedOffBlocks {
			monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 116)
			_ = fp.ProcessLater(block.Transactions)
		}
	}

	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 117)
	if fp.ChainType.HasTransactions() {
		// start restoring mempool from badgerDB
		err = fp.restoreMempoolsBackup()
		if err != nil {
			fp.Logger.Errorf("RestoreBackupFail: %s", err.Error())
		}
	}
	monitoring.IncrementMainchainDownloadCycleDebugger(fp.ChainType, 118)
	return nil
}

func (fp *ForkingProcessor) ProcessLater(txs []*model.Transaction) error {
	var (
		err     error
		txBytes []byte
		txType  transaction.TypeAction
	)
	for _, tx := range txs {
		// Validate Tx
		txType, err = fp.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return err
		}
		txBytes, err = fp.TransactionUtil.GetTransactionBytes(tx, true)

		if err != nil {
			return err
		}

		// Save to mempool
		mpTx := &model.MempoolTransaction{
			FeePerByte:              commonUtil.FeePerByteTransaction(tx.GetFee(), txBytes),
			ID:                      tx.ID,
			TransactionBytes:        txBytes,
			ArrivalTimestamp:        time.Now().Unix(),
			SenderAccountAddress:    tx.SenderAccountAddress,
			RecipientAccountAddress: tx.RecipientAccountAddress,
		}

		err = fp.MempoolService.ValidateMempoolTransaction(tx)
		if err != nil {
			return err
		}
		// Apply Unconfirmed
		err = fp.QueryExecutor.BeginTx()
		if err != nil {
			return err
		}
		err = fp.TransactionCorService.ApplyUnconfirmedTransaction(txType)
		if err != nil {
			errRollback := fp.QueryExecutor.RollbackTx()
			if errRollback != nil {
				return errRollback
			}
			return err
		}
		err = fp.MempoolService.AddMempoolTransaction(mpTx)
		if err != nil {
			errRollback := fp.QueryExecutor.RollbackTx()
			if errRollback != nil {
				return err
			}
			return err
		}
		err = fp.QueryExecutor.CommitTx()
		if err != nil {
			return err
		}
	}
	return nil
}

func (fp *ForkingProcessor) ScheduleScan(height uint32, validate bool) {
	// TODO: analyze if this mechanism is necessary
}

// restoreMempoolsBackup will restore transaction from badgerDB and try to re-ApplyUnconfirmed
func (fp *ForkingProcessor) restoreMempoolsBackup() error {

	var (
		mempoolsBackupBytes []byte
		prev                uint32
		err                 error
	)

	kvdbMempoolsBackupKey := commonUtil.GetKvDbMempoolDBKey(fp.ChainType)
	mempoolsBackupBytes, err = fp.KVExecutor.Get(kvdbMempoolsBackupKey)
	if err != nil {
		return err
	}

	for int(prev) < len(mempoolsBackupBytes) {
		var (
			transactionBytes []byte
			mempoolTX        *model.MempoolTransaction
			txType           transaction.TypeAction
			tx               *model.Transaction
			size             uint32
		)

		prev += constant.TransactionBodyLength // initiate length of size
		size = commonUtil.ConvertBytesToUint32(mempoolsBackupBytes[:prev])
		transactionBytes = mempoolsBackupBytes[prev:][:size]
		prev += size

		tx, err = fp.TransactionUtil.ParseTransactionBytes(transactionBytes, true)
		if err != nil {
			return err
		}
		mempoolTX = &model.MempoolTransaction{
			FeePerByte:              commonUtil.FeePerByteTransaction(tx.GetFee(), transactionBytes),
			ID:                      tx.ID,
			TransactionBytes:        transactionBytes,
			ArrivalTimestamp:        time.Now().Unix(),
			SenderAccountAddress:    tx.SenderAccountAddress,
			RecipientAccountAddress: tx.RecipientAccountAddress,
		}
		err = fp.MempoolService.ValidateMempoolTransaction(tx)
		if err != nil {
			// no need to break the process in this case
			fp.Logger.Warnf("Invalid mempool want to restore with ID: %d", tx.GetID())
		}

		txType, err = fp.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return err
		}
		// Apply Unconfirmed
		err = fp.QueryExecutor.BeginTx()
		if err != nil {
			return err
		}
		err = fp.TransactionCorService.ApplyUnconfirmedTransaction(txType)
		if err != nil {
			rollbackErr := fp.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				fp.Logger.Warnf("error when executing database rollback: %v", rollbackErr)
			}
			return err
		}
		err = fp.MempoolService.AddMempoolTransaction(mempoolTX)
		if err != nil {
			rollbackErr := fp.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				fp.Logger.Warnf("error when executing database rollback: %v", rollbackErr)
			}
			return err
		}
		err = fp.QueryExecutor.CommitTx()
		if err != nil {
			return err
		}
		// remove restored mempools from badger
		err = fp.KVExecutor.Delete(commonUtil.GetKvDbMempoolDBKey(fp.ChainType))
		if err != nil {
			return err
		}
	}
	return nil
}

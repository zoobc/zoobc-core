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
package blockchainsync

import (
	"bytes"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
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
		Logger                *log.Logger
		PeerExplorer          strategy.PeerExplorerStrategyInterface
		TransactionUtil       transaction.UtilInterface
		TransactionCorService service.TransactionCoreServiceInterface
		MempoolBackupStorage  storage.CacheStorageInterface
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
						block.ID, p2pUtil.GetFullAddressPeer(feederPeer), err, lastBlock.ID,
						blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()),
					)
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
					fp.Logger.Warnf(
						"\n\n[ProcessFork] PushBlock of fork blocks err %v\n\n",
						blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()),
					)
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
					block.ID, err.Error(), lastBlock.ID, blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()))
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
				fp.Logger.Warnf(
					"\n\nPushBlock of fork blocks err %v\n\n",
					blocker.NewBlocker(blockerUsed, err.Error(), block.GetID(), lastBlock.GetID()),
				)
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
		// start restoring mempool
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
		err              error
		txBytes          []byte
		txType           transaction.TypeAction
		highPriorityLock = true
	)
	err = fp.QueryExecutor.BeginTx(highPriorityLock, monitoring.ProcessMempoolLaterOwnerProcess)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		// Validate Tx
		txType, err = fp.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			fp.Logger.Warnf("ProcessLater:GetTransactionType - tx.Height: %d - txID: %d - %s", tx.GetHeight(), tx.GetID(), err.Error())
			continue
		}
		txBytes, err = fp.TransactionUtil.GetTransactionBytes(tx, true)

		if err != nil {
			fp.Logger.Warnf("ProcessLater:GetTransactionBytes - tx.Height: %d - txID: %d - %s", tx.GetHeight(), tx.GetID(), err.Error())
			continue
		}

		err = fp.MempoolService.ValidateMempoolTransaction(tx)
		if err != nil {
			fp.Logger.Warnf("ProcessLater:ValidateMempoolTransaction - tx.Height: %d - txID: %d - %s", tx.GetHeight(), tx.GetID(), err.Error())
			continue
		}

		err = fp.TransactionCorService.ApplyUnconfirmedTransaction(txType)
		if err != nil {
			fp.Logger.Warnf("ProcessLater:ApplyUnconfirmedTransaction - tx.Height: %d - txID: %d - %s", tx.GetHeight(), tx.GetID(), err.Error())
			continue
		}
		err = fp.MempoolService.AddMempoolTransaction(tx, txBytes)
		if err != nil {
			// undo spendable balance when add mempool fail
			errUndo := fp.TransactionCorService.UndoApplyUnconfirmedTransaction(txType)
			if errUndo != nil {
				fp.Logger.Warnf("ProcessLater:UndoApplyUnconfirmedTransaction - tx.Height: %d - txID: %d - %s", tx.GetHeight(), tx.GetID(), errUndo.Error())
				// rollback DB when fail undo spendable balance
				return fp.QueryExecutor.RollbackTx(highPriorityLock)
			}
			fp.Logger.Warnf("ProcessLater:AddMempoolFail - tx.Height: %d - txID: %d - %s", tx.GetHeight(), tx.GetID(), err.Error())
			continue
		}
	}
	return fp.QueryExecutor.CommitTx(highPriorityLock)
}

func (fp *ForkingProcessor) ScheduleScan(height uint32, validate bool) {
	// TODO: analyze if this mechanism is necessary
}

// restoreMempoolsBackup will restore transactions and try to re-ApplyUnconfirmed
func (fp *ForkingProcessor) restoreMempoolsBackup() error {

	var (
		err              error
		mempools         map[int64][]byte
		highPriorityLock = true
	)

	err = fp.MempoolBackupStorage.GetAllItems(&mempools)
	if err != nil {
		return err
	}
	// Apply Unconfirmed
	err = fp.QueryExecutor.BeginTx(highPriorityLock, monitoring.RestoreMempoolsBackupOwnerProcess)
	if err != nil {
		return err
	}
	for id := range mempools {
		func(mempoolID int64, errUndo *error) {
			*errUndo = nil
			var (
				tx     *model.Transaction
				txType transaction.TypeAction
			)

			defer func() {
				if removeErr := fp.MempoolBackupStorage.RemoveItem(mempoolID); removeErr != nil {
					fp.Logger.Warnf("restoreMemmpoolBackup - mempool ID: %d; %s", tx.GetID(), removeErr.Error())
				}
			}()
			tx, err = fp.TransactionUtil.ParseTransactionBytes(mempools[mempoolID], true)
			if err != nil {
				fp.Logger.Warnf(err.Error())
				return
			}

			err = fp.MempoolService.ValidateMempoolTransaction(tx)
			if err != nil {
				// no need to break the process in this case
				fp.Logger.Warnf(err.Error())
				return
			}

			txType, err = fp.ActionTypeSwitcher.GetTransactionType(tx)
			if err != nil {
				fp.Logger.Warnf(err.Error())
				return
			}

			err = fp.TransactionCorService.ApplyUnconfirmedTransaction(txType)
			if err != nil {
				fp.Logger.Warnf("restoreMempoolsBackup:ApplyUnconfirmedTransaction: %v", err)
				return
			}
			err = fp.MempoolService.AddMempoolTransaction(tx, mempools[mempoolID])
			if err != nil {
				// undo spendable balance when add mempool fail
				*errUndo = fp.TransactionCorService.UndoApplyUnconfirmedTransaction(txType)
				if *errUndo != nil {
					fp.Logger.Warnf("restoreMempoolsBackup:UndoApplyUnconfirmedTransaction %v", err)
					return
				}
				fp.Logger.Warnf("error when AddMempoolTransaction: %v", err)
				return
			}
		}(id, &err)
		// rollback DB when undo spendable balance fail
		if err != nil {
			return fp.QueryExecutor.RollbackTx(highPriorityLock)
		}

	}
	err = fp.QueryExecutor.CommitTx(highPriorityLock)
	return err
}

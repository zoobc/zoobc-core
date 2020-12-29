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
package monitoring

import (
	"database/sql"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"math"
	"net/http"
)

var (
	isMonitoringActive bool
	nodePublicKey      []byte

	sendAddressInfoToPeer              prometheus.Counter
	getAddressInfoTableFromPeer        prometheus.Counter
	receiptCounter                     prometheus.Counter
	nodeAddressInfoCounter             prometheus.Gauge
	confirmedAddressCounter            prometheus.Gauge
	pendingAddressCounter              prometheus.Gauge
	unresolvedPeersCounter             prometheus.Gauge
	resolvedPeersCounter               prometheus.Gauge
	unresolvedPriorityPeersCounter     prometheus.Gauge
	resolvedPriorityPeersCounter       prometheus.Gauge
	activeRegisteredNodesGauge         prometheus.Gauge
	nodeScore                          prometheus.Gauge
	tpsReceived                        prometheus.Gauge
	tpsProcessed                       prometheus.Gauge
	txReceived                         prometheus.Gauge
	txProcessed                        prometheus.Gauge
	txFiltered                         prometheus.Gauge
	P2PTxFilteredIncoming              prometheus.Gauge
	P2PTxFilteredOutgoing              prometheus.Gauge
	blockerCounterVector               *prometheus.CounterVec
	statusLockGaugeVector              *prometheus.GaugeVec
	blockchainStatusGaugeVector        *prometheus.GaugeVec
	blockchainSmithIndexGaugeVector    *prometheus.GaugeVec
	blockchainIDMsbGaugeVector         *prometheus.GaugeVec
	blockchainIDLsbGaugeVector         *prometheus.GaugeVec
	blockchainHeightGaugeVector        *prometheus.GaugeVec
	goRoutineActivityGaugeVector       *prometheus.GaugeVec
	downloadCycleDebuggerGaugeVector   *prometheus.GaugeVec
	apiGaugeVector                     *prometheus.GaugeVec
	apiRunningGaugeVector              *prometheus.GaugeVec
	snapshotDownloadRequestCounter     *prometheus.CounterVec
	dbStatGaugeVector                  *prometheus.GaugeVec
	cacheStorageGaugeVector            *prometheus.GaugeVec
	mempoolTransactionCountGaugeVector *prometheus.GaugeVec
	blockProcessTimeGaugeVector        *prometheus.GaugeVec

	cliMonitoringInstance CLIMonitoringInteface
)

const (
	P2pGetPeerInfoServer                = "P2pGetPeerInfoServer"
	P2pGetMorePeersServer               = "P2pGetMorePeersServer"
	P2pSendPeersServer                  = "P2pSendPeersServer"
	P2pSendBlockServer                  = "P2pSendBlockServer"
	P2pSendTransactionServer            = "P2pSendTransactionServer"
	P2pRequestBlockTransactionsServer   = "P2pRequestBlockTransactionsServer"
	P2pGetCumulativeDifficultyServer    = "P2pGetCumulativeDifficultyServer"
	P2pGetCommonMilestoneBlockIDsServer = "P2pGetCommonMilestoneBlockIDsServer"
	P2pGetNextBlockIDsServer            = "P2pGetNextBlockIDsServer"
	P2pGetNextBlocksServer              = "P2pGetNextBlocksServer"
	P2pRequestFileDownloadServer        = "P2pRequestFileDownloadServer"
	P2pGetNodeProofOfOriginServer       = "P2pGetNodeProofOfOriginServer"

	P2pGetPeerInfoClient                 = "P2pGetPeerInfoClient"
	P2pGetMorePeersClient                = "P2pGetMorePeersClient"
	P2pSendPeersClient                   = "P2pSendPeersClient"
	P2pSendNodeAddressInfoClient         = "P2pSendNodeAddressInfoClient"
	P2pGetNodeProofOfOwnershipInfoClient = "P2pGetNodeProofOfOwnershipInfoClient"
	P2pSendBlockClient                   = "P2pSendBlockClient"
	P2pSendTransactionClient             = "P2pSendTransactionClient"
	P2pRequestBlockTransactionsClient    = "P2pRequestBlockTransactionsClient"
	P2pGetCumulativeDifficultyClient     = "P2pGetCumulativeDifficultyClient"
	P2pGetCommonMilestoneBlockIDsClient  = "P2pGetCommonMilestoneBlockIDsClient"
	P2pGetNextBlockIDsClient             = "P2pGetNextBlockIDsClient"
	P2pGetNextBlocksClient               = "P2pGetNextBlocksClient"
	P2pRequestFileDownloadClient         = "P2pRequestFileDownloadClient"
)

var (
	// todo: andy-shi88 reporting data, tidy this up to let cliMonitor, prometheus, and status to fetch from single source
	lastMainBlock, lastSpineBlock            model.Block
	resolvedPeersCount, unresolvedPeersCount uint32
	blocksmithIndex                          int32
)

func Handler() http.Handler {
	return promhttp.Handler()
}

func GetNodeStatus() model.GetNodeStatusResponse {
	blockMainHashString, err := address.EncodeZbcID(constant.PrefixZoobcMainBlockHash, lastMainBlock.BlockHash)
	if err != nil {
		blockMainHashString = "-"
	}
	blockSpineHashString, err := address.EncodeZbcID(constant.PrefixZoobcSpineBlockHash, lastSpineBlock.BlockHash)
	if err != nil {
		blockSpineHashString = "-"
	}
	nodePublicKeyString, err := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, nodePublicKey)
	if err != nil {
		nodePublicKeyString = "-"
	}
	return model.GetNodeStatusResponse{
		LastMainBlockHeight:  lastMainBlock.Height,
		LastMainBlockHash:    blockMainHashString,
		LastSpineBlockHeight: lastSpineBlock.Height,
		LastSpineBlockHash:   blockSpineHashString,
		Version:              fmt.Sprintf("%s %s", constant.ApplicationCodeName, constant.ApplicationVersion),
		NodePublicKey:        nodePublicKeyString,
		UnresolvedPeers:      unresolvedPeersCount,
		ResolvedPeers:        resolvedPeersCount,
		BlocksmithIndex:      blocksmithIndex,
	}
}

func SetMonitoringActive(isActive bool) {
	isMonitoringActive = isActive

	sendAddressInfoToPeer = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "zoobc_send_address_info_request",
		Help: "send address info req",
	})
	prometheus.MustRegister(sendAddressInfoToPeer)

	getAddressInfoTableFromPeer = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "zoobc_get_address_info_table_request",
		Help: "get address info table req",
	})
	prometheus.MustRegister(getAddressInfoTableFromPeer)

	receiptCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "zoobc_receipts",
		Help: "receipts counter",
	})
	prometheus.MustRegister(receiptCounter)

	nodeAddressInfoCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_node_address_info_count",
		Help: "nodeAddressInfo counter",
	})
	prometheus.MustRegister(nodeAddressInfoCounter)

	confirmedAddressCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_node_address_info_count_status_confirmed",
		Help: "confirmed addresses by node counter",
	})
	prometheus.MustRegister(confirmedAddressCounter)

	pendingAddressCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_node_address_info_count_status_pending",
		Help: "pending addresses by node counter",
	})
	prometheus.MustRegister(pendingAddressCounter)

	unresolvedPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_unresolved_peers",
		Help: "unresolvedPeers counter",
	})
	prometheus.MustRegister(unresolvedPeersCounter)

	resolvedPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_resolved_peers",
		Help: "resolvedPeers counter",
	})
	prometheus.MustRegister(resolvedPeersCounter)

	resolvedPriorityPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_resolved_priority_peers",
		Help: "priority resolvedPeers counter",
	})
	prometheus.MustRegister(resolvedPriorityPeersCounter)

	unresolvedPriorityPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_unresolved_priority_peers",
		Help: "priority resolvedPeers counter",
	})
	prometheus.MustRegister(unresolvedPriorityPeersCounter)

	activeRegisteredNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_active_registered_nodes",
		Help: "active registered nodes counter",
	})
	prometheus.MustRegister(activeRegisteredNodesGauge)

	blockerCounterVector = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "zoobc_err",
		Help: "Error blocker error counter",
	}, []string{"blocker_type"})
	prometheus.MustRegister(blockerCounterVector)

	statusLockGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_status_lock",
		Help: "Status lock counter",
	}, []string{"chaintype", "status_type"})
	prometheus.MustRegister(statusLockGaugeVector)

	blockchainStatusGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_blockchain_status",
		Help: "Blockchain status",
	}, []string{"chaintype"})
	prometheus.MustRegister(blockchainStatusGaugeVector)

	blockchainSmithIndexGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_blockchain_smith_time",
		Help: "Smith time of each nodes to smith for each chain",
	}, []string{"chaintype"})
	prometheus.MustRegister(blockchainSmithIndexGaugeVector)

	nodeScore = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_node_score",
		Help: "The score of the node (divided by 100 to fit the max float64)",
	})
	prometheus.MustRegister(nodeScore)

	tpsReceived = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_tps_received",
		Help: "Transactions per second received",
	})
	prometheus.MustRegister(tpsReceived)

	txReceived = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_tx_received",
		Help: "Transactions received since node last start",
	})
	prometheus.MustRegister(txReceived)

	tpsProcessed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_tps_processed",
		Help: "Transactions per second processed",
	})
	prometheus.MustRegister(tpsProcessed)

	txProcessed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_tx_processed",
		Help: "Transactions processed since node last start",
	})
	prometheus.MustRegister(txProcessed)

	txFiltered = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_tx_filtered",
		Help: "Transactions filtered by anti-spam strategy",
	})
	prometheus.MustRegister(txFiltered)

	P2PTxFilteredIncoming = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_p2p_tx_filtered_incoming",
		Help: "Transactions broadcast by other nodes filtered by anti-spam strategy",
	})
	prometheus.MustRegister(P2PTxFilteredIncoming)

	P2PTxFilteredOutgoing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_p2p_tx_filtered_outgoing",
		Help: "Transactions broadcast to other nodes filtered by anti-spam strategy",
	})
	prometheus.MustRegister(P2PTxFilteredOutgoing)

	blockchainIDMsbGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_last_block_id_msb",
		Help: "Blockchain last block id MSB",
	}, []string{"chaintype"})
	prometheus.MustRegister(blockchainIDMsbGaugeVector)

	blockchainIDLsbGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_last_block_id_lsb",
		Help: "Blockchain last block id LSB",
	}, []string{"chaintype"})
	prometheus.MustRegister(blockchainIDLsbGaugeVector)

	blockchainHeightGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_blockchain_height",
		Help: "Blockchain height",
	}, []string{"chaintype"})
	prometheus.MustRegister(blockchainHeightGaugeVector)

	goRoutineActivityGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_routines_counter",
		Help: "Go routine counter for",
	}, []string{"activity"})
	prometheus.MustRegister(goRoutineActivityGaugeVector)

	downloadCycleDebuggerGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_download_cycle_debugger",
		Help: "download cycle debugger for each chain",
	}, []string{"chaintype"})
	prometheus.MustRegister(downloadCycleDebuggerGaugeVector)

	apiGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_incoming_api_calls",
		Help: "Response time of api calls",
	}, []string{"api_name"})
	prometheus.MustRegister(apiGaugeVector)

	apiRunningGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_incoming_api_running",
		Help: "Counts how many request of each api is being handled",
	}, []string{"api_name"})
	prometheus.MustRegister(apiRunningGaugeVector)

	snapshotDownloadRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "zoobc_snapshot_chunk_downloads_status",
		Help: "snapshot file chunks to download",
	}, []string{"status"})
	prometheus.MustRegister(snapshotDownloadRequestCounter)

	dbStatGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_db_stats",
		Help: "Log the database connection status",
	}, []string{"status"})
	prometheus.MustRegister(dbStatGaugeVector)

	blockProcessTimeGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_block_process_time_ms",
		Help: "Block process time",
	}, []string{"block_height"})
	prometheus.MustRegister(blockProcessTimeGaugeVector)

	mempoolTransactionCountGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_mempool_transaction_count",
		Help: "Mempool count",
	}, []string{"block_height"})
	prometheus.MustRegister(mempoolTransactionCountGaugeVector)

	// Cache Storage
	cacheStorageGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_cache_storage",
		Help: "Cache storage usage in bytes",
	}, []string{"cache_type"})
	prometheus.MustRegister(cacheStorageGaugeVector)
}

func SetCLIMonitoring(cliMonitoring CLIMonitoringInteface) {
	cliMonitoringInstance = cliMonitoring
}

func SetNodePublicKey(pk []byte) {
	nodePublicKey = pk
}

func IsMonitoringActive() bool {
	return isMonitoringActive
}

func IncrementSendAddressInfoToPeer() {
	if !isMonitoringActive {
		return
	}

	sendAddressInfoToPeer.Inc()
}

func IncrementGetAddressInfoTableFromPeer() {
	if !isMonitoringActive {
		return
	}

	getAddressInfoTableFromPeer.Inc()
}

func IncrementReceiptCounter() {
	if !isMonitoringActive {
		return
	}

	receiptCounter.Inc()
}

func SetNodeAddressInfoCount(count int) {
	if !isMonitoringActive {
		return
	}

	nodeAddressInfoCounter.Set(float64(count))
}

func SetNodeAddressStatusCount(count int, status model.NodeAddressStatus) {
	if !isMonitoringActive {
		return
	}

	switch status {
	case model.NodeAddressStatus_NodeAddressPending:
		pendingAddressCounter.Set(float64(count))
	case model.NodeAddressStatus_NodeAddressConfirmed:
		confirmedAddressCounter.Set(float64(count))
	default:
		return
	}
}

func SetUnresolvedPeersCount(count int) {
	if cliMonitoringInstance != nil {
		cliMonitoringInstance.UpdatePeersInfo(CLIMonitoringUnresolvedPeersNumber, count)
	}

	if !isMonitoringActive {
		return
	}

	unresolvedPeersCounter.Set(float64(count))
}

func SetResolvedPeersCount(count int) {
	if cliMonitoringInstance != nil {
		cliMonitoringInstance.UpdatePeersInfo(CLIMonitoringResolvePeersNumber, count)
	}
	if !isMonitoringActive {
		return
	}

	resolvedPeersCounter.Set(float64(count))
}

func SetResolvedPriorityPeersCount(count int) {
	if cliMonitoringInstance != nil {
		cliMonitoringInstance.UpdatePeersInfo(CLIMonitoringResolvedPriorityPeersNumber, count)
	}
	if !isMonitoringActive {
		return
	}
	resolvedPeersCount = uint32(count)
	resolvedPriorityPeersCounter.Set(float64(count))
}

func SetUnresolvedPriorityPeersCount(count int) {
	if cliMonitoringInstance != nil {
		cliMonitoringInstance.UpdatePeersInfo(CLIMonitoringUnresolvedPriorityPeersNumber, count)
	}
	if !isMonitoringActive {
		return
	}
	unresolvedPeersCount = uint32(count)
	unresolvedPriorityPeersCounter.Set(float64(count))
}

func SetActiveRegisteredNodesCount(count int) {
	if !isMonitoringActive {
		return
	}

	activeRegisteredNodesGauge.Set(float64(count))
}

func IncrementBlockerMetrics(typeBlocker string) {
	if !isMonitoringActive {
		return
	}

	blockerCounterVector.WithLabelValues(typeBlocker).Inc()
}

func IncrementStatusLockCounter(chaintype chaintype.ChainType, typeStatusLock int) {
	if !isMonitoringActive {
		return
	}

	statusLockGaugeVector.WithLabelValues(chaintype.GetName(), fmt.Sprintf("%d", typeStatusLock)).Inc()
}

func DecrementStatusLockCounter(chaintype chaintype.ChainType, typeStatusLock int) {
	if !isMonitoringActive {
		return
	}

	statusLockGaugeVector.WithLabelValues(chaintype.GetName(), fmt.Sprintf("%d", typeStatusLock)).Dec()
}

func SetBlockchainStatus(chainType chaintype.ChainType, newStatus int) {
	if !isMonitoringActive {
		return
	}

	blockchainStatusGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(newStatus))
}

func SetBlockchainSmithIndex(chainType chaintype.ChainType, index int64) {
	if !isMonitoringActive {
		return
	}
	blocksmithIndex = int32(index)
	blockchainSmithIndexGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(index))
}

func SetNodeScore(score int64) {
	if !isMonitoringActive {
		return
	}

	nodeScore.Set(float64(score))
}

func SetTpsReceived(tps int) {
	if !isMonitoringActive {
		return
	}
	tpsReceived.Set(float64(tps))
}

func SetTpsProcessed(tps int) {
	if !isMonitoringActive {
		return
	}
	tpsProcessed.Set(float64(tps))
}

func IncreaseTxReceived() {
	if !isMonitoringActive {
		return
	}
	txReceived.Inc()
}

func IncreaseTxProcessed() {
	if !isMonitoringActive {
		return
	}
	txProcessed.Inc()
}

func IncreaseTxFiltered() {
	if !isMonitoringActive {
		return
	}
	txFiltered.Inc()
}

func IncreaseP2PTxFilteredIncoming() {
	if !isMonitoringActive {
		return
	}
	P2PTxFilteredIncoming.Inc()
}

func IncreaseP2PTxFilteredOutgoing() {
	if !isMonitoringActive {
		return
	}
	P2PTxFilteredOutgoing.Inc()
}

func SetNextSmith(sortedBlocksmiths []*model.Blocksmith, sortedBlocksmithsMap map[string]*int64) {
	if cliMonitoringInstance != nil {
		cliMonitoringInstance.UpdateSmithingInfo(sortedBlocksmiths, sortedBlocksmithsMap)
	}
}

func SetLastBlock(chainType chaintype.ChainType, block *model.Block) {
	if cliMonitoringInstance != nil {
		cliMonitoringInstance.UpdateBlockState(chainType, block)
	}

	if !isMonitoringActive {
		return
	}
	if chainType.GetTypeInt() == (&chaintype.MainChain{}).GetTypeInt() {
		lastMainBlock = *block
	} else {
		lastSpineBlock = *block
	}
	blockchainIDMsbGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(block.GetID() / int64(1000000000)))
	blockchainIDLsbGaugeVector.WithLabelValues(chainType.GetName()).Set(math.Abs(float64(block.GetID() % int64(1000000000))))
	blockchainHeightGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(block.GetHeight()))
}

func SetBlockProcessTime(timeMs int64) {
	if !isMonitoringActive {
		return
	}
	blockProcessTimeGaugeVector.WithLabelValues("BlockProcessTime").Set(float64(timeMs))
}

func SetMempoolTransactionCount(mempoolTxCount int) {
	if !isMonitoringActive {
		return
	}
	mempoolTransactionCountGaugeVector.WithLabelValues("MempoolTransactionCount").Set(float64(mempoolTxCount))
}

func IncrementGoRoutineActivity(activityName string) {
	if !isMonitoringActive {
		return
	}

	goRoutineActivityGaugeVector.WithLabelValues(activityName).Inc()
}

func DecrementGoRoutineActivity(activityName string) {
	if !isMonitoringActive {
		return
	}

	goRoutineActivityGaugeVector.WithLabelValues(activityName).Dec()
}

func IncrementMainchainDownloadCycleDebugger(chainType chaintype.ChainType, cycleMarker int) {
	if !isMonitoringActive {
		return
	}

	downloadCycleDebuggerGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(cycleMarker))
}

func ResetMainchainDownloadCycleDebugger(chainType chaintype.ChainType) {
	if !isMonitoringActive {
		return
	}

	downloadCycleDebuggerGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(-1))
}

func SetAPIResponseTime(apiName string, responseTime float64) {
	if !isMonitoringActive {
		return
	}

	apiGaugeVector.WithLabelValues(apiName).Set(responseTime)
}

func IncrementRunningAPIHandling(apiName string) {
	if !isMonitoringActive {
		return
	}

	apiRunningGaugeVector.WithLabelValues(apiName).Inc()
}

func DecrementRunningAPIHandling(apiName string) {
	if !isMonitoringActive {
		return
	}

	apiRunningGaugeVector.WithLabelValues(apiName).Dec()
}

func IncrementSnapshotDownloadCounter(succeeded, failed int32) {
	if !isMonitoringActive {
		return
	}

	if succeeded > 0 {
		snapshotDownloadRequestCounter.WithLabelValues("success").Add(float64(succeeded))
	}
	if failed > 0 {
		snapshotDownloadRequestCounter.WithLabelValues("failed").Add(float64(failed))
	}
}

func SetDatabaseStats(dbStat sql.DBStats) {
	if !isMonitoringActive {
		return
	}

	dbStatGaugeVector.WithLabelValues("OpenConnections").Set(float64(dbStat.OpenConnections))
	dbStatGaugeVector.WithLabelValues("ConnectionsInUse").Set(float64(dbStat.InUse))
	dbStatGaugeVector.WithLabelValues("ConnectionsWaitCount").Set(float64(dbStat.WaitCount))
}

type (
	// CacheStorageType type of cache storage that needed for inc or dec the value
	CacheStorageType string
)

// Cache Storage environments
// Please add new one when add new cache storage instance
var (
	TypeMempoolCacheStorage         CacheStorageType = "mempools"
	TypeBatchReceiptCacheStorage    CacheStorageType = "batch_receipts"
	TypeScrambleNodeCacheStorage    CacheStorageType = "scramble_nodes"
	TypeMempoolBackupCacheStorage   CacheStorageType = "backup_mempools"
	TypeNodeShardCacheStorage       CacheStorageType = "node_shards"
	TypeNodeAddressInfoCacheStorage CacheStorageType = "node_address_infos"
	TypeActiveNodeRegistryStorage   CacheStorageType = "node_registry_active"
	TypePendingNodeRegistryStorage  CacheStorageType = "node_registry_pending"
	TypeMainBlocksCacheStorage      CacheStorageType = "main_blocks_cache_object"
	TypeSpineBlocksCacheStorage     CacheStorageType = "spine_blocks_cache_object"
)

func SetCacheStorageMetrics(cacheType CacheStorageType, size float64) {
	if isMonitoringActive {
		cacheStorageGaugeVector.WithLabelValues(string(cacheType)).Set(size)
	}
}

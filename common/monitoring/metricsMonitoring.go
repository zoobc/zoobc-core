package monitoring

import (
	"database/sql"
	"fmt"
	"math"
	"reflect"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	isMonitoringActive bool
	nodePublicKey      []byte

	receiptCounter                   prometheus.Counter
	unresolvedPeersCounter           prometheus.Gauge
	resolvedPeersCounter             prometheus.Gauge
	unresolvedPriorityPeersCounter   prometheus.Gauge
	resolvedPriorityPeersCounter     prometheus.Gauge
	activeRegisteredNodesGauge       prometheus.Gauge
	nodeScore                        prometheus.Gauge
	blockerCounterVector             *prometheus.CounterVec
	statusLockGaugeVector            *prometheus.GaugeVec
	blockchainStatusGaugeVector      *prometheus.GaugeVec
	blockchainSmithTimeGaugeVector   *prometheus.GaugeVec
	blockchainIDMsbGaugeVector       *prometheus.GaugeVec
	blockchainIDLsbGaugeVector       *prometheus.GaugeVec
	blockchainHeightGaugeVector      *prometheus.GaugeVec
	goRoutineActivityGaugeVector     *prometheus.GaugeVec
	downloadCycleDebuggerGaugeVector *prometheus.GaugeVec
	apiGaugeVector                   *prometheus.GaugeVec
	apiRunningGaugeVector            *prometheus.GaugeVec
	snapshotDownloadRequestCounter   *prometheus.CounterVec
	dbStatGaugeVector                *prometheus.GaugeVec
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

	P2pGetPeerInfoClient                = "P2pGetPeerInfoClient"
	P2pGetMorePeersClient               = "P2pGetMorePeersClient"
	P2pSendPeersClient                  = "P2pSendPeersClient"
	P2pSendBlockClient                  = "P2pSendBlockClient"
	P2pSendTransactionClient            = "P2pSendTransactionClient"
	P2pRequestBlockTransactionsClient   = "P2pRequestBlockTransactionsClient"
	P2pGetCumulativeDifficultyClient    = "P2pGetCumulativeDifficultyClient"
	P2pGetCommonMilestoneBlockIDsClient = "P2pGetCommonMilestoneBlockIDsClient"
	P2pGetNextBlockIDsClient            = "P2pGetNextBlockIDsClient"
	P2pGetNextBlocksClient              = "P2pGetNextBlocksClient"
	P2pRequestFileDownloadClient        = "P2pRequestFileDownloadClient"
)

func SetMonitoringActive(isActive bool) {
	isMonitoringActive = isActive

	receiptCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "zoobc_receipts",
		Help: "receipts counter",
	})
	prometheus.MustRegister(receiptCounter)

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

	blockchainSmithTimeGaugeVector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zoobc_blockchain_smith_time",
		Help: "Smith time of each nodes to smith for each chain",
	}, []string{"chaintype"})
	prometheus.MustRegister(blockchainSmithTimeGaugeVector)

	nodeScore = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zoobc_node_score",
		Help: "The score of the node (divided by 100 to fit the max float64)",
	})
	prometheus.MustRegister(nodeScore)

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

}

func SetNodePublicKey(pk []byte) {
	nodePublicKey = pk
}

func IsMonitoringActive() bool {
	return isMonitoringActive
}

func IncrementReceiptCounter() {
	if !isMonitoringActive {
		return
	}

	receiptCounter.Inc()
}

func SetUnresolvedPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	unresolvedPeersCounter.Set(float64(count))
}

func SetResolvedPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	resolvedPeersCounter.Set(float64(count))
}

func SetResolvedPriorityPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	resolvedPriorityPeersCounter.Set(float64(count))
}

func SetUnresolvedPriorityPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

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

func SetBlockchainSmithTime(chainType chaintype.ChainType, newTime int64) {
	if !isMonitoringActive {
		return
	}

	blockchainSmithTimeGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(newTime))
}

func SetNodeScore(activeBlocksmiths []*model.Blocksmith) {
	if !isMonitoringActive {
		return
	}

	var scoreInt64 int64
	for _, blockSmith := range activeBlocksmiths {
		if reflect.DeepEqual(blockSmith.NodePublicKey, nodePublicKey) {
			scoreInt64 = blockSmith.Score.Int64()
			break
		}
	}

	nodeScore.Set(float64(scoreInt64))
}

func SetLastBlock(chainType chaintype.ChainType, block *model.Block) {
	if !isMonitoringActive {
		return
	}

	blockchainIDMsbGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(block.GetID() / int64(1000000000)))
	blockchainIDLsbGaugeVector.WithLabelValues(chainType.GetName()).Set(math.Abs(float64(block.GetID() % int64(1000000000))))
	blockchainHeightGaugeVector.WithLabelValues(chainType.GetName()).Set(float64(block.GetHeight()))
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

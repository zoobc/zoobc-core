package monitoring

import (
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zoobc/zoobc-core/common/model"
)

type lastblockMetrics struct {
	IDMsb  prometheus.Gauge
	IDLsb  prometheus.Gauge
	Height prometheus.Gauge
}

var (
	isMonitoringActive bool
	nodePublicKey      []byte

	receiptCounter     prometheus.Counter
	receiptCounterSync sync.Mutex

	unresolvedPeersCounter     prometheus.Gauge
	unresolvedPeersCounterSync sync.Mutex

	resolvedPeersCounter     prometheus.Gauge
	resolvedPeersCounterSync sync.Mutex

	unresolvedPriorityPeersCounter     prometheus.Gauge
	unresolvedPriorityPeersCounterSync sync.Mutex

	resolvedPriorityPeersCounter     prometheus.Gauge
	resolvedPriorityPeersCounterSync sync.Mutex

	activeRegisteredNodesGauge     prometheus.Gauge
	activeRegisteredNodesGaugeSync sync.Mutex

	nodeScore     prometheus.Gauge
	nodeScoreSync sync.Mutex

	blockerCounter     = make(map[string]prometheus.Counter)
	blockerCounterSync sync.Mutex

	statusLockCounter     = make(map[int]prometheus.Gauge)
	statusLockCounterSync sync.Mutex

	blockchainStatus     = make(map[int32]prometheus.Gauge)
	blockchainStatusSync sync.Mutex

	blockchainSmithTime     = make(map[int32]prometheus.Gauge)
	blockchainSmithTimeSync sync.Mutex

	blockchainHeight     = make(map[int32]*lastblockMetrics)
	blockchainHeightSync sync.Mutex

	goRoutineActivityCounters     = make(map[string]prometheus.Gauge)
	goRoutineActivityCountersSync sync.Mutex

	downloadCycleDebugger     = make(map[int]prometheus.Gauge)
	downloadCycleDebuggerSync sync.Mutex
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
)

func SetMonitoringActive(isActive bool) {
	isMonitoringActive = isActive
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

	receiptCounterSync.Lock()
	defer receiptCounterSync.Unlock()
	if receiptCounter == nil {
		receiptCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("zoobc_receipts"),
			Help: fmt.Sprintf("receipts counter"),
		})
		prometheus.MustRegister(receiptCounter)
	}

	receiptCounter.Inc()
}

func SetUnresolvedPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	unresolvedPeersCounterSync.Lock()
	defer unresolvedPeersCounterSync.Unlock()

	if unresolvedPeersCounter == nil {
		unresolvedPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_unresolved_peers"),
			Help: fmt.Sprintf("unresolvedPeers counter"),
		})
		prometheus.MustRegister(unresolvedPeersCounter)
	}

	unresolvedPeersCounter.Set(float64(count))
}

func SetResolvedPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	resolvedPeersCounterSync.Lock()
	defer resolvedPeersCounterSync.Unlock()

	if resolvedPeersCounter == nil {
		resolvedPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_resolved_peers"),
			Help: fmt.Sprintf("resolvedPeers counter"),
		})
		prometheus.MustRegister(resolvedPeersCounter)
	}

	resolvedPeersCounter.Set(float64(count))
}

func SetResolvedPriorityPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	resolvedPriorityPeersCounterSync.Lock()
	defer resolvedPriorityPeersCounterSync.Unlock()

	if resolvedPriorityPeersCounter == nil {
		resolvedPriorityPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_resolved_priority_peers"),
			Help: fmt.Sprintf("priority resolvedPeers counter"),
		})
		prometheus.MustRegister(resolvedPriorityPeersCounter)
	}

	resolvedPriorityPeersCounter.Set(float64(count))
}

func SetUnresolvedPriorityPeersCount(count int) {
	if !isMonitoringActive {
		return
	}

	unresolvedPriorityPeersCounterSync.Lock()
	defer unresolvedPriorityPeersCounterSync.Unlock()

	if unresolvedPriorityPeersCounter == nil {
		unresolvedPriorityPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_unresolved_priority_peers"),
			Help: fmt.Sprintf("priority resolvedPeers counter"),
		})
		prometheus.MustRegister(unresolvedPriorityPeersCounter)
	}

	unresolvedPriorityPeersCounter.Set(float64(count))
}

func SetActiveRegisteredNodesCount(count int) {
	if !isMonitoringActive {
		return
	}

	activeRegisteredNodesGaugeSync.Lock()
	defer activeRegisteredNodesGaugeSync.Unlock()

	if activeRegisteredNodesGauge == nil {
		activeRegisteredNodesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_active_registered_nodes"),
			Help: fmt.Sprintf("active registered nodes counter"),
		})
		prometheus.MustRegister(activeRegisteredNodesGauge)
	}

	activeRegisteredNodesGauge.Set(float64(count))
}

func IncrementBlockerMetrics(typeBlocker string) {
	if !isMonitoringActive {
		return
	}

	blockerCounterSync.Lock()
	defer blockerCounterSync.Unlock()

	if blockerCounter[typeBlocker] == nil {
		blockerCounter[typeBlocker] = prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("zoobc_err_%s", typeBlocker),
			Help: fmt.Sprintf("Error %s counter", typeBlocker),
		})
		prometheus.MustRegister(blockerCounter[typeBlocker])
	}
	blockerCounter[typeBlocker].Inc()
}

func IncrementStatusLockCounter(typeStatusLock int) {
	if !isMonitoringActive {
		return
	}

	statusLockCounterSync.Lock()
	defer statusLockCounterSync.Unlock()

	if statusLockCounter[typeStatusLock] == nil {
		statusLockCounter[typeStatusLock] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_status_lock_%d", typeStatusLock),
			Help: fmt.Sprintf("Status lock %d counter", typeStatusLock),
		})
		prometheus.MustRegister(statusLockCounter[typeStatusLock])
		statusLockCounter[typeStatusLock].Set(float64(1))
	} else {
		statusLockCounter[typeStatusLock].Inc()
	}

}

func DecrementStatusLockCounter(typeStatusLock int) {
	if !isMonitoringActive {
		return
	}

	statusLockCounterSync.Lock()
	defer statusLockCounterSync.Unlock()

	if statusLockCounter[typeStatusLock] == nil {
		statusLockCounter[typeStatusLock] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_status_lock_%d", typeStatusLock),
			Help: fmt.Sprintf("Status lock %d counter", typeStatusLock),
		})
		prometheus.MustRegister(statusLockCounter[typeStatusLock])

		// to avoid below as the initial value, on creation on decrement, we exit
		return
	}
	statusLockCounter[typeStatusLock].Dec()
}

func SetBlockchainStatus(chainType int32, newStatus int) {
	if !isMonitoringActive {
		return
	}

	blockchainStatusSync.Lock()
	defer blockchainStatusSync.Unlock()

	if blockchainStatus[chainType] == nil {
		blockchainStatus[chainType] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_blockchain_status_%d", chainType),
			Help: fmt.Sprintf("Blockchain %d status", chainType),
		})
		prometheus.MustRegister(blockchainStatus[chainType])
	}
	blockchainStatus[chainType].Set(float64(newStatus))
}

func SetBlockchainSmithTime(chainType int32, newTime int64) {
	if !isMonitoringActive {
		return
	}

	blockchainSmithTimeSync.Lock()
	defer blockchainSmithTimeSync.Unlock()

	if blockchainSmithTime[chainType] == nil {
		blockchainSmithTime[chainType] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_blockchain_%d_smith_time", chainType),
			Help: fmt.Sprintf("Smith time of each nodes to smith for blockchain %d", chainType),
		})
		prometheus.MustRegister(blockchainSmithTime[chainType])
	}
	blockchainSmithTime[chainType].Set(float64(newTime))
}

func SetNodeScore(activeBlocksmiths []*model.Blocksmith) {
	if !isMonitoringActive {
		return
	}

	nodeScoreSync.Lock()
	defer nodeScoreSync.Unlock()

	if nodeScore == nil {
		nodeScore = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "zoobc_node_score",
			Help: "The score of the node (divided by 100 to fit the max float64)",
		})
		prometheus.MustRegister(nodeScore)
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

func SetLastBlock(chainType int32, block *model.Block) {
	if !isMonitoringActive {
		return
	}

	blockchainHeightSync.Lock()
	defer blockchainHeightSync.Unlock()

	if blockchainHeight[chainType] == nil {
		idMsbMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_blockchain_id_%d_msb", chainType),
			Help: fmt.Sprintf("Blockchain %d id MSB", chainType),
		})
		idLsbMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_blockchain_id_%d_lsb", chainType),
			Help: fmt.Sprintf("Blockchain %d id lsb", chainType),
		})
		heightMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_blockchain_height_%d", chainType),
			Help: fmt.Sprintf("Blockchain %d height", chainType),
		})

		blockchainHeight[chainType] = &lastblockMetrics{
			IDMsb:  idMsbMetrics,
			IDLsb:  idLsbMetrics,
			Height: heightMetrics,
		}
		prometheus.MustRegister(idMsbMetrics)
		prometheus.MustRegister(idLsbMetrics)
		prometheus.MustRegister(heightMetrics)
	}
	blockchainHeight[chainType].IDMsb.Set(float64(block.GetID() / int64(1000000000)))
	blockchainHeight[chainType].IDLsb.Set(math.Abs(float64(block.GetID() % int64(1000000000))))
	blockchainHeight[chainType].Height.Set(float64(block.GetHeight()))
}

func IncrementGoRoutineActivity(activityName string) {
	if !isMonitoringActive {
		return
	}

	goRoutineActivityCountersSync.Lock()
	defer goRoutineActivityCountersSync.Unlock()

	if goRoutineActivityCounters[activityName] == nil {
		goRoutineActivityCounters[activityName] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_routines_counter_%s", activityName),
			Help: fmt.Sprintf("Go routine counter for %s", activityName),
		})
		prometheus.MustRegister(goRoutineActivityCounters[activityName])
	}
	goRoutineActivityCounters[activityName].Inc()
}

func DecrementGoRoutineActivity(activityName string) {
	if !isMonitoringActive {
		return
	}

	goRoutineActivityCountersSync.Lock()
	defer goRoutineActivityCountersSync.Unlock()

	if goRoutineActivityCounters[activityName] == nil {
		goRoutineActivityCounters[activityName] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_routines_counter_%s", activityName),
			Help: fmt.Sprintf("Go routine counter for %s", activityName),
		})
		prometheus.MustRegister(goRoutineActivityCounters[activityName])

		// to avoid below as the initial value, on creation on decrement, we exit
		return
	}
	goRoutineActivityCounters[activityName].Dec()
}

func IncrementMainchainDownloadCycleDebugger(chainType int32, cycleMarker int) {
	if !isMonitoringActive || chainType != 0 {
		return
	}

	downloadCycleDebuggerSync.Lock()
	defer downloadCycleDebuggerSync.Unlock()

	if downloadCycleDebugger[cycleMarker] == nil {
		downloadCycleDebugger[cycleMarker] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_download_cycle_debugger_%d", cycleMarker),
			Help: fmt.Sprintf("download cycle debugger for mainchain cycle number %d", cycleMarker),
		})
		prometheus.MustRegister(downloadCycleDebugger[cycleMarker])
	}
	downloadCycleDebugger[cycleMarker].Inc()
}

func ResetMainchainDownloadCycleDebugger(chainType int32) {
	if chainType != 0 {
		return
	}

	downloadCycleDebuggerSync.Lock()
	defer downloadCycleDebuggerSync.Unlock()

	for _, counter := range downloadCycleDebugger {
		counter.Set(0)
	}
}

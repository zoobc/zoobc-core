package monitoring

import (
	"fmt"
	"math"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zoobc/zoobc-core/common/model"
)

type lastblockMetrics struct {
	IDMsb  prometheus.Gauge
	IDLsb  prometheus.Gauge
	Height prometheus.Gauge
}

var (
	isMonitoringActive         bool
	receiptCounter             prometheus.Counter
	unresolvedPeersCounter     prometheus.Gauge
	resolvedPeersCounter       prometheus.Gauge
	activeRegisteredNodesGauge prometheus.Gauge
	blockerCounter             = make(map[string]prometheus.Counter)
	statusLockCounter          = make(map[int]prometheus.Gauge)
	blockchainStatus           = make(map[int32]prometheus.Gauge)
	blockchainSmithTime        = make(map[int32]prometheus.Gauge)
	nodeScores                 = make(map[int64]prometheus.Gauge)
	blockchainHeight           = make(map[int32]*lastblockMetrics)
)

func SetMonitoringActive(isActive bool) {
	isMonitoringActive = isActive
}

func IncrementReceiptCounter() {
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
	if resolvedPeersCounter == nil {
		resolvedPeersCounter = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_resolved_peers"),
			Help: fmt.Sprintf("resolvedPeers counter"),
		})
		prometheus.MustRegister(resolvedPeersCounter)
	}

	resolvedPeersCounter.Set(float64(count))
}

func SetActiveRegisteredNodesCount(count int) {
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

	if !isMonitoringActive {
		return
	}

	if statusLockCounter[typeStatusLock] == nil {
		statusLockCounter[typeStatusLock] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_status_lock_%d", typeStatusLock),
			Help: fmt.Sprintf("Status lock %d counter", typeStatusLock),
		})
		prometheus.MustRegister(statusLockCounter[typeStatusLock])
	} else {
		statusLockCounter[typeStatusLock].Dec()
	}
}

func SetBlockchainStatus(chainType int32, newStatus int) {
	if !isMonitoringActive {
		return
	}
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
	if blockchainSmithTime[chainType] == nil {
		blockchainSmithTime[chainType] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("zoobc_blockchain_%d_smith_time", chainType),
			Help: fmt.Sprintf("Smith time of each nodes to smith for blockchain %d", chainType),
		})
		prometheus.MustRegister(blockchainSmithTime[chainType])
	}
	blockchainSmithTime[chainType].Set(float64(newTime))
}

func SetNodeScores(scores []*model.Blocksmith) {
	if !isMonitoringActive {
		return
	}
	for _, score := range scores {
		if nodeScores[score.NodeID] == nil {
			var (
				signString = ""
				absScore   = score.NodeID
			)
			if absScore < 0 {
				signString = "i"
				absScore *= -1
			}
			nodeScores[score.NodeID] = prometheus.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("zoobc_node_score_%s%d", signString, absScore),
				Help: fmt.Sprintf("The score of the node %d (divided by 100 to fit the max float64)", score.NodeID),
			})
			prometheus.MustRegister(nodeScores[score.NodeID])
		}
		scoreInt64 := score.Score.Int64()
		nodeScores[score.NodeID].Set(float64(scoreInt64))
	}
}

func SetLastBlock(chainType int32, block *model.Block) {
	if !isMonitoringActive {
		return
	}

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

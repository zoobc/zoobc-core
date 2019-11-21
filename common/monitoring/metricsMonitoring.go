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
	isMonitoringActive bool
	blockerCounter     = make(map[string]prometheus.Counter)
	statusLockCounter  = make(map[int]prometheus.Gauge)
	blockchainStatus   = make(map[int32]prometheus.Gauge)
	blockchainHeight   = make(map[int32]*lastblockMetrics)
)

func SetMonitoringActive(isActive bool) {
	isMonitoringActive = isActive
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
		// statusLockCounter[typeStatusLock].Set(float64(-1))
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

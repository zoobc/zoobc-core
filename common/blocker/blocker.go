package blocker

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type (
	TypeBlocker string

	Blocker struct {
		Type    TypeBlocker
		Message string
	}
)

var (
	DBErr                   TypeBlocker = "DBErr"
	DBRowNotFound           TypeBlocker = "DBRowNotFound"
	BlockErr                TypeBlocker = "BlockErr"
	BlockNotFoundErr        TypeBlocker = "BlockNotFoundErr"
	RequestParameterErr     TypeBlocker = "RequestParameterErr"
	AppErr                  TypeBlocker = "AppErr"
	AuthErr                 TypeBlocker = "AuthErr"
	ValidationErr           TypeBlocker = "ValidationErr"
	DuplicateMempoolErr     TypeBlocker = "DuplicateMempoolErr"
	DuplicateTransactionErr TypeBlocker = "DuplicateTransactionErr"
	ParserErr               TypeBlocker = "ParserErr"
	ServerError             TypeBlocker = "ServerError"
	SmithingErr             TypeBlocker = "SmithingErr"
	ChainValidationErr      TypeBlocker = "ChainValidationErr"

	isMonitoringActive bool
	prometheusCounter  = make(map[TypeBlocker]prometheus.Counter)
)

func SetMonitoringActive(isActive bool) {
	isMonitoringActive = isActive
}

func NewBlocker(typeBlocker TypeBlocker, message string) error {
	if isMonitoringActive {
		if prometheusCounter[typeBlocker] == nil {
			prometheusCounter[typeBlocker] = prometheus.NewCounter(prometheus.CounterOpts{
				Name: fmt.Sprintf("zoobc_err_%s", typeBlocker),
				Help: fmt.Sprintf("Error %s counter", typeBlocker),
			})
			prometheus.MustRegister(prometheusCounter[typeBlocker])
		}
		prometheusCounter[typeBlocker].Inc()
	}
	return Blocker{
		Type:    typeBlocker,
		Message: message,
	}
}

func (e Blocker) Error() string {
	return fmt.Sprintf("%v: %v", e.Type, e.Message)
}

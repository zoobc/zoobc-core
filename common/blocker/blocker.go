package blocker

import (
	"encoding/json"
	"fmt"

	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	TypeBlocker string

	Blocker struct {
		Type    TypeBlocker
		Message string
		Data    interface{}
	}
)

var (
	isDebugMode bool

	DBErr                     TypeBlocker = "DBErr"
	DBRowNotFound             TypeBlocker = "DBRowNotFound"
	BlockErr                  TypeBlocker = "BlockErr"
	BlockNotFoundErr          TypeBlocker = "BlockNotFoundErr"
	RequestParameterErr       TypeBlocker = "RequestParameterErr"
	AppErr                    TypeBlocker = "AppErr"
	AuthErr                   TypeBlocker = "AuthErr"
	ValidationErr             TypeBlocker = "ValidationErr"
	DuplicateMempoolErr       TypeBlocker = "DuplicateMempoolErr"
	DuplicateTransactionErr   TypeBlocker = "DuplicateTransactionErr"
	ParserErr                 TypeBlocker = "ParserErr"
	ServerError               TypeBlocker = "ServerError"
	SmithingErr               TypeBlocker = "SmithingErr"
	ZeroParticipationScoreErr TypeBlocker = "ZeroParticipationScoreErr"
	ChainValidationErr        TypeBlocker = "ChainValidationErr"
	P2PPeerError              TypeBlocker = "P2PPeerError"
	P2PNetworkConnectionErr   TypeBlocker = "P2PNetworkConnectionErr"
	SmithingPending           TypeBlocker = "SmithingPending"
	InvalidBlockTimestamp     TypeBlocker = "InvalidBlockTimestamp"
	TimeoutExceeded           TypeBlocker = "TimeoutExceeded"
	PushMainBlockErr          TypeBlocker = "PushMainBlockErr"
	ValidateMainBlockErr      TypeBlocker = "ValidateMainBlockErr"
	PushSpineBlockErr         TypeBlocker = "PushSpineBlockErr"
	ValidateSpineBlockErr     TypeBlocker = "ValidateSpineBlockErr"
	SchedulerError            TypeBlocker = "SchedulerError"
)

func SetIsDebugMode(val bool) {
	isDebugMode = val
}

func NewBlocker(typeBlocker TypeBlocker, message string, data ...interface{}) error {
	monitoring.IncrementBlockerMetrics(string(typeBlocker))
	blocker := Blocker{
		Type:    typeBlocker,
		Message: message,
	}
	if isDebugMode {
		blocker.Data = data
	}
	return blocker
}

func (e Blocker) Error() string {
	if isDebugMode {
		j, _ := json.Marshal(e.Data)
		return fmt.Sprintf("%v: %v > %s", e.Type, e.Message, j)
	}
	return fmt.Sprintf("%v: %v", e.Type, e.Message)
}

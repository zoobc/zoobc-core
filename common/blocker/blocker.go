package blocker

import (
	"fmt"

	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	TypeBlocker string

	Blocker struct {
		Type    TypeBlocker
		Message string
	}
)

var (
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
	P2PNetworkConnectionErr   TypeBlocker = "P2PNetworkConnectionErr"
	TimeoutExceeded           TypeBlocker = "TimeoutExceeded"
)

func NewBlocker(typeBlocker TypeBlocker, message string) error {
	monitoring.IncrementBlockerMetrics(string(typeBlocker))
	return Blocker{
		Type:    typeBlocker,
		Message: message,
	}
}

func (e Blocker) Error() string {
	return fmt.Sprintf("%v: %v", e.Type, e.Message)
}

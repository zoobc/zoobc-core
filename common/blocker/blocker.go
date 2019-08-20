package blocker

import "fmt"

type (
	TypeBlocker string

	Blocker struct {
		Type    TypeBlocker
		Message string
	}
)

var (
	DBErr               TypeBlocker = "DBErr"
	BlockErr            TypeBlocker = "BlockErr"
	BlockNotFoundErr    TypeBlocker = "BlockNotFoundErr"
	RequestParameterErr TypeBlocker = "RequestParameterErr"
	AppErr              TypeBlocker = "AppErr"
	AuthErr             TypeBlocker = "AuthErr"
	ValidationErr       TypeBlocker = "ValidationErr"
	ServerError         TypeBlocker = "ServerError"
)

func NewBlocker(typeBlocker TypeBlocker, message string) error {
	return Blocker{
		Type:    typeBlocker,
		Message: message,
	}
}

func (e Blocker) Error() string {
	return fmt.Sprintf("%v: %v", e.Type, e.Message)
}

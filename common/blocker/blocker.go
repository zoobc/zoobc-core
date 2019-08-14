package blocker

type (
	TypeBlocker int

	blocker struct {
		Type    TypeBlocker
		Message string
	}
)

var (
	DBErr    TypeBlocker = 1
	BlockErr TypeBlocker = 2
)

func NewBlocker(typeBlocker TypeBlocker, message string) error {
	return blocker{
		Type:    typeBlocker,
		Message: message,
	}
}

func (e blocker) Error() string {
	return e.Message
}

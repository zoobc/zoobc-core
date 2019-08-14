package blocker

type (
	TypeBlocker int

	Blocker struct {
		Type    TypeBlocker
		Message string
	}
)

var (
	DBErr    TypeBlocker = 1
	BlockErr TypeBlocker = 2
)

func (e Blocker) Error() string {
	return e.Error()
}

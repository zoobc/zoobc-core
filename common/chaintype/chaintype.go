// chaintype package contains type of chain implementations
package chaintype

type (
	Chaintype interface {
		GetChaintypeName() string
		GetChainSmithingDelayTime() int64
	}
)

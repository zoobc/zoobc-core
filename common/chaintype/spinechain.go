package chaintype

type Spinechain struct{}

func (*Spinechain) GetChaintypeName() string {
	return "SPINECHAIN"
}

func (*Spinechain) GetChainSmithingDelayTime() int64 {
	return 3600
}

package chaintype

type Mainchain struct{}

func (*Mainchain) GetChaintypeName() string {
	return "MAINCHAIN"
}

func (*Mainchain) GetChainSmithingDelayTime() int64 {
	return 6
}

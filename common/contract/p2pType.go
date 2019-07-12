package contract

// P2PType is
type P2PType interface {
	InitService(myAddress string, port uint32, wellknownPeers []string) (P2PType, error)
	StartP2P(chaintype ChainType)
}

package model

type ScrambledNodes struct {
	IndexNodes   map[string]*int // if we use normal int, we won't be able to detect null values
	AddressNodes []*Peer
	BlockHeight  uint32
}

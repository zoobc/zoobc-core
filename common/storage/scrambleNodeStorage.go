package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	ScrambleCacheStackStorage struct {
		itemLimit int
		sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
)

func NewScrambleCacheStackStorage() *ScrambleCacheStackStorage {
	// store 2 times rollback height worth of scramble nodes
	return &ScrambleCacheStackStorage{
		itemLimit:      int(constant.MaxScrambleCacheRound),
		scrambledNodes: make([]model.ScrambledNodes, 0, constant.MaxScrambleCacheRound),
	}
}

func (s *ScrambleCacheStackStorage) Pop() error {
	if len(s.scrambledNodes) > 0 {
		s.Lock()
		defer s.Unlock()
		s.scrambledNodes = s.scrambledNodes[:len(s.scrambledNodes)-1]
		return nil
	}
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, float64(s.size()))
	}
	// no more to pop
	return blocker.NewBlocker(blocker.ValidationErr, "StackEmpty")
}

func (s *ScrambleCacheStackStorage) Push(item interface{}) error {
	scrambleCopy, ok := item.(model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNode")
	}
	s.Lock()
	defer s.Unlock()
	if len(s.scrambledNodes) >= s.itemLimit {
		s.scrambledNodes = s.scrambledNodes[1:] // remove first (oldest) cache to make room for new scramble
	}
	s.scrambledNodes = append(s.scrambledNodes, scrambleCopy)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, float64(s.size()))
	}
	return nil
}

func (s *ScrambleCacheStackStorage) PopTo(index uint32) error {
	if len(s.scrambledNodes) <= int(index) {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	s.Lock()
	defer s.Unlock()
	s.scrambledNodes = s.scrambledNodes[:index+1]
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, float64(s.size()))
	}
	return nil
}

func (s *ScrambleCacheStackStorage) GetAll(items interface{}) error {
	scrambledNodesCopy, ok := items.(*[]model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNodeList")
	}
	s.RLock()
	defer s.RUnlock()
	for _, scrambleNode := range s.scrambledNodes {
		var tempScramble = s.copy(scrambleNode)
		*scrambledNodesCopy = append(*scrambledNodesCopy, tempScramble)
	}
	return nil
}

func (s *ScrambleCacheStackStorage) GetAtIndex(index uint32, item interface{}) error {
	if int(index) >= len(s.scrambledNodes) {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	scrambleNodeCopy, ok := item.(*model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNode")
	}
	s.RLock()
	defer s.RUnlock()
	*scrambleNodeCopy = s.copy(s.scrambledNodes[int(index)])
	return nil
}

func (s *ScrambleCacheStackStorage) GetTop(item interface{}) error {
	s.RLock()
	defer s.RUnlock()
	topIndex := len(s.scrambledNodes)
	if topIndex == 0 {
		return blocker.NewBlocker(blocker.CacheEmpty, "EmptyScramble")
	}
	scrambleNodeCopy, ok := item.(*model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNode")
	}
	*scrambleNodeCopy = s.copy(s.scrambledNodes[topIndex-1])
	return nil
}

func (s *ScrambleCacheStackStorage) Clear() error {
	s.scrambledNodes = make([]model.ScrambledNodes, 0, s.itemLimit)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, 0)
	}
	return nil
}

func (s *ScrambleCacheStackStorage) size() int {
	var size int
	var (
		scrambleBytes bytes.Buffer
		enc           = gob.NewEncoder(&scrambleBytes)
	)
	_ = enc.Encode(s.scrambledNodes)
	size = scrambleBytes.Len()
	return size
}

func (s *ScrambleCacheStackStorage) copy(src model.ScrambledNodes) model.ScrambledNodes {
	var (
		result                  model.ScrambledNodes
		newIndexNodes           = make(map[string]*int)
		newNodePublicKeyToIDMap = make(map[string]int64)
		newPeers                = make([]*model.Peer, 0)
	)
	for i, node := range src.AddressNodes {
		idx := i
		newNodePublicKeyToIDMap[hex.EncodeToString(node.GetInfo().GetPublicKey())] = node.GetInfo().GetID()
		scrambleDNodeMapKey := fmt.Sprintf("%d", node.GetInfo().GetID())
		newIndexNodes[scrambleDNodeMapKey] = &idx
		tempPeer := model.Peer{
			Info: &model.Node{
				ID:            node.GetInfo().GetID(),
				PublicKey:     node.GetInfo().GetPublicKey(),
				SharedAddress: node.GetInfo().GetSharedAddress(),
				Address:       node.GetInfo().GetAddress(),
				Port:          node.GetInfo().GetPort(),
				AddressStatus: node.GetInfo().GetAddressStatus(),
				Version:       node.GetInfo().GetVersion(),
				CodeName:      node.GetInfo().GetCodeName(),
			},
			LastInboundRequest:  node.GetLastInboundRequest(),
			BlacklistingCause:   node.GetBlacklistingCause(),
			BlacklistingTime:    node.GetBlacklistingTime(),
			ResolvingTime:       node.GetResolvingTime(),
			ConnectionAttempted: node.GetConnectionAttempted(),
			UnresolvingTime:     node.GetUnresolvingTime(),
		}
		newPeers = append(newPeers, &tempPeer)
	}
	result = model.ScrambledNodes{
		IndexNodes:           newIndexNodes,
		NodePublicKeyToIDMap: newNodePublicKeyToIDMap,
		AddressNodes:         newPeers,
		BlockHeight:          src.BlockHeight,
	}
	return result
}

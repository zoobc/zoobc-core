package storage

import (
	"encoding/json"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"sync"
)

type (
	ScrambleCacheStackStorage struct {
		itemLimit int
		sync.RWMutex
		scrambledNodes [][]byte
	}
)

func NewScrambleCacheStackStorage() *ScrambleCacheStackStorage {
	// store 2 times rollback height worth of scramble nodes
	itemLimit := (constant.MinRollbackBlocks / constant.PriorityStrategyBuildScrambleNodesGap) * 2
	return &ScrambleCacheStackStorage{
		itemLimit:      int(itemLimit),
		scrambledNodes: make([][]byte, 0, itemLimit),
	}
}

func (s *ScrambleCacheStackStorage) Pop() error {
	if len(s.scrambledNodes) > 0 {
		s.Lock()
		defer s.Unlock()
		s.scrambledNodes = s.scrambledNodes[:len(s.scrambledNodes)-1]
		return nil
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
	marshaledScramble, err := json.Marshal(scrambleCopy)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "MarshalFail")
	}
	s.scrambledNodes = append(s.scrambledNodes, marshaledScramble)
	return nil
}

func (s *ScrambleCacheStackStorage) PopTo(index uint32) error {
	if len(s.scrambledNodes) <= int(index) {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	s.Lock()
	defer s.Unlock()
	s.scrambledNodes = s.scrambledNodes[:index+1]
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
		var tempScramble model.ScrambledNodes
		err := json.Unmarshal(scrambleNode, &tempScramble)
		if err != nil {
			return blocker.NewBlocker(blocker.ValidationErr, "UnmarshalFail")
		}
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
	err := json.Unmarshal(s.scrambledNodes[int(index)], scrambleNodeCopy)
	return err
}

func (s *ScrambleCacheStackStorage) Clear() error {
	s.scrambledNodes = make([][]byte, 0, s.itemLimit)
	return nil
}

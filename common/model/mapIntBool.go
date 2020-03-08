package model

import "sync"

type MapIntBool struct {
	sync.RWMutex
	internal map[int32]bool
}

func NewMapIntBool() *MapIntBool {
	return &MapIntBool{
		internal: make(map[int32]bool),
	}
}

func (rm *MapIntBool) Load(key int32) (value, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

func (rm *MapIntBool) Delete(key int32) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *MapIntBool) Store(key int32, value bool) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

func (rm *MapIntBool) Count() int {
	rm.RLock()
	result := len(rm.internal)
	rm.RUnlock()
	return result
}

func (rm *MapIntBool) Reset() {
	rm.Lock()
	rm.internal = NewMapIntBool().internal
	rm.Unlock()
}

func (rm *MapIntBool) GetMap() map[int32]bool {
	rm.RLock()
	result := rm.internal
	rm.RUnlock()
	return result
}

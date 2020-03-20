package model

import "sync"

type MapStringInt struct {
	sync.RWMutex
	internal map[string]int64
}

func NewMapStringInt() *MapStringInt {
	return &MapStringInt{
		internal: make(map[string]int64),
	}
}

func (rm *MapStringInt) Load(key string) (value int64, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

func (rm *MapStringInt) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *MapStringInt) Store(key string, value int64) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

func (rm *MapStringInt) Count() int {
	rm.RLock()
	result := len(rm.internal)
	rm.RUnlock()
	return result
}

func (rm *MapStringInt) Reset() {
	rm.Lock()
	rm.internal = NewMapStringInt().internal
	rm.Unlock()
}

func (rm *MapStringInt) GetMap() map[string]int64 {
	rm.RLock()
	result := rm.internal
	rm.RUnlock()
	return result
}

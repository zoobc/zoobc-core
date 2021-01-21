// Copyright 2020 Quasisoft Limited - Hong Kong.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package queue

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/monitoring"
)

type PriorityLock interface {
	Lock(ownerProcess int)
	Unlock()
	HighPriorityLock(ownerProcess int)
	HighPriorityUnlock()
}

// PriorityPreferenceLock implements a simple triple-mutex priority lock
// patterns are like:
//   Low Priority would do: lock lowPriorityMutex, wait for high priority groups, lock nextToAccess, lock dataMutex, unlock nextToAccess,
//   do stuff, unlock dataMutex, unlock lowPriorityMutex
//   High Priority would do: increment high priority waiting, lock nextToAccess, lock dataMutex, unlock nextToAccess, do stuff,
//   unlock dataMutex, decrement high priority waiting
type PriorityPreferenceLock struct {
	dataMutex           sync.Mutex
	nextToAccess        sync.Mutex
	lowPriorityMutex    sync.Mutex
	highPriorityWaiting sync.WaitGroup
}

func NewPriorityPreferenceLock() *PriorityPreferenceLock {
	lock := PriorityPreferenceLock{
		highPriorityWaiting: sync.WaitGroup{},
	}

	return &lock
}

// Lock will acquire a low-priority lock
// it must wait until both low priority and all high priority lock holders are released.
func (lock *PriorityPreferenceLock) Lock(ownerProcess int) {
	monitoring.IncrementDbLockCounter(0, ownerProcess)
	lock.lowPriorityMutex.Lock()
	lock.highPriorityWaiting.Wait()
	lock.nextToAccess.Lock()
	lock.dataMutex.Lock()
	lock.nextToAccess.Unlock()
}

// Unlock will unlock the low-priority lock
func (lock *PriorityPreferenceLock) Unlock() {
	lock.dataMutex.Unlock()
	lock.lowPriorityMutex.Unlock()
	monitoring.DecrementDbLockCounter(0)
}

// HighPriorityLock will acquire a high-priority lock
// it must still wait until a low-priority lock has been released and then potentially other high priority lock contenders.
func (lock *PriorityPreferenceLock) HighPriorityLock(ownerProcess int) {
	monitoring.IncrementDbLockCounter(1, ownerProcess)
	lock.highPriorityWaiting.Add(1)
	lock.nextToAccess.Lock()
	lock.dataMutex.Lock()
	lock.nextToAccess.Unlock()
}

// HighPriorityUnlock will unlock the high-priority lock
func (lock *PriorityPreferenceLock) HighPriorityUnlock() {
	lock.dataMutex.Unlock()
	lock.highPriorityWaiting.Done()
	monitoring.DecrementDbLockCounter(1)
}

// UnrestrictiveWaitGroup
// Substitution of sync.WaitGroup that doesn't allow adding more item while the wait has started
type UnrestrictiveWaitGroup struct {
	counter int64
	sync.Mutex
}

func (uwg *UnrestrictiveWaitGroup) Add(additional int64) {
	uwg.Lock()
	defer uwg.Unlock()
	uwg.counter += additional
}

func (uwg *UnrestrictiveWaitGroup) Done() {
	uwg.Lock()
	defer uwg.Unlock()
	uwg.counter--
}

func (uwg *UnrestrictiveWaitGroup) Wait() {
	for {
		uwg.Lock()
		if uwg.counter == 0 {
			break
		}
		uwg.Unlock()
	}
}

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
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

var mockData []string
var logger = log.New(os.Stdout, "", 0)

func lowPriorityRoutine(idx int, wg sync.WaitGroup, l PriorityLock, sleepDuration time.Duration, holdDuration time.Duration) {
	wg.Add(1)
	defer wg.Done()
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
	}

	l.Lock()
	defer l.Unlock()
	logger.Println(fmt.Sprintf("[%d] [idx:%d] Acquired low priority lock", time.Now().UnixNano(), idx))
	mockData = append(mockData, "lo")
	if holdDuration > 0 {
		time.Sleep(holdDuration)
	}
	logger.Println(fmt.Sprintf("[%d] [idx:%d] Releasing low priority lock", time.Now().UnixNano(), idx))
}

func highPriorityRoutine(idx int, wg sync.WaitGroup, l PriorityLock, sleepDuration time.Duration, holdDuration time.Duration) {
	wg.Add(1)
	defer wg.Done()
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
	}

	l.HighPriorityLock()
	defer l.HighPriorityUnlock()
	mockData = append(mockData, "hi")
	logger.Println(fmt.Sprintf("[%d] [idx:%d] Acquired high priority lock", time.Now().UnixNano(), idx))
	if holdDuration > 0 {
		time.Sleep(holdDuration)
	}
	logger.Println(fmt.Sprintf("[%d] [idx:%d] Releasing high priority lock", time.Now().UnixNano(), idx))
}

type testCaseRoutine struct {
	lockType      string
	sleepDuration time.Duration
	holdDuration  time.Duration
}

var defaultSleepTime = time.Duration(30) * time.Nanosecond

func TestPriorityPreferenceLock(t *testing.T) {

	lock := NewPriorityPreferenceLock()
	wg := sync.WaitGroup{}
	notifyChan := make(chan struct{}, 1)
	defer close(notifyChan)

	runners := []testCaseRoutine{
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "hi", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "hi", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "hi", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
		{lockType: "lo", sleepDuration: defaultSleepTime, holdDuration: defaultSleepTime},
	}

	for idx, runner := range runners {
		switch runner.lockType {
		case "lo":
			go lowPriorityRoutine(idx, wg, lock, runner.sleepDuration, runner.holdDuration)
		case "hi":
			go highPriorityRoutine(idx, wg, lock, runner.sleepDuration, runner.holdDuration)
		}
	}

	go func(c chan struct{}) {
		wg.Wait()
		c <- struct{}{}
	}(notifyChan)

	select {
	case <-notifyChan:
		// all unlocked, continue
		time.Sleep(time.Millisecond * 100) // allow logs to flush
	case <-time.After(time.Second * 10):
		t.Fatal("Did not complete locking simulation withing timeout")
	}

	// check all data operations succeeded
	if len(mockData) != len(runners) {
		t.Fatal(fmt.Sprintf("Expected at least %d entries in the mock data", len(runners)))
	}

	// Check ordering is biased hi - note: some lo's will happen before given the test case
	maxHighIdx := 8 // based on sampling
	lastHighIdx := 0
	for i, v := range mockData {
		if v == "hi" {
			lastHighIdx = i
		}
	}
	log.Println(fmt.Sprintf("lastHighIdx: %d", lastHighIdx))
	if lastHighIdx > maxHighIdx {
		t.Fatal(fmt.Sprintf("lastHighIdx should be <= %d but was %d", maxHighIdx, lastHighIdx))
	}

}

// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package feedbacksystem

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type (
	// DummyFeedbackStrategy implements FeedbackStrategyInterface and it is used to switch AntiSpam filter off
	DummyFeedbackStrategy struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
)

// DummyFeedbackStrategy initialize system internal variables
func NewDummyFeedbackStrategy() *DummyFeedbackStrategy {
	return &DummyFeedbackStrategy{
		FeedbackVars: map[string]interface{}{
			"tpsReceived":         0,
			"tpsReceivedTmp":      0,
			"tpsProcessed":        0,
			"tpsProcessedTmp":     0,
			"txReceived":          0,
			"txProcessed":         0,
			"P2PIncomingRequests": 0,
			"P2POutgoingRequests": 0,
		},
	}
}

func (dfs *DummyFeedbackStrategy) StartSampling(samplingInterval time.Duration) {
	tickerResetPerSecondVars := time.NewTicker(time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-tickerResetPerSecondVars.C:
			// Reset feedback variables that are sampled 'per second'
			func() {
				if tpsReceivedTmp := dfs.GetFeedbackVar("tpsReceivedTmp"); tpsReceivedTmp != nil {
					dfs.SetFeedbackVar("tpsReceived", tpsReceivedTmp)
					monitoring.SetTpsReceived(tpsReceivedTmp.(int))
				}
				if tpsProcessedTmp := dfs.GetFeedbackVar("tpsProcessedTmp"); tpsProcessedTmp != nil {
					dfs.SetFeedbackVar("tpsProcessed", tpsProcessedTmp)
					monitoring.SetTpsProcessed(tpsProcessedTmp.(int))
				}
				// Reset the temporary tps received/processed every second
				dfs.SetFeedbackVar("tpsReceivedTmp", 0)
				dfs.SetFeedbackVar("tpsProcessedTmp", 0)
			}()
		case <-sigs:
			tickerResetPerSecondVars.Stop()
			return
		}
	}
}

func (dfs *DummyFeedbackStrategy) GetSuggestedActions() map[constant.FeedbackAction]bool {
	return nil
}

func (dfs *DummyFeedbackStrategy) IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsP2PRequestLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsCPULimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) SetFeedbackVar(k string, v interface{}) {
	dfs.FeedbackVarsLock.Lock()
	defer dfs.FeedbackVarsLock.Unlock()
	dfs.FeedbackVars[k] = v
}

func (dfs *DummyFeedbackStrategy) GetFeedbackVar(k string) interface{} {
	dfs.FeedbackVarsLock.RLock()
	defer dfs.FeedbackVarsLock.RUnlock()
	v, ok := dfs.FeedbackVars[k]
	if !ok {
		return nil
	}
	return v
}

func (dfs *DummyFeedbackStrategy) IncrementVarCount(k string) interface{} {
	var (
		v        = dfs.GetFeedbackVar(k)
		newCount = 1
	)
	if v != nil {
		newCount = v.(int) + 1
		dfs.SetFeedbackVar(k, newCount)
	}
	return newCount
}

func (dfs *DummyFeedbackStrategy) DecrementVarCount(k string) interface{} {
	var (
		v        = dfs.GetFeedbackVar(k)
		newCount = 0
	)
	if v != nil {
		newCount = v.(int) - 1
		dfs.SetFeedbackVar(k, newCount)
	}
	return newCount
}

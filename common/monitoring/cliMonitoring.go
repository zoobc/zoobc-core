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
package monitoring

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tm "github.com/buger/goterm"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	CLIMonitoring struct {
		ConfigInfo        *model.Config
		BlocksInfo        map[int32]*model.Block
		PeersInfo         map[string]int
		SmithInfo         *model.Blocksmith
		NextSmithingIndex *int64
		PeersInfoLock     sync.RWMutex
		BlocksInfoLock    sync.RWMutex
	}
	CLIMonitoringInteface interface {
		UpdateBlockState(chaintype chaintype.ChainType, block *model.Block)
		UpdatePeersInfo(peersType string, peersNumber int)
		UpdateSmithingInfo(sortedBlocksmiths []*model.Blocksmith, sortedBlocksmithsMap map[string]*int64)
		Start()
	}
)

var (
	CLIMonitoringResolvePeersNumber            = "ResolvePeersNumber"
	CLIMonitoringUnresolvedPeersNumber         = "UnresolvedPeersNumber"
	CLIMonitoringResolvedPriorityPeersNumber   = "ResolvedPriorityPeersNumber"
	CLIMonitoringUnresolvedPriorityPeersNumber = "UnresolvedPriorityPeersNumber"
)

func NewCLIMonitoring(configInfo *model.Config) CLIMonitoringInteface {
	return &CLIMonitoring{
		ConfigInfo: configInfo,
	}
}

func (cm *CLIMonitoring) UpdateBlockState(chaintype chaintype.ChainType, block *model.Block) {
	cm.BlocksInfoLock.Lock()
	defer cm.BlocksInfoLock.Unlock()
	if cm.BlocksInfo == nil {
		cm.BlocksInfo = make(map[int32]*model.Block)
	}
	// Note: Sometimes received block is nil, when updating spine block in new joining block
	if block != nil {
		cm.BlocksInfo[chaintype.GetTypeInt()] = block
	}
}

func (cm *CLIMonitoring) UpdatePeersInfo(peersType string, peersNumber int) {
	cm.PeersInfoLock.Lock()
	defer cm.PeersInfoLock.Unlock()
	if cm.PeersInfo == nil {
		cm.PeersInfo = make(map[string]int)
	}
	cm.PeersInfo[peersType] = peersNumber

}

func (cm *CLIMonitoring) UpdateSmithingInfo(sortedBlocksmiths []*model.Blocksmith, sortedBlocksmithsMap map[string]*int64) {
	cm.NextSmithingIndex = sortedBlocksmithsMap[string(cm.ConfigInfo.NodeKey.PublicKey)]
	if cm.NextSmithingIndex != nil {
		if int(*cm.NextSmithingIndex) <= len(sortedBlocksmiths) { // safety check since we are using index of different model
			cm.SmithInfo = sortedBlocksmiths[*cm.NextSmithingIndex]
		}
	}
}

func (cm *CLIMonitoring) Start() {
	var (
		mainChain                       = &chaintype.MainChain{}
		spainChain                      = &chaintype.SpineChain{}
		nodePublicKey, errNodePublicKey = address.EncodeZbcID(
			constant.PrefixZoobcNodeAccount,
			cm.ConfigInfo.NodeKey.PublicKey,
		)
	)
	tm.Clear() // Clear current screen
	for {
		tm.MoveCursor(1, 1)
		cm.print("Application Codename", constant.ApplicationCodeName)
		cm.print("Application Version", constant.ApplicationVersion)
		cm.print("Node IP Address / DNS", cm.ConfigInfo.MyAddress)
		cm.print("Peer Communication Port", cm.ConfigInfo.PeerPort)
		cm.print("RPC API Port", cm.ConfigInfo.RPCAPIPort)
		cm.print("HTTP API Port", cm.ConfigInfo.HTTPAPIPort)
		cm.print("Monitoring Port", cm.ConfigInfo.MonitoringPort)
		cm.print("Well Known Peers", strings.Join(cm.ConfigInfo.WellknownPeers, ", "))

		if errNodePublicKey == nil {
			cm.print("Node Public Key", nodePublicKey)
		}
		cm.print("Owner Account Address", cm.ConfigInfo.OwnerAccountAddress)
		cm.print("Owner Account Address (hex)", cm.ConfigInfo.OwnerAccountAddressHex)
		cm.print("Owner Account Address (encoded)", cm.ConfigInfo.OwnerEncodedAccountAddress)
		cm.print("Smithing Status", cm.ConfigInfo.Smithing)
		cm.printLineBreak()

		cm.print(
			fmt.Sprintf("%s Block ID", spainChain.GetName()),
			cm.BlocksInfo[spainChain.GetTypeInt()].GetID(),
		)
		cm.print(
			fmt.Sprintf("%s Block Height", spainChain.GetName()),
			cm.BlocksInfo[spainChain.GetTypeInt()].GetHeight(),
		)

		cm.print(
			fmt.Sprintf("%s Block ID", mainChain.GetName()),
			cm.BlocksInfo[mainChain.GetTypeInt()].GetID(),
		)
		cm.print(
			fmt.Sprintf("%s Block Height", mainChain.GetName()),
			cm.BlocksInfo[mainChain.GetTypeInt()].GetHeight(),
		)
		cm.printLineBreak()

		cm.print("Resolved Peers Number", cm.PeersInfo[CLIMonitoringResolvePeersNumber])
		cm.print("Unresolved Peers Number", cm.PeersInfo[CLIMonitoringUnresolvedPeersNumber])
		if cm.NextSmithingIndex != nil {
			cm.print("Priority Resolved  Peers Number", cm.PeersInfo[CLIMonitoringResolvedPriorityPeersNumber])
			cm.print("Priority Unresolved Peers Number", cm.PeersInfo[CLIMonitoringUnresolvedPriorityPeersNumber])
			cm.printLineBreak()

			cm.print("Next Smithing Position", *cm.NextSmithingIndex)
			cm.print("Node ID", cm.SmithInfo.NodeID)
			cm.print("Node Score", cm.SmithInfo.Score)
		}

		// note. Please add number of clearLine as many as number of print in conditional state
		cm.clearLine(1)
		tm.Flush() // Call it every time at the end of rendering
		time.Sleep(2 * time.Second)
	}
}

func (*CLIMonitoring) print(label string, value interface{}) {
	tm.Printf("%s\n", tm.ResetLine(fmt.Sprintf("%s: %v", tm.Bold(label), value)))
}

func (*CLIMonitoring) printLineBreak() {
	if _, err := tm.Println(tm.ResetLine("")); err != nil {
		return
	}
}

// clearLine to clear unused line on the screen, numberLine depends on number of print in the conditional state
func (cm *CLIMonitoring) clearLine(numberLine int) {
	for i := 0; i < numberLine; i++ {
		cm.printLineBreak()
	}
}

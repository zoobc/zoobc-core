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

func (nl *CLIMonitoring) UpdateBlockState(chaintype chaintype.ChainType, block *model.Block) {
	nl.BlocksInfoLock.Lock()
	defer nl.BlocksInfoLock.Unlock()
	if nl.BlocksInfo == nil {
		nl.BlocksInfo = make(map[int32]*model.Block)
	}
	nl.BlocksInfo[chaintype.GetTypeInt()] = block
}

func (nl *CLIMonitoring) UpdatePeersInfo(peersType string, peersNumber int) {
	nl.PeersInfoLock.Lock()
	defer nl.PeersInfoLock.Unlock()
	if nl.PeersInfo == nil {
		nl.PeersInfo = make(map[string]int)
	}
	nl.PeersInfo[peersType] = peersNumber

}

func (nl *CLIMonitoring) UpdateSmithingInfo(sortedBlocksmiths []*model.Blocksmith, sortedBlocksmithsMap map[string]*int64) {
	nl.NextSmithingIndex = sortedBlocksmithsMap[string(nl.ConfigInfo.NodeKey.PublicKey)]
	if nl.NextSmithingIndex != nil {
		nl.SmithInfo = sortedBlocksmiths[*nl.NextSmithingIndex]
	}
}

func (nl *CLIMonitoring) Start() {
	var (
		mainChain                          = &chaintype.MainChain{}
		spainChain                         = &chaintype.SpineChain{}
		nodeAccountAddress, errNodeAddress = address.EncodeZbcID(
			constant.PrefixZoobcNodeAccount,
			nl.ConfigInfo.NodeKey.PublicKey,
		)
	)
	tm.Clear() // Clear current screen
	for {
		tm.MoveCursor(1, 1)

		tm.Printf("%s: %v\n", tm.Bold("Node IP Address / DNS "), nl.ConfigInfo.MyAddress)
		tm.Printf("%s: %v\n", tm.Bold("Peer Communication Port"), (nl.ConfigInfo.PeerPort))
		tm.Printf("%s: %v\n", tm.Bold("RPC API Port"), nl.ConfigInfo.RPCAPIPort)
		tm.Printf("%s: %v\n", tm.Bold("HTTP API Port"), nl.ConfigInfo.HTTPAPIPort)
		tm.Printf("%s: %v\n", tm.Bold("Monitoring Port"), nl.ConfigInfo.MonitoringPort)
		tm.Printf("%s: %v\n", tm.Bold("Well Known Peers"), strings.Join(nl.ConfigInfo.WellknownPeers, ","))

		if errNodeAddress == nil {
			tm.Printf("%s: %v\n", tm.Bold("Node Account Address"), nodeAccountAddress)
		}
		tm.Printf("%s: %v\n", tm.Bold("Owner Account Address"), nl.ConfigInfo.OwnerAccountAddress)
		tm.Printf("%s: %v\n\n", tm.Bold("Smithing Status"), nl.ConfigInfo.Smithing)

		if nl.BlocksInfo[spainChain.GetTypeInt()] != nil {
			tm.Printf("%s: %v\n",
				tm.Bold(fmt.Sprintf("%s Block ID", spainChain.GetName())),
				nl.BlocksInfo[spainChain.GetTypeInt()].GetID(),
			)
			tm.Printf("%s: %v\n",
				tm.Bold(fmt.Sprintf("%s Block Height", spainChain.GetName())),
				nl.BlocksInfo[spainChain.GetTypeInt()].GetHeight(),
			)
		}

		tm.Printf("%s: %v\n",
			tm.Bold(fmt.Sprintf("%s Block ID", mainChain.GetName())),
			nl.BlocksInfo[mainChain.GetTypeInt()].GetID(),
		)
		tm.Printf("%s: %v\n\n",
			tm.Bold(fmt.Sprintf("%s Block Height", mainChain.GetName())),
			nl.BlocksInfo[mainChain.GetTypeInt()].GetHeight(),
		)

		tm.Printf("%s: %v\n", tm.Bold("Resolved Peers Number"), nl.PeersInfo[CLIMonitoringResolvePeersNumber])
		tm.Printf("%s: %v\n", tm.Bold("Unresolved Peers Number"), nl.PeersInfo[CLIMonitoringUnresolvedPeersNumber])
		if nl.NextSmithingIndex != nil {
			tm.Printf("%s: %v\n", tm.Bold("Priority Resolved  Peers Number"), nl.PeersInfo[CLIMonitoringResolvedPriorityPeersNumber])
			tm.Printf("%s: %v\n\n", tm.Bold("Priority Unresolved Peers Number"), nl.PeersInfo[CLIMonitoringUnresolvedPriorityPeersNumber])

			tm.Printf("%s: %d\n", tm.Bold("Next Smithing Potition"), *nl.NextSmithingIndex)
			tm.Printf("%s: %v\n", tm.Bold("Node ID"), nl.SmithInfo.NodeID)
			tm.Printf("%s: %v\n", tm.Bold("Node Score"), nl.SmithInfo.Score)
		}

		tm.Flush() // Call it every time at the end of rendering
		time.Sleep(2 * time.Second)
	}
}

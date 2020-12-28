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
package service

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/zoobc/zoobc-core/common/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"runtime"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeHardwareServiceInterface interface {
		GetNodeHardware(request *model.GetNodeHardwareRequest) (*model.GetNodeHardwareResponse, error)
	}

	NodeHardwareService struct {
		OwnerAccountAddress []byte
		Signature           crypto.SignatureInterface
	}
)

func NewNodeHardwareService(
	ownerAccountAddress []byte,
	signature crypto.SignatureInterface,
) *NodeHardwareService {
	return &NodeHardwareService{
		OwnerAccountAddress: ownerAccountAddress,
		Signature:           signature,
	}
}

func (nhs *NodeHardwareService) GetNodeHardware(request *model.GetNodeHardwareRequest) (*model.GetNodeHardwareResponse, error) {
	var (
		nodeHardware *model.NodeHardware
		cpuStats     []*model.CPUInformation
		err          error
		diskStat     *disk.UsageStat
	)

	runtimeOS := runtime.GOOS
	// memory
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if runtimeOS == "windows" {
		diskStat, err = disk.Usage("\\")
	} else {
		diskStat, err = disk.Usage("/")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// host or machine kernel, uptime, platform Info
	hostStat, err := host.Info()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// cpu - get CPU number of cores and speed
	cpuStat, err := cpu.Info()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	percentage, err := cpu.Percent(0, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// host or machine kernel, uptime, platform Info
	for i := 0; i < len(cpuStat); i++ {
		cpuStats = append(cpuStats, &model.CPUInformation{
			Family:      cpuStat[i].Family,
			CPUIndex:    cpuStat[i].CPU,
			ModelName:   cpuStat[i].ModelName,
			VendorId:    cpuStat[i].VendorID,
			Mhz:         cpuStat[i].Mhz,
			CacheSize:   cpuStat[i].CacheSize,
			UsedPercent: percentage[i],
			Cores:       cpuStat[i].Cores,
			CoreID:      cpuStat[i].CoreID,
		})

	}
	nodeHardware = &model.NodeHardware{
		CPUInformation: cpuStats,
		MemoryInformation: &model.MemoryInformation{
			Total:       vmStat.Total,
			Free:        vmStat.Free,
			Available:   vmStat.Available,
			Used:        vmStat.Used,
			UsedPercent: vmStat.UsedPercent,
		},
		StorageInformation: &model.StorageInformation{
			FsType:      diskStat.Fstype,
			Total:       diskStat.Total,
			Free:        diskStat.Free,
			Used:        diskStat.Used,
			UsedPercent: diskStat.UsedPercent,
		},
		HostInformation: &model.HostInformation{
			Uptime:                 hostStat.Uptime,
			OS:                     hostStat.OS,
			Platform:               hostStat.Platform,
			PlatformFamily:         hostStat.PlatformFamily,
			PlatformVersion:        hostStat.PlatformVersion,
			NumberOfRunningProcess: hostStat.Procs,
			HostID:                 hostStat.HostID,
			HostName:               hostStat.Hostname,
		},
	}
	return &model.GetNodeHardwareResponse{
		NodeHardware: nodeHardware,
	}, nil
}

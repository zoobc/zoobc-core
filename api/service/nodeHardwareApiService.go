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

package util

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"runtime"
	"time"
)

// GetHwStats info about system cpu/mem usage
func GetHwStats(samplingInterval time.Duration) (cpuPercent float64, virtualMem mem.VirtualMemoryStat, swapMem mem.SwapMemoryStat) {
	vm := mem.VirtualMemoryStat{}
	sm := mem.SwapMemoryStat{}

	if v, err := mem.VirtualMemory(); err == nil {
		vm = *v
	}
	vm.Total /= 1024 * 1024
	vm.Used /= 1024 * 1024
	if v, err := mem.SwapMemory(); err == nil {
		sm = *v
	}
	sm.Total /= 1024 * 1024
	sm.Used /= 1024 * 1024

	ps, err := cpu.Percent(samplingInterval, true)
	if err == nil && len(ps) > 0 {
		for _, v := range ps {
			cpuPercent += v
		}
		cpuPercent /= float64(len(ps))
	}

	return cpuPercent, vm, sm
}

// GetGoRoutineStats info about concurrent processes spawned by main process
func GetGoRoutineStats() int {
	return runtime.NumGoroutine()
}

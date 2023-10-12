// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package psutil

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func psutilMemoryFromGopsUtil() (*Memory, error) {
	virt, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	swap, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	mem := &Memory{
		Virtual: MemoryStat{
			Total:       int(virt.Total) / 1024,
			Free:        int(virt.Free) / 1024,
			Used:        int(virt.Used) / 1024,
			UsedPercent: float32(virt.UsedPercent),
		},
		Swap: MemoryStat{
			Total:       int(swap.Total) / 1024,
			Free:        int(swap.Free) / 1024,
			Used:        int(swap.Used) / 1024,
			UsedPercent: float32(swap.UsedPercent),
		},
		Buffers: int(virt.Buffers) / 1024,
		Cached:  int(virt.Cached) / 1024,
	}
	return mem, nil
}

func psutilProcFromGopsUtil(cpuWatermark int) ([]*Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	newProcs := make([]*Process, 0)

	for _, proc := range procs {
		cpuPct, err := proc.CPUPercent()
		if err != nil {
			return nil, err
		}
		// if cpuPct < float64(cpuWatermark) {
		// continue
		// }
		name, err := proc.Name()
		if err != nil {
			return nil, err
		}

		status, err := proc.Status()
		if err != nil {
			return nil, err
		}

		memPct, err := proc.MemoryPercent()
		if err != nil {
			return nil, err
		}

		command, err := proc.Cmdline()
		if err != nil {
			return nil, err
		}

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			return nil, err
		}

		newProc := &Process{
			Name:          name,
			Pid:           int(proc.Pid),
			Status:        status[0],
			MemoryPercent: memPct,
			CpuPercent:    float32(cpuPct),
			Command:       command,
			RSS:           int(memInfo.RSS) / 1024,
			VMS:           int(memInfo.VMS) / 1024,
			Data:          int(memInfo.Data) / 1024,
			Stack:         int(memInfo.Stack) / 1024,
			Locked:        int(memInfo.Locked) / 1024,
			Swap:          int(memInfo.Swap) / 1024,
		}

		newProcs = append(newProcs, newProc)

		cpuTimes, _ := proc.Times()

		var _, _ = fmt.Printf("Process pid: %d, had cpu times %#v\n", proc.Pid, cpuTimes)
	}

	return newProcs, nil
}

func TestReadingProcesses(t *testing.T) {
	newProcs, err := ProcessesAboveWatermark(1)
	if err != nil {
		t.Fatalf("Unexpected error, %v", err)
	}

	procsByPid := make(map[int]Process, len(newProcs))

	for _, proc := range newProcs {
		procsByPid[proc.Pid] = *proc
	}

	oldProcs, err := psutilProcFromGopsUtil(1)
	if err != nil {
		t.Fatalf("Unexpected error, %v", err)
	}

	for _, oldProc := range oldProcs {
		newProc := procsByPid[oldProc.Pid]
		t.Logf("Old: %+v", *oldProc)
		t.Logf("New: %+v", newProc)
	}
}

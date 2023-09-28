// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package psutil

type LoadAvg struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

type MemoryStat struct {
	Total       int
	Free        int
	Used        int
	UsedPercent float32
}

type Memory struct {
	Virtual MemoryStat
	Swap    MemoryStat
	Buffers int
	Cached  int
}

type Process struct {
	Name          string
	Pid           int
	Status        string
	MemoryPercent float32
	CpuPercent    float32
	Command       string
	RSS           int
	VMS           int
	Data          int
	Stack         int
	Locked        int
	Swap          int
}

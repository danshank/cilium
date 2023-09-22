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

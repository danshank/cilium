// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package psutil

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Reads the average load statistics from /proc/loadavg
func Load() (*LoadAvg, error) {
	f, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}

	values := strings.Fields(string(f))

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return nil, err
	}

	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return nil, err
	}

	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return nil, err
	}

	return &LoadAvg{load1, load5, load15}, nil
}

func calculateMemory(total int, free int) *MemoryStat {
	used := total - free
	usedPct := float32(used) / float32(total) * 100.0
	return &MemoryStat{Total: total, Free: free, Used: used, UsedPercent: usedPct}
}

func MemoryInfo() (*Memory, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	fieldSplit := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	var totalVir, freeVir, totalSwp, freeSwp, buffers, cached int

	lines := bufio.NewScanner(f)
	lines.Split(bufio.ScanLines)
	for lines.Scan() {
		line := lines.Text()

		fields := strings.FieldsFunc(line, fieldSplit)

		switch fields[0] {
		case "MemTotal":
			totalVir, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}

		case "MemFree":
			freeVir, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}

		case "SwapTotal":
			totalSwp, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}

		case "SwapFree":
			freeSwp, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}

		case "Buffers":
			buffers, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}

		case "Cached":
			cached, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}
		}
	}

	memVir, memSwp := calculateMemory(totalVir, freeVir), calculateMemory(totalSwp, freeSwp)
	return &Memory{Virtual: *memVir, Swap: *memSwp, Buffers: buffers, Cached: cached}, nil
}

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package psutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var clkTck = uint64(100)

// Reads the average load statistics from /proc/loadavg
func Load() (*LoadAvg, error) {
	return loadFromFile("/proc/loadavg")
}

func loadFromFile(filename string) (*LoadAvg, error) {
	f, err := os.ReadFile(filename)
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

func memInfoFieldFieldSplit(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

func statusFileFieldSplit(c rune) bool {
	return unicode.IsSpace(c)
}

func MemoryInfo() (*Memory, error) {
	return memoryInfoFromFile("/proc/meminfo")
}

func memoryInfoFromFile(filename string) (*Memory, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	var totalVir, freeVir, totalSwp, freeSwp, buffers, cached, sReclaimable int

	lines := bufio.NewScanner(f)
	lines.Split(bufio.ScanLines)
	for lines.Scan() {
		line := lines.Text()

		fields := strings.FieldsFunc(line, memInfoFieldFieldSplit)

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

		case "SReclaimable":
			sReclaimable, err = strconv.Atoi(fields[1])
			if err != nil {
				return nil, err
			}
		}

	}

	memVir, memSwp := calculateMemory(totalVir, freeVir+buffers+cached+sReclaimable), calculateMemory(totalSwp, freeSwp)
	memVir.Free = freeVir
	return &Memory{Virtual: *memVir, Swap: *memSwp, Buffers: buffers, Cached: cached}, nil
}

func ProcessesAboveWatermark(cpuWatermark float64) ([]*Process, error) {
	file, _ := os.ReadFile("/proc/uptime")
	contents := strings.Fields(string(file))
	var _, _ = fmt.Printf("Contents are: %v\n", contents)

	uptimeSecs, _ := strconv.ParseFloat(contents[0], 64)
	uptime := uint64(uptimeSecs * float64(clkTck))
	var _, _ = fmt.Printf("Uptime seconds is: %f\n", uptimeSecs)
	var _, _ = fmt.Printf("Uptime is: %d\n", uptime)

	mem, err := MemoryInfo()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	processes := make([]*Process, 0)

	for _, f := range files {
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}

		statFile := fmt.Sprintf("/proc/%d/stat", pid)
		stats, err := os.ReadFile(statFile)
		statRaw := string(stats)
		parenEnd := bytes.IndexByte(stats, ')')
		statFields := strings.Fields(string(stats[parenEnd+1:]))

		var _, _ = fmt.Printf("For pid: %d\n", pid)
		var _, _ = fmt.Printf("Raw stat file is: %v\n", statRaw)
		var _, _ = fmt.Printf("Contents of statFields is: %v\n", statFields)
		utime, _ := strconv.ParseUint(statFields[11], 10, 64)
		var _, _ = fmt.Printf("Found user time to be: %d\n", utime)
		stime, _ := strconv.ParseUint(statFields[12], 10, 64)
		var _, _ = fmt.Printf("Found scheduled time to be: %d\n", stime)
		blkiotime, _ := strconv.ParseUint(statFields[39], 10, 64)
		var _, _ = fmt.Printf("Found block time to be: %d\n", blkiotime)

		active := utime + stime + blkiotime

		start, _ := strconv.ParseUint(statFields[19], 10, 64)

		total := uptime - start
		var _, _ = fmt.Printf("Total time is: %d\n", total)

		pctCPU := float64(active) / float64(total) * 100.0

		// if pctCPU < cpuWatermark {
		// continue
		// }

		commandFile := fmt.Sprintf("/proc/%d/cmdline", pid)
		command, err := os.ReadFile(commandFile)
		if err != nil {
			return nil, err
		}

		statusFile := fmt.Sprintf("/proc/%d/status", pid)
		status, err := os.Open(statusFile)
		if err != nil {
			return nil, err
		}
		defer status.Close()

		lines := bufio.NewScanner(status)
		lines.Split(bufio.ScanLines)
		var name, state string
		var rss, vsz, data, stack, locked, swap int
		for lines.Scan() {
			line := lines.Text()
			fields := strings.FieldsFunc(line, statusFileFieldSplit)

			switch fields[0] {
			case "Name":
				name = fields[1]
			case "State":
				state = fields[2]
			case "VmRSS":
				rss, err = strconv.Atoi(fields[1])
				if err != nil {
					return nil, err
				}
			case "VmSize":
				vsz, err = strconv.Atoi(fields[1])
				if err != nil {
					return nil, err
				}
			case "VmData":
				data, err = strconv.Atoi(fields[1])
				if err != nil {
					return nil, err
				}
			case "VmStk":
				stack, err = strconv.Atoi(fields[1])
				if err != nil {
					return nil, err
				}
			case "VmLck":
				locked, err = strconv.Atoi(fields[1])
				if err != nil {
					return nil, err
				}
			case "VmSwp":
				swap, err = strconv.Atoi(fields[1])
				if err != nil {
					return nil, err
				}
			}
		}

		memPct := float32(rss) / float32(mem.Virtual.Total) * 100.0

		process := &Process{
			Name:          name,
			Pid:           pid,
			Status:        state,
			MemoryPercent: memPct,
			CpuPercent:    float32(pctCPU),
			Command:       string(command),
			RSS:           rss,
			VMS:           vsz,
			Data:          data,
			Stack:         stack,
			Locked:        locked,
			Swap:          swap,
		}

		processes = append(processes, process)
	}

	return processes, nil
}

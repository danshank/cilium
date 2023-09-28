// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package loadinfo

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cilium/cilium/pkg/inctimer"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/psutil"
)

const (
	// backgroundInterval is the interval in which system load information is logged
	backgroundInterval = 5 * time.Second

	// cpuWatermark is the minimum percentage of CPU to have a process
	// listed in the log
	cpuWatermark = 1.0
)

var log = logging.DefaultLogger.WithField(logfields.LogSubsys, "loadinfo")

// LogFunc is the function to used to log the system load
type LogFunc func(format string, args ...interface{})

func toMB(total int) int {
	return total / 1024
}

// LogCurrentSystemLoad logs the current system load and lists all processes
// consuming more than cpuWatermark of the CPU
func LogCurrentSystemLoad(logFunc LogFunc) {
	loadInfo, err := psutil.Load()
	if err == nil {
		logFunc("Load 1-min: %.2f 5-min: %.2f 15min: %.2f",
			loadInfo.Load1, loadInfo.Load5, loadInfo.Load15)
	}

	memoryInfo, err := psutil.MemoryInfo()
	if err == nil && memoryInfo != nil {
		logFunc("Memory: Total: %d Used: %d (%.2f%%) Free: %d Buffers: %d Cached: %d",
			toMB(memoryInfo.Virtual.Total), toMB(memoryInfo.Virtual.Used), memoryInfo.Virtual.UsedPercent, toMB(memoryInfo.Virtual.Free), toMB(memoryInfo.Buffers), toMB(memoryInfo.Cached))
		logFunc("Swap: Total: %d Used: %d (%.2f%%) Free: %d",
			toMB(memoryInfo.Swap.Total), toMB(memoryInfo.Swap.Used), memoryInfo.Swap.UsedPercent, toMB(memoryInfo.Swap.Free))
	}

	procs, err := psutil.ProcessesAboveWatermark(cpuWatermark)
	if err == nil {
		for _, p := range procs {
			memExt := fmt.Sprintf("RSS: %d VMS: %d Data: %d Stack: %d Locked: %d Swap: %d",
				toMB(p.RSS), toMB(p.VMS), toMB(p.Data),
				toMB(p.Stack), toMB(p.Locked), toMB(p.Swap))

			logFunc("NAME %s STATUS %s PID %d CPU: %.2f%% MEM: %.2f%% CMDLINE: %s MEM-EXT: %s",
				p.Name, p.Status, p.Pid, p.CpuPercent, p.MemoryPercent, p.Command, memExt)
		}
	}
}

// LogPeriodicSystemLoad logs the system load in the interval specified until
// the given ctx is canceled.
func LogPeriodicSystemLoad(ctx context.Context, logFunc LogFunc, interval time.Duration) {
	go func() {
		LogCurrentSystemLoad(logFunc)

		timer, timerDone := inctimer.New()
		defer timerDone()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.After(interval):
				LogCurrentSystemLoad(logFunc)
			}
		}
	}()
}

// StartBackgroundLogger starts background logging
func StartBackgroundLogger() {
	LogPeriodicSystemLoad(context.Background(), log.WithFields(logrus.Fields{"type": "background"}).Debugf, backgroundInterval)
}

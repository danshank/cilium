// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package psutil

import (
	"os"
	"strconv"
	"strings"
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

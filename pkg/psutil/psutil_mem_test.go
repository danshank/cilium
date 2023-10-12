// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package psutil

import (
	"reflect"
	"testing"
)

func TestReadingMem(t *testing.T) {
	testFile := "testdata/meminfo"

	expected := &Memory{
		Virtual: MemoryStat{
			Total:       100000,
			Free:        20000,
			Used:        64900,
			UsedPercent: float32(64.9),
		},
		Swap: MemoryStat{
			Total:       50000,
			Free:        40000,
			Used:        10000,
			UsedPercent: float32(20),
		},
		Buffers: 10000,
		Cached:  5000,
	}

	actual, err := memoryInfoFromFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error, %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected: %+v, but got: %+v", expected, actual)
	}
}

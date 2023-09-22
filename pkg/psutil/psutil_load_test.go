// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium
package psutil

import (
	"reflect"
	"testing"
)

func TestReadingLoad(t *testing.T) {
	testFile := "testdata/loadavg"

	expected := &LoadAvg{
		Load1:  float64(0.2),
		Load5:  float64(0.28),
		Load15: float64(0.27),
	}
	actual, err := loadFromFile(testFile)
	if err != nil {
		t.Fatalf("Unexpected error, %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected: %+v, but got: %+v", expected, actual)
	}
}

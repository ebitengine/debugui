// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package caller_test

import (
	"testing"

	"github.com/ebitengine/debugui/internal/caller"
)

func TestMultipleCallersInForLoop(t *testing.T) {
	var pc uintptr
	for range 10 {
		pc2 := caller.Caller()
		if pc2 == 0 {
			t.Errorf("Caller() returned 0")
			continue
		}
		if pc == 0 {
			pc = pc2
			continue
		}
		if pc != pc2 {
			t.Errorf("Caller() returned different values: %d and %d", pc, pc2)
		}
	}
}

func TestMultipleCallersOnOneLine(t *testing.T) {
	a, b := caller.Caller(), caller.Caller()
	if a == b {
		t.Errorf("Caller() returned the same value twice: %d", a)
	}
}

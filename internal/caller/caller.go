// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package caller

import (
	"runtime"
)

// Caller returns a program counter of the caller.
func Caller() uintptr {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return 0
	}
	return pc
}

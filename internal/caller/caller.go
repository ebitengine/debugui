// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package caller

import (
	"go/build"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	debugUIFileDir     string
	debugUIFileDirOnce sync.Once
)

// Caller returns a program counter of the caller outside of this module.
func Caller() uintptr {
	debugUIFileDirOnce.Do(func() {
		pkg, err := build.Default.Import("github.com/ebitengine/debugui", "", build.FindOnly)
		if err != nil {
			return
		}
		debugUIFileDir = filepath.ToSlash(pkg.Dir)
	})

	if debugUIFileDir == "" {
		return 0
	}

	var debugUIPackageReached bool
	for i := 0; ; i++ {
		pc, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// The file should be with a slash, but just in case, convert it.
		file = filepath.ToSlash(file)

		// TODO: This is heuristic. Use a file set.
		if strings.HasSuffix(path.Base(file), "_test.go") && path.Base(file) != "export_test.go" {
			return pc
		}

		if !debugUIPackageReached {
			if path.Dir(file) == debugUIFileDir {
				debugUIPackageReached = true
			}
			continue
		}

		if path.Dir(file) == debugUIFileDir {
			continue
		}
		if strings.HasPrefix(path.Dir(file), debugUIFileDir+"/") && !strings.HasPrefix(path.Dir(file), debugUIFileDir+"/example") {
			continue
		}
		return pc
	}

	return 0
}

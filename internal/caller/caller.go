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

// FirstCaller returns the file and line number of the first caller outside of DebugUI module.
func FirstCaller() (file string, line int) {
	debugUIFileDirOnce.Do(func() {
		pkg, err := build.Default.Import("github.com/ebitengine/debugui", "", build.FindOnly)
		if err != nil {
			return
		}
		debugUIFileDir = filepath.ToSlash(pkg.Dir)
	})

	if debugUIFileDir == "" {
		return "", 0
	}

	var debugUIPackageReached bool
	for i := 0; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// The file should be with a slash, but just in case, convert it.
		file = filepath.ToSlash(file)

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
		return file, line
	}

	return "", 0
}

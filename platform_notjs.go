// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !js

package debugui

import "runtime"

// isApplePlatform reports whether the current platform follows Apple's keyboard
// conventions, where Command (Meta) is the primary shortcut modifier.
func isApplePlatform() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "ios"
}

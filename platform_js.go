// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui

import (
	"strings"
	"syscall/js"
)

// applePlatform caches whether the page runs on an Apple platform. On wasm
// runtime.GOOS is always "js", so the underlying OS is detected from the browser
// instead. The result is fixed for the page's lifetime.
var applePlatform = detectApplePlatform()

func detectApplePlatform() bool {
	nav := js.Global().Get("navigator")
	if !nav.Truthy() {
		return false
	}
	if p := nav.Get("platform"); p.Truthy() {
		s := p.String()
		return strings.HasPrefix(s, "Mac") || s == "iPhone" || s == "iPad" || s == "iPod"
	}
	ua := nav.Get("userAgent").String()
	return strings.Contains(ua, "Macintosh") || strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod")
}

// isApplePlatform reports whether the current platform follows Apple's keyboard
// conventions, where Command (Meta) is the primary shortcut modifier.
func isApplePlatform() bool {
	return applePlatform
}

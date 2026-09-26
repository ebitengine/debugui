// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ebitengine/debugui"
)

func TestLogicalClockTPSChanges(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var clock debugui.LogicalClock
		check := func(tick int64, tps int, want time.Duration) {
			t.Helper()
			if got := clock.Read(tick, tps); got != want {
				t.Errorf("Read(%d, %d) = %v, want %v", tick, tps, got, want)
			}
		}
		check(0, 60, 0)
		time.Sleep(10 * time.Second)
		// Fixed-TPS time advances with ticks, regardless of elapsed wall time.
		check(60, 120, time.Second)
		check(120, ebiten.SyncWithFPS, 1500*time.Millisecond)
		time.Sleep(100 * time.Millisecond)
		check(121, ebiten.SyncWithFPS, 1600*time.Millisecond)
		time.Sleep(150 * time.Millisecond)
		check(122, 60, 1750*time.Millisecond)
		time.Sleep(time.Second)
		check(182, 60, 2750*time.Millisecond)
	})
}

func TestLogicalClockFractionalTicks(t *testing.T) {
	var clock debugui.LogicalClock
	clock.Read(0, 60)
	for tick := int64(1); tick <= 60; tick++ {
		if got, want := clock.Read(tick, 60), time.Duration(tick)*time.Second/60; got != want {
			t.Errorf("tick %d: time = %v, want %v", tick, got, want)
		}
	}
}

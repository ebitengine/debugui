// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/ebitengine/debugui"
)

func TestInputRepeat(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var repeat debugui.InputRepeat
		start := time.Now()
		check := func(pressed bool, elapsed time.Duration, want bool) {
			t.Helper()
			time.Sleep(elapsed - time.Since(start))
			if got := repeat.Update(pressed, time.Now()); got != want {
				t.Errorf("Update(%v, %v) = %v, want %v", pressed, elapsed, got, want)
			}
		}

		check(false, 0, false)
		check(true, 0, true)
		// Additional updates do not shorten the initial delay.
		for ms := 1; ms < 400; ms++ {
			check(true, time.Duration(ms)*time.Millisecond, false)
		}
		check(true, 400*time.Millisecond, true)
		check(true, 466*time.Millisecond, false)
		check(true, 467*time.Millisecond, true)

		// A stalled update produces one repeat without a catch-up burst.
		check(true, 2*time.Second, true)
		check(true, 2*time.Second, false)

		// Release clears the repeat state, and a new press starts a fresh delay.
		check(false, 3*time.Second, false)
		check(false, 4*time.Second, false)
		check(true, 5*time.Second, true)
		check(true, 5399*time.Millisecond, false)
		check(true, 5400*time.Millisecond, true)
	})
}

func TestInputRepeatTicks(t *testing.T) {
	for _, tps := range []int{60, 120} {
		var repeat debugui.InputRepeat
		if repeat.UpdateTick(false, tps) {
			t.Error("unpressed input repeats")
		}
		if !repeat.UpdateTick(true, tps) {
			t.Error("initial press does not repeat")
		}
		delay := tps * 2 / 5
		interval := tps / 15
		for tick := 1; tick <= delay+2*interval; tick++ {
			want := tick >= delay && (tick-delay)%interval == 0
			if got := repeat.UpdateTick(true, tps); got != want {
				t.Errorf("tick %d at %d TPS: repeated = %v, want %v", tick, tps, got, want)
			}
		}
	}
}

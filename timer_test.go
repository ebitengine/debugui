// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/ebitengine/debugui"
)

func TestTimerFired(t *testing.T) {
	for _, interval := range []time.Duration{0, 100 * time.Millisecond} {
		t.Run(interval.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var timer debugui.Timer
				timer.Start(400*time.Millisecond, interval)
				time.Sleep(400 * time.Millisecond)
				timer.Update()
				for range 2 {
					if !timer.Fired() {
						t.Error("firing must remain observable throughout the update")
					}
				}
				if got, want := timer.Expired(), interval == 0; got != want {
					t.Errorf("Expired() = %v, want %v", got, want)
				}
				time.Sleep(time.Millisecond)
				timer.Update()
				if timer.Fired() {
					t.Error("firing was not cleared by the next update")
				}
				time.Sleep(time.Second)
				timer.Update()
				if got, want := timer.Fired(), interval > 0; got != want {
					t.Errorf("Fired() after a stall = %v, want %v", got, want)
				}
			})
		})
	}
}

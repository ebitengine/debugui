// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui

import "time"

type timer struct {
	running   bool
	last      time.Duration
	remaining time.Duration
	interval  time.Duration
	didFire   bool
}

func (t *timer) update(now time.Duration) {
	t.didFire = false
	if !t.running {
		return
	}
	wasExpired := t.expired()
	t.remaining -= now - t.last
	t.last = now
	if wasExpired || !t.expired() {
		return
	}
	t.didFire = true
	if t.interval > 0 {
		// Skip missed repeats while preserving the repeat cadence.
		t.remaining += (-t.remaining/t.interval + 1) * t.interval
	}
}

func (t *timer) start(now, duration, interval time.Duration) {
	t.running = true
	t.last = now
	t.remaining = duration
	t.interval = interval
	t.didFire = false
}

func (t *timer) stop() {
	*t = timer{}
}

func (t *timer) active() bool {
	return t.running
}

func (t *timer) expired() bool {
	return t.running && t.remaining <= 0
}

func (t *timer) fired() bool {
	return t.didFire
}

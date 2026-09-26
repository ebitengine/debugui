// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var theTimerClock logicalClock

func currentTimerTime() time.Duration {
	tick := ebiten.Tick()
	if theTimerClock.initialized && theTimerClock.tick == tick {
		return theTimerClock.elapsed
	}
	return theTimerClock.read(tick, ebiten.TPS(), time.Now())
}

type logicalClock struct {
	initialized bool
	tick        int64
	tps         int
	wall        time.Time
	elapsed     time.Duration
	fraction    int64
}

func (c *logicalClock) read(tick int64, tps int, wall time.Time) time.Duration {
	if c.initialized {
		if c.tps > 0 {
			// Carry fractional nanoseconds so fixed-rate ticks do not accumulate drift.
			numerator := (tick-c.tick)*int64(time.Second) + c.fraction
			c.elapsed += time.Duration(numerator / int64(c.tps))
			c.fraction = numerator % int64(c.tps)
		} else {
			c.elapsed += wall.Sub(c.wall)
		}
	}
	if c.tps != tps {
		c.fraction = 0
	}
	c.initialized = true
	c.tick = tick
	c.tps = tps
	c.wall = wall
	return c.elapsed
}

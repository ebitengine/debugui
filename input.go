// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type pointing struct {
	justPressedTouchIDs []ebiten.TouchID
	touchIDs            []ebiten.TouchID
	hasPrimaryTouchID   bool
	primaryTouchID      ebiten.TouchID
	repeat              inputRepeat
}

func (p *pointing) update(now time.Duration) {
	p.justPressedTouchIDs = inpututil.AppendJustPressedTouchIDs(p.justPressedTouchIDs[:0])
	p.touchIDs = ebiten.AppendTouchIDs(p.touchIDs[:0])

	if len(p.touchIDs) == 0 {
		p.hasPrimaryTouchID = false
		p.primaryTouchID = 0
	} else if !p.hasPrimaryTouchID {
		p.hasPrimaryTouchID = true
		p.primaryTouchID = p.touchIDs[0]
	}

	p.repeat.update(p.pressed(), now)
}

func (p *pointing) isTouchActive() bool {
	if !p.hasPrimaryTouchID {
		return false
	}
	return slices.Contains(p.touchIDs, p.primaryTouchID)
}

func (p *pointing) position() image.Point {
	if p.isTouchActive() {
		return image.Pt(ebiten.TouchPosition(p.primaryTouchID))
	}
	return image.Pt(ebiten.CursorPosition())
}

func (p *pointing) pressed() bool {
	if p.isTouchActive() {
		return true
	}
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (p *pointing) justPressed() bool {
	if p.isTouchActive() {
		return slices.Contains(p.justPressedTouchIDs, p.primaryTouchID)
	}
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

func (p *pointing) repeated() bool {
	return p.repeat.repeated
}

func (c *Context) keyRepeated(key ebiten.Key) bool {
	return c.keyRepeats[key].repeated
}

func (c *Context) updateInput() {
	c.now = currentTimerTime()
	c.pointing.update(c.now)
	for _, key := range [...]ebiten.Key{
		ebiten.KeyLeft, ebiten.KeyRight, ebiten.KeyUp, ebiten.KeyDown,
		ebiten.KeyHome, ebiten.KeyEnd, ebiten.KeyBackspace, ebiten.KeyDelete,
	} {
		c.keyRepeats[key].update(ebiten.IsKeyPressed(key), c.now)
	}
}

type inputRepeat struct {
	timer    timer
	repeated bool
}

func (r *inputRepeat) update(pressed bool, now time.Duration) {
	r.timer.update(now)
	r.repeated = false
	if !pressed {
		r.timer.stop()
		return
	}
	if !r.timer.active() {
		r.repeated = true
		r.timer.start(now, 400*time.Millisecond, time.Second/15)
		return
	}
	r.repeated = r.timer.fired()
}

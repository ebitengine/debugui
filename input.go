// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (c *Context) inputScroll(x, y int) {
	c.scrollDelta.X += x
	c.scrollDelta.Y += y
}

func keyToInt(key ebiten.Key) int {
	switch key {
	case ebiten.KeyShift:
		return keyShift
	case ebiten.KeyControl:
		return keyControl
	case ebiten.KeyAlt:
		return keyAlt
	case ebiten.KeyBackspace:
		return keyBackspace
	case ebiten.KeyEnter:
		return keyReturn
	}
	return 0
}

func (c *Context) inputKeyDown(key ebiten.Key) {
	c.keyDown |= keyToInt(key)
}

func (c *Context) inputKeyUp(key ebiten.Key) {
	c.keyDown &= ^keyToInt(key)
}

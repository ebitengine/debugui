// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type pointing struct {
	justPressedTouchIDs []ebiten.TouchID
	touchIDs            []ebiten.TouchID
	hasPrimaryTouchID   bool
	primaryTouchID      ebiten.TouchID
}

func (p *pointing) update() {
	p.justPressedTouchIDs = inpututil.AppendJustPressedTouchIDs(p.justPressedTouchIDs[:0])
	p.touchIDs = ebiten.AppendTouchIDs(p.touchIDs[:0])

	if len(p.touchIDs) == 0 {
		p.hasPrimaryTouchID = false
		p.primaryTouchID = 0
	} else if !p.hasPrimaryTouchID {
		p.hasPrimaryTouchID = true
		p.primaryTouchID = p.touchIDs[0]
	}
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

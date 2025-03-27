// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type pointing struct {
}

func (p *pointing) update() {
	// TODO: Implement this for touches.
}

func (p *pointing) position() image.Point {
	return image.Pt(ebiten.CursorPosition())
}

func (p *pointing) pressed() bool {
	return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (p *pointing) justPressed() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import "github.com/hajimehoshi/ebiten/v2"

// DebugUI is a debug UI.
//
// The zero value for DebugUI is ready to use.
type DebugUI struct {
	ctx Context
}

// Update updates the debug UI.
//
// Update returns true if the debug UI is capturing input, e.g. when a widget has focus.
// Otherwise, Update returns false.
//
// Update should be called once in the game's Update function.
func (d *DebugUI) Update(f func(ctx *Context) error) (bool, error) {
	captured, err := d.ctx.update(f)
	if err != nil {
		return false, err
	}
	return captured, nil
}

// Draw draws the debug UI.
//
// Draw should be called once in the game's Draw function.
func (d *DebugUI) Draw(screen *ebiten.Image) {
	d.ctx.draw(screen)
	d.ctx.screenWidth, d.ctx.screenHeight = screen.Bounds().Dx(), screen.Bounds().Dy()
}

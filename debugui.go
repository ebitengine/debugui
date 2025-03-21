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
// Update should be called once in the game's Update function.
func (d *DebugUI) Update(f func(ctx *Context) error) error {
	return d.ctx.update(f)
}

// Draw draws the debug UI.
//
// Draw should be called once in the game's Draw function.
func (d *DebugUI) Draw(screen *ebiten.Image) {
	d.ctx.draw(screen)
}

// IsCapturingInput reports whether the debug UI is capturing input, e.g. when a control has focus.
func (d *DebugUI) IsCapturingInput() bool {
	return d.ctx.hoverRoot != nil || d.ctx.focus != 0
}

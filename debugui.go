// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import "github.com/hajimehoshi/ebiten/v2"

type DebugUI struct {
	ctx *Context
}

func New() *DebugUI {
	return &DebugUI{
		ctx: &Context{
			scale: 1,
			style: &defaultStyle,
		},
	}
}

func (d *DebugUI) Update(f func(ctx *Context)) {
	d.ctx.update(f)
}

func (d *DebugUI) Draw(screen *ebiten.Image) {
	d.ctx.draw(screen)
}

func (d *DebugUI) IsCapturingInput() bool {
	return d.ctx.hoverRoot != nil || d.ctx.focus != 0
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"image/color"
	"iter"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	commandJump = 1 + iota
	commandClip
	commandRect
	commandText
	commandIcon
	commandDraw
)

type baseCommand struct {
	typ int
}

type jumpCommand struct {
	dstIdx int
}

type clipCommand struct {
	rect image.Rectangle
}

type rectCommand struct {
	rect  image.Rectangle
	color color.Color
}

type textCommand struct {
	pos   image.Point
	color color.Color
	str   string
}

type iconCommand struct {
	rect  image.Rectangle
	icon  icon
	color color.Color
}

type drawCommand struct {
	f func(screen *ebiten.Image)
}

type command struct {
	typ  int
	idx  int
	base baseCommand // type 0 (TODO)
	jump jumpCommand // type 1
	clip clipCommand // type 2
	rect rectCommand // type 3
	text textCommand // type 4
	icon iconCommand // type 5
	draw drawCommand // type 6
}

// appendCommand adds a new command with type cmd_type to the command list.
func (c *Context) appendCommand(cmd_type int) *command {
	cmd := command{
		typ: cmd_type,
	}
	cmd.base.typ = cmd_type
	cmd.idx = len(c.commandList)
	c.commandList = append(c.commandList, &cmd)
	return &cmd
}

// appendJumpCommand appends a new jump command to the command list.
// dstIdx is set to -1. This can be updated later.
func (c *Context) appendJumpCommand() int {
	cmd := c.appendCommand(commandJump)
	cmd.jump.dstIdx = -1
	return len(c.commandList) - 1
}

func (c *Context) commands() iter.Seq[*command] {
	return func(yield func(command *command) bool) {
		if len(c.commandList) == 0 {
			return
		}

		cmd := c.commandList[0]
		for cmd.idx < len(c.commandList) {
			if cmd.typ != commandJump {
				if !yield(cmd) {
					return
				}
				cmd = c.commandList[cmd.idx+1]
				continue
			}
			idx := cmd.jump.dstIdx
			if idx > len(c.commandList)-1 {
				return
			}
			cmd = c.commandList[idx]
		}
	}
}

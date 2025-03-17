// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
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

func (c *Context) nextCommand(cmd **command) bool {
	if len(c.commandList) == 0 {
		return false
	}
	if *cmd == nil {
		*cmd = c.commandList[0]
	} else {
		*cmd = c.commandList[(*cmd).idx+1]
	}

	for (*cmd).idx < len(c.commandList) {
		if (*cmd).typ != commandJump {
			return true
		}
		idx := (*cmd).jump.dstIdx
		if idx > len(c.commandList)-1 {
			break
		}
		*cmd = c.commandList[idx]
	}
	return false
}

// pushJump pushes a new jump command to the command list.
func (c *Context) pushJump(dstIdx int) int {
	cmd := c.appendCommand(commandJump)
	cmd.jump.dstIdx = dstIdx
	return len(c.commandList) - 1
}

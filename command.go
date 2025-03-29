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
	commandClip = 1 + iota
	commandRect
	commandText
	commandIcon
	commandDraw
)

type baseCommand struct {
	typ int
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
	base baseCommand // type 0 (TODO)
	clip clipCommand // type 1
	rect rectCommand // type 2
	text textCommand // type 3
	icon iconCommand // type 4
	draw drawCommand // type 5
}

// appendCommand adds a new command with type cmd_type to the command list.
func (c *Context) appendCommand(cmdType int) *command {
	cmd := command{
		typ: cmdType,
	}
	cmd.base.typ = cmdType
	cnt := c.currentRootContainer()
	cnt.commandList = append(cnt.commandList, &cmd)
	return &cmd
}

func (c *Context) commands() iter.Seq[*command] {
	return func(yield func(command *command) bool) {
		for _, cnt := range c.rootContainers {
			for _, cmd := range cnt.commandList {
				if !yield(cmd) {
					return
				}
			}
		}
		/*if len(c.commandList) == 0 {
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
		}*/
	}
}

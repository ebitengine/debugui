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
	clip clipCommand
	rect rectCommand
	text textCommand
	icon iconCommand
	draw drawCommand
}

// appendCommand adds a new command with type cmdType to the command list.
func (c *Context) appendCommand(cmdType int) *command {
	cmd := command{
		typ: cmdType,
	}
	cnt := c.currentRootContainer()
	cnt.commandList = append(cnt.commandList, &cmd)
	return &cmd
}

// commands returns a sequence of commands from all root containers.
func (c *Context) commands() iter.Seq[*command] {
	return func(yield func(command *command) bool) {
		for _, cnt := range c.rootContainers {
			for _, cmd := range cnt.commandList {
				if !yield(cmd) {
					return
				}
			}
		}
	}
}

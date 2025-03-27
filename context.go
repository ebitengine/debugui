// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"errors"
	"image"
	"slices"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

type Context struct {
	pointing pointing

	scaleMinus1   int
	hover         WidgetID
	focus         WidgetID
	currentID     WidgetID
	lastZIndex    int
	keepFocus     bool
	hoverRoot     *container
	nextHoverRoot *container
	scrollTarget  *container
	numberEditBuf string
	numberEdit    WidgetID

	idStack        []WidgetID
	commandList    []*command
	rootList       []*container
	containerStack []*container
	usedContainers map[WidgetID]struct{}
	clipStack      []image.Rectangle
	layoutStack    []layout

	idToContainer map[WidgetID]*container

	lastPointingPos image.Point

	screenWidth  int
	screenHeight int

	err error
}

func (c *Context) update(f func(ctx *Context) error) (err error) {
	if c.err != nil {
		return c.err
	}

	c.pointing.update()

	c.begin()
	defer func() {
		if err2 := c.end(); err2 != nil && err == nil {
			err = err2
		}
	}()

	if err := f(c); err != nil {
		return err
	}
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *Context) begin() {
	c.commandList = slices.Delete(c.commandList, 0, len(c.commandList))
	c.rootList = slices.Delete(c.rootList, 0, len(c.rootList))
	c.scrollTarget = nil
	c.hoverRoot = c.nextHoverRoot
	c.nextHoverRoot = nil
	c.currentID = emptyWidgetID
}

func (c *Context) end() error {
	// check stacks
	if len(c.idStack) > 0 {
		return errors.New("debugui: id stack must be empty")
	}
	if len(c.containerStack) > 0 {
		return errors.New("debugui: container stack must be empty")
	}
	if len(c.clipStack) > 0 {
		return errors.New("debugui: clip stack must be empty")
	}
	if len(c.layoutStack) > 0 {
		return errors.New("debugui: layout stack must be empty")
	}

	// handle scroll input
	if c.scrollTarget != nil {
		wx, wy := ebiten.Wheel()
		c.scrollTarget.layout.ScrollOffset.X += int(wx * -30)
		c.scrollTarget.layout.ScrollOffset.Y += int(wy * -30)
	}

	// unset focus if focus id was not touched this frame
	if !c.keepFocus {
		c.focus = emptyWidgetID
	}
	c.keepFocus = false

	// bring hover root to front if the pointing device was pressed
	if c.pointing.justPressed() && c.nextHoverRoot != nil &&
		c.nextHoverRoot.zIndex < c.lastZIndex &&
		c.nextHoverRoot.zIndex >= 0 {
		c.bringToFront(c.nextHoverRoot)
	}

	// reset input state
	c.lastPointingPos = c.pointingPosition()

	// sort root containers by zindex
	sort.SliceStable(c.rootList, func(i, j int) bool {
		return c.rootList[i].zIndex < c.rootList[j].zIndex
	})

	// set root container jump commands
	for i := range c.rootList {
		cnt := c.rootList[i]
		// if this is the first container then make the first command jump to it.
		// otherwise set the previous container's tail to jump to this one
		if i == 0 {
			cmd := c.commandList[0]
			if cmd.typ != commandJump {
				panic("debugui: expected jump command")
			}
			cmd.jump.dstIdx = cnt.headIdx + 1
			if cnt.headIdx >= len(c.commandList) {
				panic("debugui: invalid head index")
			}
		} else {
			prev := c.rootList[i-1]
			c.commandList[prev.tailIdx].jump.dstIdx = cnt.headIdx + 1
		}
		// make the last container's tail jump to the end of command list
		if i == len(c.rootList)-1 {
			c.commandList[cnt.tailIdx].jump.dstIdx = len(c.commandList)
		}
	}

	for id := range c.idToContainer {
		if _, ok := c.usedContainers[id]; !ok {
			delete(c.idToContainer, id)
		}
	}
	clear(c.usedContainers)

	return nil
}

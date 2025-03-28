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
	rootContainers []*container
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
	for _, cnt := range c.rootContainers {
		cnt.commandList = slices.Delete(cnt.commandList, 0, len(cnt.commandList))
	}
	c.rootContainers = slices.Delete(c.rootContainers, 0, len(c.rootContainers))
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
	sort.SliceStable(c.rootContainers, func(i, j int) bool {
		return c.rootContainers[i].zIndex < c.rootContainers[j].zIndex
	})

	for id := range c.idToContainer {
		if _, ok := c.usedContainers[id]; !ok {
			delete(c.idToContainer, id)
		}
	}
	clear(c.usedContainers)

	return nil
}

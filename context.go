// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"errors"
	"image"
	"maps"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

func clamp[T int | float64](x, a, b T) T {
	return min(b, max(a, x))
}

type Context struct {
	pointing pointing

	scaleMinus1   int
	hover         WidgetID
	focus         WidgetID
	currentID     WidgetID
	keepFocus     bool
	hoverRoot     *container
	nextHoverRoot *container
	scrollTarget  *container
	numberEditBuf string
	numberEdit    WidgetID

	idStack []WidgetID

	// idToContainer maps widget IDs to containers.
	// Only unused containers are removed from this map at the end of Update.
	idToContainer map[WidgetID]*container

	// rootContainers is a list of root containers.
	// rootContainers contains only root containers. For example, a panel is not contained.
	//
	// The order represents the z-order of the containers.
	// Only unused containers are removed from this list at the end of Update.
	rootContainers []*container

	containerStack []*container

	clipStack   []image.Rectangle
	layoutStack []layout

	lastPointingPos image.Point

	screenWidth  int
	screenHeight int

	err error
}

func (c *Context) wrapError(f func() error) {
	if c.err != nil {
		return
	}
	c.err = f()
}

func (c *Context) update(f func(ctx *Context) error) (err error) {
	if c.err != nil {
		return c.err
	}

	c.pointing.update()

	c.beginUpdate()
	defer func() {
		if err2 := c.endUpdate(); err2 != nil && err == nil {
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

func (c *Context) beginUpdate() {
	for _, cnt := range c.rootContainers {
		cnt.commandList = slices.Delete(cnt.commandList, 0, len(cnt.commandList))
	}
	c.scrollTarget = nil
	c.hoverRoot = c.nextHoverRoot
	c.nextHoverRoot = nil
	c.currentID = emptyWidgetID
}

func (c *Context) endUpdate() error {
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

	// Bring hover root to front if the pointing device was pressed
	if c.pointing.justPressed() && c.nextHoverRoot != nil {
		c.bringToFront(c.nextHoverRoot)
	}

	// reset input state
	c.lastPointingPos = c.pointingPosition()

	// Remove unused containers.
	c.rootContainers = slices.DeleteFunc(c.rootContainers, func(cnt *container) bool {
		return !cnt.used
	})
	maps.DeleteFunc(c.idToContainer, func(id WidgetID, cnt *container) bool {
		return !cnt.used
	})
	for _, cnt := range c.idToContainer {
		cnt.used = false
	}

	return nil
}

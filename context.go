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

// Context is the main context for the debug UI.
type Context struct {
	pointing pointing

	scaleMinus1   int
	hover         widgetID
	focus         widgetID
	currentID     widgetID
	keepFocus     bool
	scrollTarget  *container
	numberEditBuf string
	numberEdit    widgetID

	idStack widgetID

	// idToContainer maps widget IDs to containers.
	//
	// Unused containers are removed from this map at the end of Update.
	idToContainer map[widgetID]*container

	// rootContainers is a list of root containers.
	// rootContainers contains only root containers. For example, a panel is not contained.
	//
	// The order represents the z-order of the containers.
	//
	// Unused containers are removed from this list at the end of Update.
	rootContainers []*container

	containerStack []*container

	clipStack   []image.Rectangle
	layoutStack []layout

	lastPointingPos image.Point

	screenWidth  int
	screenHeight int

	err error
}

func (c *Context) wrapEventHandlerAndError(f func() (EventHandler, error)) EventHandler {
	if c.err != nil {
		return &nullEventHandler{}
	}
	e, err := f()
	if err != nil {
		c.err = err
		return &nullEventHandler{}
	}
	if e == nil {
		return &nullEventHandler{}
	}
	return e
}

func (c *Context) update(f func(ctx *Context) error) (inputCapturingState InputCapturingState, err error) {
	if c.err != nil {
		return 0, c.err
	}

	c.pointing.update()

	c.beginUpdate()
	defer func() {
		if err2 := c.endUpdate(); err2 != nil && err == nil {
			err = err2
		}
	}()

	if err := f(c); err != nil {
		return 0, err
	}
	if c.err != nil {
		return 0, c.err
	}

	// Check whether the cursor is on any of the root containers.
	pt := c.pointingPosition()
	for _, cnt := range c.rootContainers {
		bounds := cnt.layout.Bounds
		if cnt.collapsed {
			bounds.Max.Y = cnt.layout.BodyBounds.Min.Y
		}
		if pt.In(bounds) {
			inputCapturingState |= InputCapturingStateHover
		}
	}

	// Check whether there is a focused widget like a text field.
	if c.focus != (widgetID{}) {
		inputCapturingState |= InputCapturingStateFocus
	}
	return inputCapturingState, nil
}

func (c *Context) beginUpdate() {
	for _, cnt := range c.idToContainer {
		cnt.used = false
	}
	for _, cnt := range c.rootContainers {
		cnt.commandList = slices.Delete(cnt.commandList, 0, len(cnt.commandList))
	}
	c.scrollTarget = nil
	c.currentID = widgetID{}
}

func (c *Context) endUpdate() error {
	// check stacks
	if c.idStack.size > 0 {
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
		c.focus = widgetID{}
	}
	c.keepFocus = false

	// Bring the hovering root container to front if the pointing device was pressed.
	if c.pointing.justPressed() {
		// TODO: When showing a popup, the position might be on the popup and the parent container might not be brought to front.
		// Fix this issue.
		if cnt := c.hoveringRootContainer(); cnt != nil {
			c.bringToFront(cnt)
		}
	}

	// reset input state
	c.lastPointingPos = c.pointingPosition()

	// Remove unused containers.
	c.rootContainers = slices.DeleteFunc(c.rootContainers, func(cnt *container) bool {
		return !cnt.used
	})
	maps.DeleteFunc(c.idToContainer, func(id widgetID, cnt *container) bool {
		return !cnt.used
	})

	return nil
}

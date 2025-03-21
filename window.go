// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/ebitengine/debugui/internal/caller"
)

func (c *Context) Window(title string, rect image.Rectangle, f func(layout ContainerLayout)) {
	pc := caller.Caller()
	c.wrapError(func() error {
		if err := c.window(title, rect, 0, pc, f); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) window(title string, bounds image.Rectangle, opt option, callerPC uintptr, f func(layout ContainerLayout)) (err error) {
	id := c.idFromGlobalUniqueString(title)

	cnt := c.container(id, opt)
	if cnt == nil || !cnt.open {
		return nil
	}
	if cnt.layout.Bounds.Dx() == 0 {
		cnt.layout.Bounds = bounds
	}

	c.containerStack = append(c.containerStack, cnt)
	defer c.popContainer()

	// push container to roots list and push head command
	c.rootList = append(c.rootList, cnt)
	cnt.headIdx = c.appendJumpCommand(-1)
	defer func() {
		// push tail 'goto' jump command and set head 'skip' command. the final steps
		// on initing these are done in End
		cnt := c.currentContainer()
		cnt.tailIdx = c.appendJumpCommand(-1)
		c.commandList[cnt.headIdx].jump.dstIdx = len(c.commandList) //- 1
	}()

	// set as hover root if the mouse is overlapping this container and it has a
	// higher zindex than the current hover root
	if c.cursorPosition().In(cnt.layout.Bounds) && (c.nextHoverRoot == nil || cnt.zIndex > c.nextHoverRoot.zIndex) {
		c.nextHoverRoot = cnt
	}

	// clipping is reset here in case a root-container is made within
	// another root-containers's begin/end block; this prevents the inner
	// root-container being clipped to the outer
	c.clipStack = append(c.clipStack, unclippedRect)
	defer c.popClipRect()

	body := cnt.layout.Bounds
	bounds = body

	// draw frame
	collapsed := cnt.collapsed
	if (^opt&optionNoFrame) != 0 && !collapsed {
		c.drawFrame(bounds, ColorWindowBG)
	}

	// do title bar
	if (^opt & optionNoTitle) != 0 {
		tr := bounds
		tr.Max.Y = tr.Min.Y + c.style().titleHeight
		c.drawFrame(tr, ColorTitleBG)

		// do title text
		if (^opt & optionNoTitle) != 0 {
			id := c.idFromCaller(callerPC, "!title")
			r := image.Rect(tr.Min.X+tr.Dy()-c.style().padding, tr.Min.Y, tr.Max.X, tr.Max.Y)
			c.updateControl(id, r, opt)
			c.drawControlText(title, r, ColorTitleText, opt)
			if id == c.focus && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				cnt.layout.Bounds = cnt.layout.Bounds.Add(c.mouseDelta())
			}
			body.Min.Y += tr.Dy()
		}

		// do `collapse` button
		if (^opt & optionNoClose) != 0 {
			id := c.idFromCaller(callerPC, "!collapse")
			r := image.Rect(tr.Min.X, tr.Min.Y, tr.Min.X+tr.Dy(), tr.Max.Y)
			icon := iconExpanded
			if collapsed {
				icon = iconCollapsed
			}
			c.drawIcon(icon, r, c.style().colors[ColorTitleText])
			c.updateControl(id, r, opt)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && id == c.focus {
				cnt.collapsed = !cnt.collapsed
			}
		}
	}

	if collapsed {
		return nil
	}

	c.pushContainerBodyLayout(cnt, body, opt, callerPC)
	defer func() {
		if err2 := c.popLayout(); err2 != nil && err == nil {
			err = err2
		}
	}()

	// do `resize` handle
	if (^opt & optionNoResize) != 0 {
		sz := c.style().titleHeight
		id := c.idFromCaller(callerPC, "!resize")
		r := image.Rect(bounds.Max.X-sz, bounds.Max.Y-sz, bounds.Max.X, bounds.Max.Y)
		c.updateControl(id, r, opt)
		if id == c.focus && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.Bounds.Max.X = cnt.layout.Bounds.Min.X + max(96, cnt.layout.Bounds.Dx()+c.mouseDelta().X)
			cnt.layout.Bounds.Max.Y = cnt.layout.Bounds.Min.Y + max(64, cnt.layout.Bounds.Dy()+c.mouseDelta().Y)
		}
	}

	// resize to content size
	if (opt & optionAutoSize) != 0 {
		l, err := c.layout()
		if err != nil {
			return err
		}
		r := l.body
		cnt.layout.Bounds.Max.X = cnt.layout.Bounds.Min.X + cnt.layout.ContentSize.X + (cnt.layout.Bounds.Dx() - r.Dx())
		cnt.layout.Bounds.Max.Y = cnt.layout.Bounds.Min.Y + cnt.layout.ContentSize.Y + (cnt.layout.Bounds.Dy() - r.Dy())
	}

	// close if this is a popup window and elsewhere was clicked
	if (opt&optionPopup) != 0 && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.hoverRoot != cnt {
		cnt.open = false
	}

	c.pushClipRect(cnt.layout.BodyBounds)
	defer c.popClipRect()

	f(c.currentContainer().layout)

	return nil
}

func (c *Context) OpenPopup(name string) {
	c.wrapError(func() error {
		id := c.idFromGlobalUniqueString(name)
		cnt := c.container(id, 0)
		// set as hover root so popup isn't closed in begin_window_ex()
		c.nextHoverRoot = cnt
		c.hoverRoot = c.nextHoverRoot
		// position at mouse cursor, open and bring-to-front
		pt := c.cursorPosition()
		cnt.layout.Bounds = image.Rectangle{
			Min: pt,
			Max: pt.Add(image.Pt(1, 1)),
		}
		cnt.open = true
		c.bringToFront(cnt)
		return nil
	})
}

func (c *Context) ClosePopup(name string) {
	c.wrapError(func() error {
		id := c.idFromGlobalUniqueString(name)
		cnt := c.container(id, 0)
		cnt.open = false
		return nil
	})
}

func (c *Context) Popup(name string, f func(layout ContainerLayout)) {
	pc := caller.Caller()
	c.wrapError(func() error {
		opt := optionPopup | optionAutoSize | optionNoResize | optionNoScroll | optionNoTitle | optionClosed
		if err := c.window(name, image.Rectangle{}, opt, pc, f); err != nil {
			return err
		}
		return nil
	})
}

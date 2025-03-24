// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type container struct {
	layout    ContainerLayout
	headIdx   int
	tailIdx   int
	zIndex    int
	open      bool
	collapsed bool

	toggledIDs          map[controlID]struct{}
	textInputTextFields map[controlID]*textinput.Field
}

// ContainerLayout represents the layout of a container control.
type ContainerLayout struct {
	// Bounds is the bounds of the control.
	Bounds image.Rectangle

	// BodyBounds is the bounds of the body area of the container.
	BodyBounds image.Rectangle

	// ContentSize is the size of the content.
	// ContentSize can be larger than Bounds or BodyBounds. In this case, the control should be scrollable.
	ContentSize image.Point

	// ScrollOffset is the offset of the scroll.
	ScrollOffset image.Point
}

func (c *Context) container(id controlID, opt option) *container {
	if container, ok := c.idToContainer[id]; ok {
		c.addUsedContainer(id)
		return container
	}

	if (opt & optionClosed) != 0 {
		return nil
	}

	if c.idToContainer == nil {
		c.idToContainer = map[controlID]*container{}
	}
	cnt := &container{
		headIdx: -1,
		tailIdx: -1,
		open:    true,
	}
	c.idToContainer[id] = cnt
	c.addUsedContainer(id)
	c.bringToFront(cnt)
	return cnt
}

func (c *Context) currentContainer() *container {
	return c.containerStack[len(c.containerStack)-1]
}

func (c *Context) Window(title string, rect image.Rectangle, f func(layout ContainerLayout)) {
	c.wrapError(func() error {
		if err := c.window(title, rect, 0, f); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) window(title string, bounds image.Rectangle, opt option, f func(layout ContainerLayout)) (err error) {
	id := c.idFromGlobalString(title)
	c.idScopeFromGlobalString(title, func() {
		err = c.doWindow(title, bounds, opt, id, f)
	})
	return
}

func (c *Context) doWindow(title string, bounds image.Rectangle, opt option, id controlID, f func(layout ContainerLayout)) (err error) {
	cnt := c.container(id, opt)
	if cnt == nil || !cnt.open {
		return nil
	}
	if cnt.layout.Bounds.Dx() == 0 {
		cnt.layout.Bounds = bounds
	}

	c.pushContainer(cnt)
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
		c.drawFrame(bounds, colorWindowBG)
	}

	// do title bar
	if (^opt & optionNoTitle) != 0 {
		tr := bounds
		tr.Max.Y = tr.Min.Y + c.style().titleHeight
		c.drawFrame(tr, colorTitleBG)

		// do title text
		if (^opt & optionNoTitle) != 0 {
			id := c.idFromString("title")
			r := image.Rect(tr.Min.X+tr.Dy()-c.style().padding, tr.Min.Y, tr.Max.X, tr.Max.Y)
			c.updateControl(id, r, opt)
			c.drawControlText(title, r, colorTitleText, opt)
			if id == c.focus && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				b := cnt.layout.Bounds.Add(c.mouseDelta())
				if c.screenWidth > 0 {
					maxX := b.Max.X
					if maxX >= c.screenWidth/c.Scale() {
						b = b.Add(image.Pt(c.screenWidth/c.Scale()-maxX, 0))
					}
				}
				if b.Min.X < 0 {
					b = b.Add(image.Pt(-b.Min.X, 0))
				}
				if c.screenHeight > 0 {
					maxY := b.Min.Y + tr.Dy()
					if maxY >= c.screenHeight/c.Scale()-c.style().padding {
						b = b.Add(image.Pt(0, c.screenHeight/c.Scale()-maxY))
					}
				}
				if b.Min.Y < 0 {
					b = b.Add(image.Pt(0, -b.Min.Y))
				}
				cnt.layout.Bounds = b
			}
			body.Min.Y += tr.Dy()
		}

		// do `collapse` button
		if (^opt & optionNoClose) != 0 {
			id := c.idFromString("collapse")
			r := image.Rect(tr.Min.X, tr.Min.Y, tr.Min.X+tr.Dy(), tr.Max.Y)
			icon := iconExpanded
			if collapsed {
				icon = iconCollapsed
			}
			c.drawIcon(icon, r, c.style().colors[colorTitleText])
			c.updateControl(id, r, opt)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && id == c.focus {
				cnt.collapsed = !cnt.collapsed
			}
		}
	}

	if collapsed {
		return nil
	}

	if err := c.pushContainerBodyLayout(cnt, body, opt); err != nil {
		return err
	}
	defer func() {
		if err2 := c.popLayout(); err2 != nil && err == nil {
			err = err2
		}
	}()

	// do `resize` handle
	if (^opt & optionNoResize) != 0 {
		sz := c.style().titleHeight
		id := c.idFromString("resize")
		r := image.Rect(bounds.Max.X-sz, bounds.Max.Y-sz, bounds.Max.X, bounds.Max.Y)
		c.updateControl(id, r, opt)
		if id == c.focus && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.Bounds.Max.X = min(cnt.layout.Bounds.Min.X+max(96, cnt.layout.Bounds.Dx()+c.mouseDelta().X), c.screenWidth/c.Scale())
			cnt.layout.Bounds.Max.Y = min(cnt.layout.Bounds.Min.Y+max(64, cnt.layout.Bounds.Dy()+c.mouseDelta().Y), c.screenHeight/c.Scale())
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
	id := c.idFromGlobalString(name)
	c.wrapError(func() error {
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
	id := c.idFromGlobalString(name)
	c.wrapError(func() error {
		cnt := c.container(id, 0)
		cnt.open = false
		return nil
	})
}

func (c *Context) Popup(name string, f func(layout ContainerLayout)) {
	c.wrapError(func() error {
		opt := optionPopup | optionAutoSize | optionNoResize | optionNoScroll | optionNoTitle | optionClosed
		if err := c.window(name, image.Rectangle{}, opt, f); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) pushContainer(cnt *container) {
	c.containerStack = append(c.containerStack, cnt)
}

func (c *Context) popContainer() {
	c.containerStack = c.containerStack[:len(c.containerStack)-1]
}

func (c *Context) SetScroll(scroll image.Point) {
	c.currentContainer().layout.ScrollOffset = scroll
}

func (c *container) textInputTextField(id controlID) *textinput.Field {
	if id == emptyControlID {
		return nil
	}
	if _, ok := c.textInputTextFields[id]; !ok {
		if c.textInputTextFields == nil {
			c.textInputTextFields = make(map[controlID]*textinput.Field)
		}
		c.textInputTextFields[id] = &textinput.Field{}
	}
	return c.textInputTextFields[id]
}

func (c *container) toggled(id controlID) bool {
	_, ok := c.toggledIDs[id]
	return ok
}

func (c *container) toggle(id controlID) {
	if _, toggled := c.toggledIDs[id]; toggled {
		delete(c.toggledIDs, id)
		return
	}
	if c.toggledIDs == nil {
		c.toggledIDs = map[controlID]struct{}{}
	}
	c.toggledIDs[id] = struct{}{}
}

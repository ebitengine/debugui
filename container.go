// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

type container struct {
	parent *container

	layout    ContainerLayout
	open      bool
	collapsed bool

	// commandList is valid only for root containers.
	// See the implementation of appendCommand which is the only place to append commands.
	commandList []*command

	toggledIDs          map[widgetID]struct{}
	textInputTextFields map[widgetID]*textinput.Field

	// dropdownCloseDelay is used for delayed closing of dropdowns
	dropdownCloseDelay int

	used bool
}

// ContainerLayout represents the layout of a container widget.
type ContainerLayout struct {
	// Bounds is the bounds of the widget.
	Bounds image.Rectangle

	// BodyBounds is the bounds of the body area of the container.
	BodyBounds image.Rectangle

	// ContentSize is the size of the content.
	// ContentSize can be larger than Bounds or BodyBounds. In this case, the widget should be scrollable.
	ContentSize image.Point

	// ScrollOffset is the offset of the scroll.
	ScrollOffset image.Point
}

func (c *Context) container(id widgetID, opt option) *container {
	if container, ok := c.idToContainer[id]; ok {
		container.used = true
		return container
	}

	if (opt & optionClosed) != 0 {
		return nil
	}

	if c.idToContainer == nil {
		c.idToContainer = map[widgetID]*container{}
	}
	cnt := &container{
		open: true,
	}
	c.idToContainer[id] = cnt
	cnt.used = true
	return cnt
}

func (c *Context) currentContainer() *container {
	return c.containerStack[len(c.containerStack)-1]
}

func (c *Context) currentRootContainer() *container {
	var cnt *container
	for cnt = c.currentContainer(); cnt != nil && cnt.parent != nil; cnt = cnt.parent {
	}
	return cnt
}

// Window creates a new window with the contents defined by the function f.
//
// title is the title of the window.
// rect is the initial size and position of the window.
func (c *Context) Window(title string, initialBounds image.Rectangle, f func(layout ContainerLayout)) {
	pc := caller()
	idPart := idPartFromCaller(pc)
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if err := c.window(title, initialBounds, 0, idPart, f); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (c *Context) window(title string, initialBounds image.Rectangle, opt option, idPart string, f func(layout ContainerLayout)) error {
	// A window is not a widget in the current implementation, but a window is a widget in the concept.
	var err error
	c.idScopeFromIDPart(idPart, func(id widgetID) {
		c.currentID = id
		err = c.doWindow(title, initialBounds, opt, id, f)
	})
	return err
}

func (c *Context) doWindow(title string, initialBounds image.Rectangle, opt option, id widgetID, f func(layout ContainerLayout)) (err error) {
	cnt := c.container(id, opt)
	if cnt == nil || !cnt.open {
		return nil
	}
	if cnt.layout.Bounds.Dx() == 0 {
		cnt.layout.Bounds = initialBounds
	}

	c.pushContainer(cnt, true)
	defer c.popContainer()

	if !slices.Contains(c.rootContainers, cnt) {
		c.rootContainers = append(c.rootContainers, cnt)
	}

	// clipping is reset here in case a root-container is made within
	// another root-containers's begin/end block; this prevents the inner
	// root-container being clipped to the outer
	c.clipStack = append(c.clipStack, unclippedRect)
	defer c.popClipRect()

	body := cnt.layout.Bounds
	bounds := body

	// draw frame
	collapsed := cnt.collapsed
	if (^opt&optionNoFrame) != 0 && !collapsed {
		c.drawFrame(bounds, colorWindowBG)
	}

	// do title bar
	if (^opt & optionNoTitle) != 0 {
		tr := bounds
		tr.Max.Y = tr.Min.Y + c.style().titleHeight
		if !collapsed {
			c.drawFrame(tr, colorTitleBG)
		} else {
			c.drawFrame(tr, colorTitleBGTransparent)
		}

		// do title text
		if (^opt & optionNoTitle) != 0 {
			titleID := id.push(idPartFromString("title"))
			r := image.Rect(tr.Min.X+tr.Dy()-c.style().padding, tr.Min.Y, tr.Max.X, tr.Max.Y)
			_ = c.widgetWithBounds(titleID, opt, r, func(bounds image.Rectangle, wasFocused bool) EventHandler {
				if titleID == c.focus && c.pointing.pressed() {
					b := cnt.layout.Bounds.Add(c.pointingDelta())
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
				return nil
			}, func(bounds image.Rectangle) {
				c.drawWidgetText(title, r, colorTitleText, opt)
			})
		}

		// do `collapse` button
		if (^opt & optionNoClose) != 0 {
			collapseID := id.push(idPartFromString("collapse"))
			r := image.Rect(tr.Min.X, tr.Min.Y, tr.Min.X+tr.Dy(), tr.Max.Y)
			_ = c.widgetWithBounds(collapseID, opt, r, func(bounds image.Rectangle, wasFocused bool) EventHandler {
				if c.pointing.justPressed() && collapseID == c.focus {
					cnt.collapsed = !cnt.collapsed
				}
				return nil
			}, func(bounds image.Rectangle) {
				icon := iconExpanded
				if collapsed {
					icon = iconCollapsed
				}
				c.drawIcon(icon, r, c.style().colors[colorTitleText])
			})
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
		resizeID := id.push(idPartFromString("resize"))
		r := image.Rect(bounds.Max.X-sz, bounds.Max.Y-sz, bounds.Max.X, bounds.Max.Y)
		_ = c.widgetWithBounds(resizeID, 0, r, func(bounds image.Rectangle, wasFocused bool) EventHandler {
			if resizeID == c.focus && c.pointing.pressed() {
				cnt.layout.Bounds.Max.X = min(cnt.layout.Bounds.Min.X+max(96, cnt.layout.Bounds.Dx()+c.pointingDelta().X), c.screenWidth/c.Scale())
				cnt.layout.Bounds.Max.Y = min(cnt.layout.Bounds.Min.Y+max(64, cnt.layout.Bounds.Dy()+c.pointingDelta().Y), c.screenHeight/c.Scale())
			}
			return nil
		}, nil)
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
	if (opt&optionPopup) != 0 && c.pointing.justPressed() && c.hoveringRootContainer() != cnt {
		cnt.open = false
	}

	c.pushClipRect(cnt.layout.BodyBounds)
	defer c.popClipRect()

	f(c.currentContainer().layout)

	return nil
}

// PopupID is the ID of a popup window.
type PopupID widgetID

// OpenPopup opens a popup window at the current pointing position.
func (c *Context) OpenPopup(popupID PopupID) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		cnt := c.container(widgetID(popupID), 0)
		// Position at pointing cursor, open and bring-to-front.
		pt := c.pointingPosition()
		cnt.layout.Bounds = image.Rectangle{
			Min: pt,
			Max: pt.Add(image.Pt(1, 1)),
		}
		cnt.open = true
		c.bringToFront(cnt)
		return nil, nil
	})
}

// ClosePopup closes a popup window.
func (c *Context) ClosePopup(popupID PopupID) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		cnt := c.container(widgetID(popupID), 0)
		cnt.open = false
		return nil, nil
	})
}

// Popup creates a popup window with the content defined by the provided function,
// and returns the PopupID of the popup window.
//
// By default, the popup window is hidden.
// To show the popup window, call OpenPopup with the PopupID returned by this function.
func (c *Context) Popup(f func(layout ContainerLayout, popupID PopupID)) PopupID {
	pc := caller()
	idPart := idPartFromCaller(pc)
	id := c.idStack.push(idPart)
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		opt := optionPopup | optionAutoSize | optionNoResize | optionNoScroll | optionNoTitle | optionClosed
		if err := c.window("", image.Rectangle{}, opt, idPart, func(layout ContainerLayout) {
			f(layout, PopupID(id))
		}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return PopupID(id)
}

func (c *Context) pushContainer(cnt *container, root bool) {
	if !root && len(c.containerStack) > 0 {
		cnt.parent = c.containerStack[len(c.containerStack)-1]
	}
	c.containerStack = append(c.containerStack, cnt)
}

func (c *Context) pushContainerBodyLayout(cnt *container, body image.Rectangle, opt option) error {
	if (^opt & optionNoScroll) != 0 {
		body = c.scrollbars(cnt, body)
	}
	if err := c.pushLayout(body.Inset(c.style().padding), cnt.layout.ScrollOffset, opt&optionAutoSize != 0); err != nil {
		return err
	}
	cnt.layout.BodyBounds = body
	return nil
}

func (c *Context) popContainer() {
	c.containerStack = c.containerStack[:len(c.containerStack)-1]
}

// SetScroll sets the scroll offset of the current container.
func (c *Context) SetScroll(scroll image.Point) {
	c.currentContainer().layout.ScrollOffset = scroll
}

func (c *container) textInputTextField(id widgetID, createIfNeeded bool) *textinput.Field {
	if id == (widgetID{}) {
		return nil
	}
	if _, ok := c.textInputTextFields[id]; !ok {
		if !createIfNeeded {
			return nil
		}
		if c.textInputTextFields == nil {
			c.textInputTextFields = make(map[widgetID]*textinput.Field)
		}
		c.textInputTextFields[id] = &textinput.Field{}
	}
	return c.textInputTextFields[id]
}

func (c *container) toggled(id widgetID) bool {
	_, ok := c.toggledIDs[id]
	return ok
}

func (c *container) toggle(id widgetID) {
	if _, toggled := c.toggledIDs[id]; toggled {
		delete(c.toggledIDs, id)
		return
	}
	if c.toggledIDs == nil {
		c.toggledIDs = map[widgetID]struct{}{}
	}
	c.toggledIDs[id] = struct{}{}
}

func (c *Context) bringToFront(cnt *container) {
	idx := slices.IndexFunc(c.rootContainers, func(c *container) bool {
		return c == cnt
	})
	if idx == len(c.rootContainers)-1 {
		return
	}
	if idx >= 0 {
		c.rootContainers = slices.Delete(c.rootContainers, idx, idx+1)
	}
	c.rootContainers = append(c.rootContainers, cnt)
}

func (c *Context) hoveringRootContainer() *container {
	p := c.pointingPosition()
	for i := len(c.rootContainers) - 1; i >= 0; i-- {
		cnt := c.rootContainers[i]
		if !cnt.open {
			continue
		}
		if p.In(cnt.layout.Bounds) {
			return cnt
		}
	}
	return nil
}

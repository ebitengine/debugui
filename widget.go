// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
)

// widgetID is a unique identifier for a widget.
//
// Do not rely on the string value of widgetID, as it is not guaranteed to be stable across different runs of the program.
type widgetID struct {
	idParts [16]string
	size    int
}

func (w widgetID) push(idPart string) widgetID {
	if w.size >= len(w.idParts) {
		panic("debugui: too many ID parts")
	}
	w.idParts[w.size] = idPart
	w.size++
	return w
}

func (w widgetID) pop() widgetID {
	if w.size <= 0 {
		panic("debugui: no ID parts to pop")
	}
	w.size--
	w.idParts[w.size] = ""
	return w
}

type option int

const (
	optionAlignCenter option = (1 << iota)
	optionAlignRight
	optionNoInteract
	optionNoFrame
	optionNoResize
	optionNoScroll
	optionNoClose
	optionNoTitle
	optionHoldFocus
	optionAutoSize
	optionPopup
	optionClosed
	optionExpanded
)

func (c *Context) pointingOver(bounds image.Rectangle) bool {
	p := c.pointingPosition()
	if !p.In(bounds) {
		return false
	}
	if !p.In(c.clipRect()) {
		return false
	}
	return c.hoveringRootContainer() == c.currentRootContainer()
}

func (c *Context) pointingDelta() image.Point {
	// The delta is always (0, 0) when a touch just started.
	if c.pointing.isTouchActive() && c.pointing.justPressed() {
		return image.Point{}
	}
	return c.pointingPosition().Sub(c.lastPointingPos)
}

func (c *Context) pointingPosition() image.Point {
	p := c.pointing.position()
	p.X /= c.Scale()
	p.Y /= c.Scale()
	return p
}

func (c *Context) handleInputForWidget(id widgetID, bounds image.Rectangle, opt option) (wasFocused bool) {
	if id == (widgetID{}) {
		return false
	}

	if c.focus == id {
		c.keepFocus = true
	}
	if (opt & optionNoInteract) != 0 {
		return false
	}

	hover := c.pointingOver(bounds)
	if hover {
		c.hover = id
	}

	if c.focus == id {
		if c.pointing.justPressed() && !hover {
			c.setFocus(widgetID{})
			wasFocused = true
		}
		if !c.pointing.pressed() && (^opt&optionHoldFocus) != 0 {
			c.setFocus(widgetID{})
			wasFocused = true
		}
	}

	if c.hover == id {
		if c.pointing.justPressed() {
			c.setFocus(id)
		} else if !hover {
			c.hover = widgetID{}
		}
	}

	return
}

func (c *Context) widget(id widgetID, opt option, layout func(bounds image.Rectangle), handleInput func(bounds image.Rectangle, wasFocused bool) EventHandler, draw func(bounds image.Rectangle)) (EventHandler, error) {
	c.currentID = id
	bounds, err := c.layoutNext()
	if err != nil {
		return nil, err
	}

	if layout != nil {
		if err := c.pushLayout(bounds, image.Pt(0, 0), false); err != nil {
			return nil, err
		}
		defer func() {
			b := &c.layoutStack[len(c.layoutStack)-1]
			// inherit position/next_row/max from child layout if they are greater
			a := &c.layoutStack[len(c.layoutStack)-2]
			a.position.X = max(a.position.X, b.position.X+b.body.Min.X-a.body.Min.X)
			a.nextRowY = max(a.nextRowY, b.nextRowY+b.body.Min.Y-a.body.Min.Y)
			a.max.X = max(a.max.X, b.max.X)
			a.max.Y = max(a.max.Y, b.max.Y)
			if err2 := c.popLayout(); err2 != nil && err == nil {
				err = err2
			}
		}()
		layout(bounds)
	}

	wasFocused := c.handleInputForWidget(id, bounds, opt)
	var e EventHandler
	if handleInput != nil {
		e = handleInput(bounds, wasFocused)
	}
	// Handling input is still needed even if the widget is out of bounds, especially for Header.
	if !c.currentContainer().layout.BodyBounds.Overlaps(bounds) {
		return e, nil
	}

	if draw != nil {
		draw(bounds)
	}
	return e, nil
}

func (c *Context) widgetWithBounds(id widgetID, opt option, bounds image.Rectangle, handleInput func(bounds image.Rectangle, wasFocused bool) EventHandler, draw func(bounds image.Rectangle)) EventHandler {
	c.currentID = id

	wasFocused := c.handleInputForWidget(id, bounds, opt)
	var e EventHandler
	if handleInput != nil {
		e = handleInput(bounds, wasFocused)
	}
	if draw != nil {
		draw(bounds)
	}
	return e
}

// Checkbox creates a checkbox with the given boolean state and text label.
//
// A Checkbox widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Checkbox(state *bool, label string) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.widget(id, 0, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
			var e EventHandler
			c.handleInputForWidget(id, bounds, 0)
			if c.pointing.justPressed() && c.focus == id {
				e = &eventHandler{}
				*state = !*state
			}
			return e
		}, func(bounds image.Rectangle) {
			box := image.Rect(bounds.Min.X, bounds.Min.Y+(bounds.Dy()-lineHeight())/2, bounds.Min.X+lineHeight(), bounds.Max.Y-(bounds.Dy()-lineHeight())/2)
			c.drawWidgetFrame(id, box, colorBase, 0)
			if *state {
				c.drawIcon(iconCheck, box, c.style().colors[colorText])
			}
			if label != "" {
				bounds = image.Rect(bounds.Min.X+lineHeight(), bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
				c.drawWidgetText(label, bounds, colorText, 0)
			}
		})
	})
}

func (c *Context) setFocus(id widgetID) {
	c.focus = id
	c.keepFocus = true
}

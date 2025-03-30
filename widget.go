// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
	"slices"
)

// WidgetID is a unique identifier for a widget.
//
// Do not rely on the string value of WidgetID, as it is not guaranteed to be stable across different runs of the program.
type WidgetID string

const emptyWidgetID WidgetID = ""

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
	return p.In(bounds) && p.In(c.clipRect()) && slices.Contains(c.containerStack, c.hoverRoot)
}

func (c *Context) pointingDelta() image.Point {
	return c.pointingPosition().Sub(c.lastPointingPos)
}

func (c *Context) pointingPosition() image.Point {
	p := c.pointing.position()
	p.X /= c.Scale()
	p.Y /= c.Scale()
	return p
}

func (c *Context) updateWidget(id WidgetID, bounds image.Rectangle, opt option) (wasFocused bool) {
	if id == emptyWidgetID {
		return false
	}

	pointingOver := c.pointingOver(bounds)

	if c.focus == id {
		c.keepFocus = true
	}
	if (opt & optionNoInteract) != 0 {
		return false
	}
	if pointingOver && !c.pointing.pressed() {
		c.hover = id
	}

	if c.focus == id {
		if c.pointing.justPressed() && !pointingOver {
			c.setFocus(emptyWidgetID)
			wasFocused = true
		}
		if !c.pointing.pressed() && (^opt&optionHoldFocus) != 0 {
			c.setFocus(emptyWidgetID)
			wasFocused = true
		}
	}

	if c.hover == id {
		if c.pointing.justPressed() {
			c.setFocus(id)
		} else if !pointingOver {
			c.hover = emptyWidgetID
		}
	}

	return
}

func (c *Context) Widget(f func(bounds image.Rectangle) bool) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.widget(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
			return f(bounds), nil
		})
	})
}

func (c *Context) widget(id WidgetID, opt option, f func(bounds image.Rectangle, wasFocused bool) (bool, error)) (EventHandler, error) {
	c.currentID = id
	r, err := c.layoutNext()
	if err != nil {
		return nil, err
	}
	wasFocused := c.updateWidget(id, r, opt)
	res, err := f(r, wasFocused)
	if err != nil {
		return nil, err
	}
	return &eventHandler{res: res}, nil
}

// Checkbox creates a checkbox with the given boolean state and text label.
//
// A Checkbox widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Checkbox(state *bool, label string) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.widget(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
			var res bool
			box := image.Rect(bounds.Min.X, bounds.Min.Y+(bounds.Dy()-lineHeight())/2, bounds.Min.X+lineHeight(), bounds.Max.Y-(bounds.Dy()-lineHeight())/2)
			c.updateWidget(id, bounds, 0)
			if c.pointing.justPressed() && c.focus == id {
				res = true
				*state = !*state
			}
			c.drawWidgetFrame(id, box, colorBase, 0)
			if *state {
				c.drawIcon(iconCheck, box, c.style().colors[colorText])
			}
			if label != "" {
				bounds = image.Rect(bounds.Min.X+lineHeight(), bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
				c.drawWidgetText(label, bounds, colorText, 0)
			}
			return res, nil
		})
	})
}

func (c *Context) isCapturingInput() bool {
	if c.err != nil {
		return false
	}

	// Check whether the cursor is on any of the root containers.
	pt := c.pointingPosition()
	for _, cnt := range c.rootContainers {
		if pt.In(cnt.layout.Bounds) {
			return true
		}
	}

	// Check whether there is a focused widget like a text field.
	return c.focus != emptyWidgetID
}

// CurrentWidgetID returns the ID of the current widget being processed.
func (c *Context) CurrentWidgetID() WidgetID {
	return c.currentID
}

func (c *Context) setFocus(id WidgetID) {
	c.focus = id
	c.keepFocus = true
}

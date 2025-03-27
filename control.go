// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
)

type controlID string

const emptyControlID controlID = ""

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

func (c *Context) inHoverRoot() bool {
	for i := len(c.containerStack) - 1; i >= 0; i-- {
		if c.containerStack[i] == c.hoverRoot {
			return true
		}
		// only root containers have their `head` field set; stop searching if we've
		// reached the current root container
		if c.containerStack[i].headIdx >= 0 {
			break
		}
	}
	return false
}

func (c *Context) pointingOver(bounds image.Rectangle) bool {
	p := c.pointingPosition()
	return p.In(bounds) && p.In(c.clipRect()) && c.inHoverRoot()
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

func (c *Context) updateControl(id controlID, bounds image.Rectangle, opt option) (wasFocused bool) {
	if id == emptyControlID {
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
			c.setFocus(emptyControlID)
			wasFocused = true
		}
		if !c.pointing.pressed() && (^opt&optionHoldFocus) != 0 {
			c.setFocus(emptyControlID)
			wasFocused = true
		}
	}

	if c.hover == id {
		if c.pointing.justPressed() {
			c.setFocus(id)
		} else if !pointingOver {
			c.hover = emptyControlID
		}
	}

	return
}

func (c *Context) Control(f func(bounds image.Rectangle) bool) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
			return f(bounds), nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) control(id controlID, opt option, f func(bounds image.Rectangle, wasFocused bool) (bool, error)) (bool, error) {
	r, err := c.layoutNext()
	if err != nil {
		return false, err
	}
	wasFocused := c.updateControl(id, r, opt)
	res, err := f(r, wasFocused)
	if err != nil {
		return false, err
	}
	return res, nil
}

// Checkbox creates a checkbox with the given boolean state and text label.
//
// A Checkbox control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Checkbox(state *bool, label string) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
			var res bool
			box := image.Rect(bounds.Min.X, bounds.Min.Y+(bounds.Dy()-lineHeight())/2, bounds.Min.X+lineHeight(), bounds.Max.Y-(bounds.Dy()-lineHeight())/2)
			c.updateControl(id, bounds, 0)
			if c.pointing.justPressed() && c.focus == id {
				res = true
				*state = !*state
			}
			c.drawControlFrame(id, box, colorBase, 0)
			if *state {
				c.drawIcon(iconCheck, box, c.style().colors[colorText])
			}
			if label != "" {
				bounds = image.Rect(bounds.Min.X+lineHeight(), bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
				c.drawControlText(label, bounds, colorText, 0)
			}
			return res, nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) isCapturingInput() bool {
	if c.err != nil {
		return false
	}

	return c.hoverRoot != nil || c.focus != emptyControlID
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// Button creates a button widget with the given text.
//
// Button returns an EventHandler to handle click events.
// A returned EventHandler is never nil.
//
// A Button widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Button(text string) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.button(text, optionAlignCenter, id)
	})
}

func (c *Context) button(text string, opt option, id widgetID) (EventHandler, error) {
	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler
		if c.pointing.justPressed() && c.focus == id {
			e = &eventHandler{}
		}
		return e
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorButton, opt)
		if len(text) > 0 {
			c.drawWidgetText(text, bounds, colorText, opt)
		}
	})
}

func (c *Context) spinButtons(id widgetID) (up, down EventHandler) {
	upID := id.push(idPartFromString("up"))
	downID := id.push(idPartFromString("down"))
	c.GridCell(func(bounds image.Rectangle) {
		c.SetGridLayout(nil, []int{-1, -1})
		up = c.wrapEventHandlerAndError(func() (EventHandler, error) {
			e, err := c.spinButton(true, optionAlignCenter, upID, downID)
			if err != nil {
				return nil, err
			}
			return e, nil
		})
		down = c.wrapEventHandlerAndError(func() (EventHandler, error) {
			e, err := c.spinButton(false, optionAlignCenter, upID, downID)
			if err != nil {
				return nil, err
			}
			return e, nil
		})
	})
	return up, down
}

func (c *Context) spinButton(up bool, opt option, upID, downID widgetID) (EventHandler, error) {
	id := downID
	if up {
		id = upID
	}
	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler
		if c.pointing.repeated() && (c.focus == upID || c.focus == downID) && c.pointingPosition().In(bounds) {
			e = &eventHandler{}
		}
		return e
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorButton, opt)
		icon := iconDown
		if up {
			icon = iconUp
		}
		c.drawIcon(icon, bounds, c.style().colors[colorText])
	})
}

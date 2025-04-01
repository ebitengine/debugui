// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// Button creates a button widget with the given label.
//
// Button returns an EventHandler to handle click events.
// A returned EventHandler is never nil.
//
// A Button widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Button(label string) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.button(label, optionAlignCenter, id)
	})
}

func (c *Context) button(label string, opt option, id WidgetID) (EventHandler, error) {
	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) (EventHandler, error) {
		var e EventHandler
		if c.pointing.justPressed() && c.focus == id {
			e = &eventHandler{}
		}
		return e, nil
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorButton, opt)
		if len(label) > 0 {
			c.drawWidgetText(label, bounds, colorText, opt)
		}
	})
}

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
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.button(text, iconNone, optionAlignCenter, id)
	})
}

func (c *Context) iconButton(icon icon) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.button("", icon, optionAlignCenter, id)
	})
}

func (c *Context) button(text string, icon icon, opt option, id widgetID) (EventHandler, error) {
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
		if icon != iconNone {
			c.drawIcon(icon, bounds, c.style().colors[colorText])
		}
	})
}

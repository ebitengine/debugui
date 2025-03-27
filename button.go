// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// Button creates a button widget with the given label.
//
// Button returns true if the button has been clicked, otherwise false.
//
// A Button widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Button(label string) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.button(label, optionAlignCenter, id)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) button(label string, opt option, id widgetID) (bool, error) {
	res, err := c.widget(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		// handle click
		if c.pointing.justPressed() && c.focus == id {
			res = true
		}
		// draw
		c.drawWidgetFrame(id, bounds, colorButton, opt)
		if len(label) > 0 {
			c.drawWidgetText(label, bounds, colorText, opt)
		}
		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// Button creates a button widget with the given label.
//
// Button returns true if the button has been clicked, otherwise false.
//
// A Button control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
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

func (c *Context) button(label string, opt option, id controlID) (bool, error) {
	res, err := c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		// handle click
		if c.pointing.justPressed() && c.focus == id {
			res = true
		}
		// draw
		c.drawControlFrame(id, bounds, colorButton, opt)
		if len(label) > 0 {
			c.drawControlText(label, bounds, colorText, opt)
		}
		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

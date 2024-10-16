// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import "image"

func (c *Context) Button(label string) Response {
	return c.buttonEx(label, optionAlignCenter)
}

func (c *Context) TextBox(buf *string) Response {
	return c.textBoxEx(buf, 0)
}

func (c *Context) Slider(value *float64, lo, hi float64, step float64, digits int) Response {
	return c.sliderEx(value, lo, hi, step, digits, optionAlignCenter)
}

func (c *Context) Number(value *float64, step float64, digits int) Response {
	return c.numberEx(value, step, digits, optionAlignCenter)
}

func (c *Context) Header(label string, expanded bool) Response {
	var opt option
	if expanded {
		opt |= optionExpanded
	}
	return c.headerEx(label, opt)
}

func (c *Context) TreeNode(label string, f func(res Response)) {
	c.treeNode(label, 0, f)
}

func (c *Context) Window(title string, rect image.Rectangle, f func(res Response, layout Layout)) {
	c.window(title, rect, 0, f)
}

func (c *Context) Panel(name string, f func(layout Layout)) {
	c.panel(name, 0, f)
}

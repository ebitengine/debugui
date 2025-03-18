// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
)

func (c *Context) Button(label string) bool {
	_, result := c.button(label, optionAlignCenter)
	return result
}

func (c *Context) TextBox(buf *string) Response {
	return c.textBox(buf, 0)
}

func (c *Context) Slider(value *float64, lo, hi float64, step float64, digits int) bool {
	return c.slider(value, lo, hi, step, digits, optionAlignCenter)
}

func (c *Context) Number(value *float64, step float64, digits int) bool {
	return c.number(value, step, digits, optionAlignCenter)
}

func (c *Context) Header(label string, expanded bool, f func()) {
	var opt option
	if expanded {
		opt |= optionExpanded
	}
	c.header(label, false, opt, f)
}

func (c *Context) TreeNode(label string, f func()) {
	c.treeNode(label, 0, f)
}

func (c *Context) Window(title string, rect image.Rectangle, f func(layout ContainerLayout)) {
	c.window(title, rect, 0, f)
}

func (c *Context) Panel(name string, f func(layout ContainerLayout)) {
	c.panel(name, 0, f)
}

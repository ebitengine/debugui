// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

func (c *Context) Button(label string) bool {
	_, result := c.button(label, optionAlignCenter)
	return result
}

func (c *Context) Slider(value *float64, lo, hi float64, step float64, digits int) bool {
	return c.slider(value, lo, hi, step, digits, optionAlignCenter)
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

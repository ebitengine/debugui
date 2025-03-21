// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

func (c *Context) Button(label string) bool {
	pc := caller()
	var res bool
	c.wrapError(func() error {
		var err error
		_, res, err = c.button(label, optionAlignCenter, pc)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) Slider(value *float64, lo, hi float64, step float64, digits int) bool {
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.slider(value, lo, hi, step, digits, optionAlignCenter)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) Header(label string, expanded bool, f func()) {
	pc := caller()
	c.wrapError(func() error {
		var opt option
		if expanded {
			opt |= optionExpanded
		}
		if err := c.header(label, false, opt, pc, func() error {
			f()
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) TreeNode(label string, f func()) {
	pc := caller()
	c.wrapError(func() error {
		if err := c.treeNode(label, 0, pc, f); err != nil {
			return err
		}
		return nil
	})
}

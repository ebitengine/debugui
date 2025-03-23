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

// Slider cretes a slider widget with the given int value, range, and step.
//
// lo and hi specify the range of the slider.
//
// Slider returns true if the value of the slider has been changed, otherwise false.
//
// A Slider control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Slider(value *int, lo, hi int, step int) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.slider(value, lo, hi, step, id, optionAlignCenter)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

// SliderF cretes a slider widget with the given float64 value, range, step, and number of digits.
//
// lo and hi specify the range of the slider.
// digits specifies the number of digits to display after the decimal point.
//
// SliderF returns true if the value of the slider has been changed, otherwise false.
//
// A SliderF control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) SliderF(value *float64, lo, hi float64, step float64, digits int) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.sliderF(value, lo, hi, step, digits, id, optionAlignCenter)
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

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"math"
)

// Slider cretes a slider widget with the given int value, range, and step.
//
// lo and hi specify the range of the slider.
//
// Slider returns true if the value of the slider has been changed, otherwise false.
//
// A Slider widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
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
// A SliderF widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
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

func (c *Context) slider(value *int, low, high, step int, id WidgetID, opt option) (bool, error) {
	last := *value
	v := last

	res, err := c.numberTextField(&v, id)
	if err != nil {
		return false, err
	}
	if res {
		*value = v
		return false, nil
	}

	res, err = c.widget(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		if c.focus == id && c.pointing.pressed() {
			if w := bounds.Dx() - defaultStyle.thumbSize; w > 0 {
				v = low + (c.pointingPosition().X-bounds.Min.X)*(high-low)/w
			}
			if step != 0 {
				v = v / step * step
			}
		}
		*value = clamp(v, low, high)
		v = *value
		if last != v {
			res = true
		}

		c.drawWidgetFrame(id, bounds, colorBase, opt)
		w := c.style().thumbSize
		x := int((v - low) * (bounds.Dx() - w) / (high - low))
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawWidgetFrame(id, thumb, colorButton, opt)
		text := fmt.Sprintf("%d", v)
		c.drawWidgetText(text, bounds, colorText, opt)

		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) sliderF(value *float64, low, high, step float64, digits int, id WidgetID, opt option) (bool, error) {
	last := *value
	v := last

	res, err := c.numberTextFieldF(&v, id)
	if err != nil {
		return false, err
	}
	if res {
		*value = v
		return false, nil
	}

	res, err = c.widget(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		if c.focus == id && c.pointing.pressed() {
			if w := float64(bounds.Dx() - defaultStyle.thumbSize); w > 0 {
				v = low + float64(c.pointingPosition().X-bounds.Min.X)*(high-low)/w
			}
			if step != 0 {
				v = math.Round(v/step) * step
			}
		}
		*value = clamp(v, low, high)
		v = *value
		if last != v {
			res = true
		}

		c.drawWidgetFrame(id, bounds, colorBase, opt)
		w := c.style().thumbSize
		x := int((v - low) * float64(bounds.Dx()-w) / (high - low))
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawWidgetFrame(id, thumb, colorButton, opt)
		text := formatNumber(v, digits)
		c.drawWidgetText(text, bounds, colorText, opt)

		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

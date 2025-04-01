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
// Slider returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A Slider widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Slider(value *int, lo, hi int, step int) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.slider(value, lo, hi, step, id, optionAlignCenter)
	})
}

// SliderF cretes a slider widget with the given float64 value, range, step, and number of digits.
//
// lo and hi specify the range of the slider.
// digits specifies the number of digits to display after the decimal point.
//
// SliderF returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A SliderF widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) SliderF(value *float64, lo, hi float64, step float64, digits int) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.sliderF(value, lo, hi, step, digits, id, optionAlignCenter)
	})
}

func (c *Context) slider(value *int, low, high, step int, id WidgetID, opt option) (EventHandler, error) {
	last := *value
	v := last

	if err := c.numberTextField(&v, id); err != nil {
		return nil, err
	}
	if c.numberEdit == id {
		return nil, nil
	}
	*value = v

	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) (EventHandler, error) {
		var e EventHandler
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
			e = &eventHandler{}
		}
		return e, nil
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorBase, opt)
		w := c.style().thumbSize
		x := int((v - low) * (bounds.Dx() - w) / (high - low))
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawWidgetFrame(id, thumb, colorButton, opt)
		text := fmt.Sprintf("%d", v)
		c.drawWidgetText(text, bounds, colorText, opt)
	})
}

func (c *Context) sliderF(value *float64, low, high, step float64, digits int, id WidgetID, opt option) (EventHandler, error) {
	last := *value
	v := last

	if err := c.numberTextFieldF(&v, id); err != nil {
		return nil, err
	}
	if c.numberEdit == id {
		return nil, nil
	}
	*value = v

	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) (EventHandler, error) {
		var e EventHandler
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
			e = &eventHandler{}
		}
		return e, nil
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorBase, opt)
		w := c.style().thumbSize
		x := int((v - low) * float64(bounds.Dx()-w) / (high - low))
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawWidgetFrame(id, thumb, colorButton, opt)
		text := formatNumber(v, digits)
		c.drawWidgetText(text, bounds, colorText, opt)
	})
}

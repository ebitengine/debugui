// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// Slider cretes a slider widget with the given int value, range, and step.
//
// low and high specify the range of the slider.
//
// Slider returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A Slider widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Slider(value *int, low, high int, step int) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.slider(value, low, high, step, id, optionAlignCenter)
	})
}

// SliderF cretes a slider widget with the given float64 value, range, step, and number of digits.
//
// low and high specify the range of the slider.
// digits specifies the number of digits to display after the decimal point.
//
// SliderF returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A SliderF widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) SliderF(value *float64, low, high float64, step float64, digits int) EventHandler {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.sliderF(value, low, high, step, digits, id, optionAlignCenter)
	})
}

func (c *Context) slider(value *int, low, high, step int, id widgetID, opt option) (EventHandler, error) {
	if low > high {
		return nil, fmt.Errorf("debugui: slider low (%d) must be less than or equal to high (%d)", low, high)
	}

	last := *value
	v := last

	if err := c.numberTextField(&v, id); err != nil {
		return nil, err
	}
	if c.numberEdit == id {
		return nil, nil
	}
	*value = v

	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler
		if c.focus == id && c.pointing.pressed() {
			if w := bounds.Dx() - defaultStyle.thumbSize; w > 0 {
				v = low + (c.pointingPosition().X-bounds.Min.X-defaultStyle.thumbSize/2)*(high-low+step)/w
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
		return e
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorBase, opt)
		w := c.style().thumbSize
		var x int
		if low < high {
			x = int((v - low) * (bounds.Dx() - w) / (high - low))
		}
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawWidgetFrame(id, thumb, colorButton, opt)
		text := fmt.Sprintf("%d", v)
		c.drawWidgetText(text, bounds, colorText, opt)
	})
}

func (c *Context) sliderF(value *float64, low, high, step float64, digits int, id widgetID, opt option) (EventHandler, error) {
	if low > high {
		return nil, fmt.Errorf("debugui: slider low (%f) must be less than or equal to high (%f)", low, high)
	}

	last := *value
	v := last

	if err := c.numberTextFieldF(&v, id); err != nil {
		return nil, err
	}
	if c.numberEdit == id {
		return nil, nil
	}
	*value = v

	return c.widget(id, opt, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler
		if c.focus == id && c.pointing.pressed() {
			if w := float64(bounds.Dx() - defaultStyle.thumbSize); w > 0 {
				v = low + float64(c.pointingPosition().X-bounds.Min.X-defaultStyle.thumbSize/2)*(high-low+step)/w
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
		return e
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorBase, opt)
		w := c.style().thumbSize
		var x int
		if low < high {
			x = int((v - low) * float64(bounds.Dx()-w) / (high - low))
		}
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawWidgetFrame(id, thumb, colorButton, opt)
		text := formatNumber(v, digits)
		c.drawWidgetText(text, bounds, colorText, opt)
	})
}

func (c *Context) numberTextField(value *int, id widgetID) error {
	if c.pointing.justPressed() && ebiten.IsKeyPressed(ebiten.KeyShift) && c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf("%d", *value)
	}
	if c.numberEdit == id {
		e, err := c.textFieldRaw(&c.numberEditBuf, id, optionAlignRight)
		if err != nil {
			return err
		}
		if e != nil {
			e.On(func() {
				nval, err := strconv.ParseInt(c.numberEditBuf, 10, 64)
				if err != nil {
					nval = 0
				}
				*value = int(nval)
				c.numberEdit = widgetID{}
			})
		}
	}
	return nil
}

func (c *Context) numberTextFieldF(value *float64, id widgetID) error {
	if c.pointing.justPressed() && ebiten.IsKeyPressed(ebiten.KeyShift) && c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.numberEdit == id {
		e, err := c.textFieldRaw(&c.numberEditBuf, id, optionAlignRight)
		if err != nil {
			return err
		}
		if e != nil {
			e.On(func() {
				nval, err := strconv.ParseFloat(c.numberEditBuf, 64)
				if err != nil {
					nval = 0
				}
				*value = float64(nval)
				c.numberEdit = widgetID{}
			})
		}
	}
	return nil
}

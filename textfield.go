// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

// TextField creates a text field to modify the value of a string buf.
//
// TextField returns true when this TextField is unfocused or the user pressed Enter, otherwise false.
//
// A TextField control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) TextField(buf *string) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.textField(buf, id, 0)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) textFieldRaw(buf *string, id controlID, opt option) (bool, error) {
	res, err := c.control(id, opt|optionHoldFocus, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool

		f := c.currentContainer().textInputTextField(id)
		if c.focus == id {
			// handle text input
			f.Focus()
			x := bounds.Min.X + c.style().padding + textWidth(*buf)
			y := bounds.Min.Y + lineHeight()
			handled, err := f.HandleInput(x, y)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return false, nil
			}
			if *buf != f.Text() {
				*buf = f.Text()
			}

			if !handled {
				if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(*buf) > 0 {
					_, size := utf8.DecodeLastRuneInString(*buf)
					*buf = (*buf)[:len(*buf)-size]
					f.SetTextAndSelection(*buf, len(*buf), len(*buf))
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
					res = true
				}
			}
		} else {
			if *buf != f.Text() {
				f.SetTextAndSelection(*buf, len(*buf), len(*buf))
			}
			if wasFocused {
				res = true
			}
		}

		// draw
		c.drawControlFrame(id, bounds, colorBase, opt)
		if c.focus == id {
			color := c.style().colors[colorText]
			textw := textWidth(*buf)
			texth := lineHeight()
			ofx := bounds.Dx() - c.style().padding - textw - 1
			textx := bounds.Min.X + min(ofx, c.style().padding)
			texty := bounds.Min.Y + (bounds.Dy()-texth)/2
			c.pushClipRect(bounds)
			c.drawText(f.TextForRendering(), image.Pt(textx, texty), color)
			c.drawRect(image.Rect(textx+textw, texty, textx+textw+1, texty+texth), color)
			c.popClipRect()
		} else {
			c.drawControlText(*buf, bounds, colorText, opt)
		}
		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil

}

// SetTextFieldValue sets the value of the last text field with the given value.
//
// If there is no text field, SetTextFieldValue does nothing.
func (c *Context) SetTextFieldValue(value string) {
	id := c.lastTextFieldID
	if id == emptyControlID {
		return
	}
	c.wrapError(func() error {
		f := c.currentContainer().textInputTextField(id)
		f.SetTextAndSelection(value, 0, 0)
		return nil
	})
}

func (c *Context) textField(buf *string, id controlID, opt option) (bool, error) {
	c.lastTextFieldID = id
	res, err := c.textFieldRaw(buf, id, opt)
	if err != nil {
		return false, err
	}
	return res, nil
}

// NumberField creates a number field to modify the value of a int value.
//
// step is the amount to increment or decrement the value when the user drags the thumb.
//
// NumberField returns true when the value has been changed, otherwise false.
//
// A NumberField control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberField(value *int, step int) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.numberField(value, step, id, optionAlignCenter)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

// NumberFieldF creates a number field to modify the value of a float64 value.
//
// step is the amount to increment or decrement the value when the user drags the thumb.
// digits is the number of decimal places to display.
//
// NumberFieldF returns true when the value has been changed, otherwise false.
//
// A NumberFieldF control is uniquely determined by its call location.
// Function calls made in different locations will create different controls.
// If you want to generate different controls with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberFieldF(value *float64, step float64, digits int) bool {
	pc := caller()
	id := c.idFromCaller(pc)
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.numberFieldF(value, step, digits, id, optionAlignCenter)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) numberField(value *int, step int, id controlID, opt option) (bool, error) {
	last := *value

	res, err := c.numberTextField(value, id)
	if err != nil {
		return false, err
	}
	if res {
		return false, nil
	}

	// handle normal mode
	res, err = c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		if c.focus == id && c.pointing.pressed() {
			*value += (c.pointingDelta().X) * step
		}
		if *value != last {
			res = true
		}

		c.drawControlFrame(id, bounds, colorBase, opt)
		text := fmt.Sprintf("%d", *value)
		c.drawControlText(text, bounds, colorText, opt)

		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) numberFieldF(value *float64, step float64, digits int, id controlID, opt option) (bool, error) {
	last := *value

	res, err := c.numberTextFieldF(value, id)
	if err != nil {
		return false, err
	}
	if res {
		return false, nil
	}

	// handle normal mode
	res, err = c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		if c.focus == id && c.pointing.pressed() {
			*value += float64(c.pointingDelta().X) * step
		}
		if *value != last {
			res = true
		}

		c.drawControlFrame(id, bounds, colorBase, opt)
		text := formatNumber(*value, digits)
		c.drawControlText(text, bounds, colorText, opt)

		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) numberTextField(value *int, id controlID) (bool, error) {
	if c.pointing.justPressed() && ebiten.IsKeyPressed(ebiten.KeyShift) &&
		c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf("%d", *value)
	}
	if c.numberEdit == id {
		res, err := c.textFieldRaw(&c.numberEditBuf, id, 0)
		if err != nil {
			return false, err
		}
		if res || c.focus != id {
			nval, err := strconv.ParseInt(c.numberEditBuf, 10, 64)
			if err != nil {
				nval = 0
			}
			*value = int(nval)
			c.numberEdit = emptyControlID
		}
		return true, nil
	}
	return false, nil
}

func (c *Context) numberTextFieldF(value *float64, id controlID) (bool, error) {
	if c.pointing.justPressed() && ebiten.IsKeyPressed(ebiten.KeyShift) &&
		c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.numberEdit == id {
		res, err := c.textFieldRaw(&c.numberEditBuf, id, 0)
		if err != nil {
			return false, err
		}
		if res || c.focus != id {
			nval, err := strconv.ParseFloat(c.numberEditBuf, 64)
			if err != nil {
				nval = 0
			}
			*value = float64(nval)
			c.numberEdit = emptyControlID
		}
		return true, nil
	}
	return false, nil
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

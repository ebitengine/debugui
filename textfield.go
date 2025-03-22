// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

func (c *Context) TextField(buf *string) bool {
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.textField(buf, 0)
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

func (c *Context) SetTextFieldValue(value *string) {
	c.wrapError(func() error {
		id := c.idFromPointer(unsafe.Pointer(value))
		f := c.currentContainer().textInputTextField(id)
		f.SetTextAndSelection(*value, 0, 0)
		return nil
	})
}

func (c *Context) textField(buf *string, opt option) (bool, error) {
	id := c.idFromPointer(unsafe.Pointer(buf))
	res, err := c.textFieldRaw(buf, id, opt)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) NumberField(value *float64, step float64, digits int) bool {
	var res bool
	c.wrapError(func() error {
		var err error
		res, err = c.numberField(value, step, digits, optionAlignCenter)
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) numberField(value *float64, step float64, digits int, opt option) (bool, error) {
	id := c.idFromPointer(unsafe.Pointer(value))
	last := *value

	// handle text input mode
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
		// handle input
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			*value += float64(c.mouseDelta().X) * step
		}
		// set flag if value changed
		if *value != last {
			res = true
		}

		// draw base
		c.drawControlFrame(id, bounds, colorBase, opt)
		// draw text
		text := formatNumber(*value, digits)
		c.drawControlText(text, bounds, colorText, opt)

		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) numberTextField(value *float64, id controlID) (bool, error) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && ebiten.IsKeyPressed(ebiten.KeyShift) &&
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
			nval, err := strconv.ParseFloat(c.numberEditBuf, 32)
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

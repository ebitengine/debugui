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
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

func (c *Context) TextField(buf *string) bool {
	return c.textField(buf, 0)
}

func (c *Context) textInputTextField(id controlID) *textinput.Field {
	if id == 0 {
		return nil
	}
	if _, ok := c.textInputTextFields[id]; !ok {
		if c.textInputTextFields == nil {
			c.textInputTextFields = make(map[controlID]*textinput.Field)
		}
		// TODO: Remove unused fields.
		c.textInputTextFields[id] = &textinput.Field{}
	}
	return c.textInputTextFields[id]
}

func (c *Context) textFieldRaw(buf *string, id controlID, opt option) bool {
	return c.control(id, opt|optionHoldFocus, func(bounds image.Rectangle, wasFocused bool) bool {
		var res bool

		f := c.textInputTextField(id)
		if c.focus == id {
			// handle text input
			f.Focus()
			x := bounds.Min.X + c.style().padding + textWidth(*buf)
			y := bounds.Min.Y + lineHeight()
			handled, err := f.HandleInput(x, y)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return false
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
		c.drawControlFrame(id, bounds, ColorBase, opt)
		if c.focus == id {
			color := c.style().colors[ColorText]
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
			c.drawControlText(*buf, bounds, ColorText, opt)
		}
		return res
	})
}

func (c *Context) SetTextFieldValue(value *string) {
	id := c.idFromGlobalUniquePointer(unsafe.Pointer(value))
	f := c.textInputTextField(id)
	f.SetTextAndSelection(*value, 0, 0)
}

func (c *Context) textField(buf *string, opt option) bool {
	id := c.idFromGlobalUniquePointer(unsafe.Pointer(buf))
	return c.textFieldRaw(buf, id, opt)
}

func (c *Context) NumberField(value *float64, step float64, digits int) bool {
	return c.numberField(value, step, digits, optionAlignCenter)
}

func (c *Context) numberField(value *float64, step float64, digits int, opt option) bool {
	id := c.idFromGlobalUniquePointer(unsafe.Pointer(value))
	last := *value

	// handle text input mode
	if c.numberTextField(value, id) {
		return false
	}

	// handle normal mode
	return c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) bool {
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
		c.drawControlFrame(id, bounds, ColorBase, opt)
		// draw text
		text := formatNumber(*value, digits)
		c.drawControlText(text, bounds, ColorText, opt)

		return res
	})
}

func (c *Context) numberTextField(value *float64, id controlID) bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && ebiten.IsKeyPressed(ebiten.KeyShift) &&
		c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.numberEdit == id {
		res := c.textFieldRaw(&c.numberEditBuf, id, 0)
		if res || c.focus != id {
			nval, err := strconv.ParseFloat(c.numberEditBuf, 32)
			if err != nil {
				nval = 0
			}
			*value = float64(nval)
			c.numberEdit = 0
		}
		return true
	}
	return false
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

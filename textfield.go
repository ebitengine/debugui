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
// TextField returns an EventHandler to handle events when the value is confirmed, such as on blur or Enter key press.
// A returned EventHandler is never nil.
//
// A TextField widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) TextField(buf *string) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.textField(buf, id, 0)
	})
}

func (c *Context) textFieldRaw(buf *string, id WidgetID, opt option) (EventHandler, error) {
	return c.widget(id, opt|optionHoldFocus, func(bounds image.Rectangle, wasFocused bool) (EventHandler, error) {
		var e EventHandler

		f := c.currentContainer().textInputTextField(id, true)
		if c.focus == id {
			// handle text input
			f.Focus()
			x := bounds.Min.X + c.style().padding + textWidth(*buf)
			y := bounds.Min.Y + lineHeight()
			handled, err := f.HandleInput(x, y)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil, nil
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
					e = &eventHandler{}
				}
			}
		} else {
			if *buf != f.Text() {
				f.SetTextAndSelection(*buf, len(*buf), len(*buf))
			}
			if wasFocused {
				e = &eventHandler{}
			}
		}

		// draw
		c.drawWidgetFrame(id, bounds, colorBase, opt)
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
			c.drawWidgetText(*buf, bounds, colorText, opt)
		}
		return e, nil
	})
}

// SetTextFieldValue sets the value of the text field with the given widgetID.
//
// If widgetID is not for a text field, this function does nothing.
func (c *Context) SetTextFieldValue(widgetID WidgetID, value string) {
	if widgetID == emptyWidgetID {
		return
	}
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if f := c.currentContainer().textInputTextField(widgetID, false); f != nil {
			f.SetTextAndSelection(value, 0, 0)
		}
		return nil, nil
	})
}

func (c *Context) textField(buf *string, id WidgetID, opt option) (EventHandler, error) {
	return c.textFieldRaw(buf, id, opt)
}

// NumberField creates a number field to modify the value of a int value.
//
// step is the amount to increment or decrement the value when the user drags the thumb.
//
// NumberField returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A NumberField widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberField(value *int, step int) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.numberField(value, step, id, optionAlignCenter)
	})
}

// NumberFieldF creates a number field to modify the value of a float64 value.
//
// step is the amount to increment or decrement the value when the user drags the thumb.
// digits is the number of decimal places to display.
//
// NumberFieldF returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A NumberFieldF widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberFieldF(value *float64, step float64, digits int) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.numberFieldF(value, step, digits, id, optionAlignCenter)
	})
}

func (c *Context) numberField(value *int, step int, id WidgetID, opt option) (EventHandler, error) {
	last := *value

	if err := c.numberTextField(value, id); err != nil {
		return nil, err
	}
	if c.numberEdit == id {
		return nil, nil
	}

	// handle normal mode
	return c.widget(id, opt, func(bounds image.Rectangle, wasFocused bool) (EventHandler, error) {
		var e EventHandler
		if c.focus == id && c.pointing.pressed() {
			*value += (c.pointingDelta().X) * step
		}
		if *value != last {
			e = &eventHandler{}
		}

		c.drawWidgetFrame(id, bounds, colorBase, opt)
		text := fmt.Sprintf("%d", *value)
		c.drawWidgetText(text, bounds, colorText, opt)

		return e, nil
	})
}

func (c *Context) numberFieldF(value *float64, step float64, digits int, id WidgetID, opt option) (EventHandler, error) {
	last := *value

	if err := c.numberTextFieldF(value, id); err != nil {
		return nil, err
	}
	if c.numberEdit == id {
		return nil, nil
	}

	// handle normal mode
	return c.widget(id, opt, func(bounds image.Rectangle, wasFocused bool) (EventHandler, error) {
		var e EventHandler
		if c.focus == id && c.pointing.pressed() {
			*value += float64(c.pointingDelta().X) * step
		}
		if *value != last {
			e = &eventHandler{}
		}

		c.drawWidgetFrame(id, bounds, colorBase, opt)
		text := formatNumber(*value, digits)
		c.drawWidgetText(text, bounds, colorText, opt)

		return e, nil
	})
}

func (c *Context) numberTextField(value *int, id WidgetID) error {
	if c.pointing.justPressed() && ebiten.IsKeyPressed(ebiten.KeyShift) && c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf("%d", *value)
	}
	if c.numberEdit == id {
		e, err := c.textFieldRaw(&c.numberEditBuf, id, 0)
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
				c.numberEdit = emptyWidgetID
			})
		}
	}
	return nil
}

func (c *Context) numberTextFieldF(value *float64, id WidgetID) error {
	if c.pointing.justPressed() && ebiten.IsKeyPressed(ebiten.KeyShift) && c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.numberEdit == id {
		e, err := c.textFieldRaw(&c.numberEditBuf, id, 0)
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
				c.numberEdit = emptyWidgetID
			})
		}
	}
	return nil
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

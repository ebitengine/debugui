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
	id := c.idStack.push(idPartFromCaller(pc))
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.textField(buf, id, 0)
	})
}

func (c *Context) textFieldRaw(buf *string, id widgetID, opt option) (EventHandler, error) {
	return c.widget(id, opt|optionHoldFocus, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
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
				return nil
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
		return e
	}, func(bounds image.Rectangle) {
		c.drawWidgetFrame(id, bounds, colorBase, opt)
		if c.focus == id {
			f := c.currentContainer().textInputTextField(id, true)

			color := c.style().colors[colorText]
			textw := textWidth(*buf)
			texth := lineHeight()
			ofx := bounds.Dx() - c.style().padding - textw - 1
			textx := bounds.Min.X + min(ofx, c.style().padding)
			switch {
			case opt&optionAlignCenter != 0:
				textx = bounds.Min.X + (bounds.Dx()-textw)/2
			case opt&optionAlignRight != 0:
				textx = bounds.Min.X + bounds.Dx() - textw - c.style().padding
			}
			texty := bounds.Min.Y + (bounds.Dy()-texth)/2
			c.pushClipRect(bounds)
			c.drawText(f.TextForRendering(), image.Pt(textx, texty), color)
			c.drawRect(image.Rect(textx+textw, texty, textx+textw+1, texty+texth), color)
			c.popClipRect()
		} else {
			c.drawWidgetText(*buf, bounds, colorText, opt)
		}
	})
}

// SetTextFieldValue sets the value of the current text field.
//
// If the last widget is not a text field, this function does nothing.
func (c *Context) SetTextFieldValue(value string) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if f := c.currentContainer().textInputTextField(c.currentID, false); f != nil {
			f.SetTextAndSelection(value, 0, 0)
		}
		return nil, nil
	})
}

func (c *Context) textField(buf *string, id widgetID, opt option) (EventHandler, error) {
	return c.textFieldRaw(buf, id, opt)
}

// NumberField creates a number field to modify the value of a int value.
//
// step is the amount to increment or decrement the value when the user drags the mouse cursor.
//
// NumberField returns an EventHandler to handle value change events.
// A returned EventHandler is never nil.
//
// A NumberField widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) NumberField(value *int, step int) EventHandler {
	pc := caller()
	idPart := idPartFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.numberField(value, step, idPart, optionAlignRight)
	})
}

// NumberFieldF creates a number field to modify the value of a float64 value.
//
// step is the amount to increment or decrement the value when the user drags the mouse cursor.
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
	idPart := idPartFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.numberFieldF(value, step, digits, idPart, optionAlignRight)
	})
}

func (c *Context) numberField(value *int, step int, idPart string, opt option) (EventHandler, error) {
	last := *value

	var e EventHandler
	var err error
	c.idScopeFromIDPart(idPart, func(id widgetID) {
		c.GridCell(func(bounds image.Rectangle) {
			c.SetGridLayout([]int{-1, lineHeight()}, nil)

			buf := fmt.Sprintf("%d", *value)
			e1, err1 := c.textFieldRaw(&buf, id, opt)
			if err1 != nil {
				err = err1
				return
			}
			if e1 != nil {
				e1.On(func() {
					c.setFocus(widgetID{})
					v, err := strconv.ParseInt(buf, 10, 64)
					if err != nil {
						v = 0
					}
					*value = int(v)
					if *value != last {
						e = &eventHandler{}
					}
				})
			}
			if c.focus == id {
				var updated bool
				if keyRepeated(ebiten.KeyUp) || keyRepeated(ebiten.KeyDown) {
					v, err := strconv.ParseInt(buf, 10, 64)
					if err != nil {
						v = 0
					}
					*value = int(v)
					updated = true
					if keyRepeated(ebiten.KeyUp) {
						*value += step
					}
					if keyRepeated(ebiten.KeyDown) {
						*value -= step
						updated = true
					}
				}
				if updated {
					buf := fmt.Sprintf("%d", *value)
					if f := c.currentContainer().textInputTextField(id, false); f != nil {
						f.SetTextAndSelection(buf, len(buf), len(buf))
					}
					e = &eventHandler{}
				}
			}

			c.GridCell(func(bounds image.Rectangle) {
				c.SetGridLayout(nil, []int{-1, -1})
				up, down := c.spinButtons(id)
				up.On(func() {
					*value += step
					e = &eventHandler{}
				})
				down.On(func() {
					*value -= step
					e = &eventHandler{}
				})
			})
		})
	})

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (c *Context) numberFieldF(value *float64, step float64, digits int, idPart string, opt option) (EventHandler, error) {
	last := *value

	var e EventHandler
	var err error
	c.idScopeFromIDPart(idPart, func(id widgetID) {
		c.GridCell(func(bounds image.Rectangle) {
			c.SetGridLayout([]int{-1, lineHeight()}, nil)

			buf := formatNumber(*value, digits)
			e1, err1 := c.textFieldRaw(&buf, id, opt)
			if err1 != nil {
				err = err1
				return
			}
			if e1 != nil {
				e1.On(func() {
					c.setFocus(widgetID{})
					v, err := strconv.ParseFloat(buf, 64)
					if err != nil {
						v = 0
					}
					*value = float64(v)
					if *value != last {
						e = &eventHandler{}
					}
				})
			}
			if c.focus == id {
				var updated bool
				if keyRepeated(ebiten.KeyUp) || keyRepeated(ebiten.KeyDown) {
					v, err := strconv.ParseFloat(buf, 64)
					if err != nil {
						v = 0
					}
					*value = float64(v)
					updated = true
					if keyRepeated(ebiten.KeyUp) {
						*value += step
					}
					if keyRepeated(ebiten.KeyDown) {
						*value -= step
						updated = true
					}
				}
				if updated {
					buf := formatNumber(*value, digits)
					if f := c.currentContainer().textInputTextField(id, false); f != nil {
						f.SetTextAndSelection(buf, len(buf), len(buf))
					}
					e = &eventHandler{}
				}
			}

			c.GridCell(func(bounds image.Rectangle) {
				c.SetGridLayout(nil, []int{-1, -1})
				up, down := c.spinButtons(id)
				up.On(func() {
					*value += step
					e = &eventHandler{}
				})
				down.On(func() {
					*value -= step
					e = &eventHandler{}
				})
			})
		})
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

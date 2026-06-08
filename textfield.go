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
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

// textFieldState drives IME input for a single text field via a [textinput.Composer].
//
// The caret is fixed at the end of the text, so the committed text is entirely before
// the caret and any active composition is appended after it.
type textFieldState struct {
	composer textinput.Composer

	// committedText is the confirmed text, excluding any in-progress IME composition.
	// While the field is focused it is the source of truth and is synced to the caller's
	// buffer; otherwise it mirrors the buffer.
	committedText string

	// caretBounds is the caret rectangle handed to the IME to position the candidate window.
	caretBounds image.Rectangle

	// composition is the active IME preedit text.
	composition string
}

func newTextFieldState() *textFieldState {
	t := &textFieldState{}
	t.composer.OnNewSession = t.onNewSession
	t.composer.OnComposition = t.onComposition
	t.composer.OnCommit = t.onCommit
	return t
}

func (t *textFieldState) onNewSession() *textinput.SessionOptions {
	return &textinput.SessionOptions{
		CaretBounds:     t.caretBounds,
		TextBeforeCaret: t.committedText,
	}
}

func (t *textFieldState) onComposition(c *textinput.Composition) {
	t.composition = c.Text()
}

func (t *textFieldState) onCommit(c *textinput.Commit) {
	if before, after := c.IsSurroundingTextReplaced(); before || after {
		newBefore, newAfter := c.SurroundingText()
		t.committedText = newBefore + c.Text() + newAfter
		return
	}
	t.committedText += c.Text()
}

// update positions the caret and runs one tick of IME handling.
// It reports whether the IME consumed input this tick; if so, the caller must not process further key input.
func (t *textFieldState) update(caretBounds image.Rectangle) (handled bool, err error) {
	t.caretBounds = caretBounds
	return t.composer.Update()
}

// text returns the committed text, excluding any in-progress composition.
func (t *textFieldState) text() string {
	return t.committedText
}

// textForRendering returns the committed text with the active composition appended at the caret.
func (t *textFieldState) textForRendering() string {
	return t.committedText + t.composition
}

// setText replaces the committed text and abandons any in-progress composition.
func (t *textFieldState) setText(text string) {
	t.composer.Finish()
	t.committedText = text
	t.composition = ""
}

// deleteBackward removes the last rune from the committed text.
func (t *textFieldState) deleteBackward() {
	if len(t.committedText) == 0 {
		return
	}
	_, size := utf8.DecodeLastRuneInString(t.committedText)
	t.committedText = t.committedText[:len(t.committedText)-size]
}

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
			// While focused, f's committed text is the source of truth. The caret is fixed at the end of the text.
			x := bounds.Min.X + c.style().padding + textWidth(f.text())
			y := bounds.Min.Y + lineHeight()
			handled, err := f.update(image.Rect(x, y, x+1, y+1))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}

			if !handled {
				if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
					f.deleteBackward()
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
					e = &eventHandler{}
				}
			}

			*buf = f.text()
		} else {
			// The buffer is the source of truth while unfocused. setText also ends any IME session.
			f.setText(*buf)
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
			textw := textWidth(f.text())
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
			c.drawText(f.textForRendering(), image.Pt(textx, texty), color)
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
			f.setText(value)
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
						f.setText(buf)
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
						f.setText(buf)
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

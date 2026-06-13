// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"unicode"
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
// The selection is a byte range in committedText, tracked as an anchor and a caret.
// While an IME session is active, the committed text and the selection must not change
// except through onCommit; every external mutation ends the session first so that the
// IME's view of the surrounding text stays in sync.
type textFieldState struct {
	composer textinput.Composer

	// committedText is the confirmed text, excluding any in-progress IME composition.
	// While the field is focused it is the source of truth and is synced to the caller's
	// buffer; otherwise it mirrors the buffer.
	committedText string

	// selectionAnchor and selectionCaret are byte offsets in committedText delimiting
	// the selection. The anchor is the fixed end and the caret is the end that moves
	// when the selection is extended. They are equal when the selection is empty.
	selectionAnchor int
	selectionCaret  int

	// caretBounds is the caret rectangle handed to the IME to position the candidate window.
	caretBounds image.Rectangle

	// composition is the active IME preedit text. It is rendered in place of the selection.
	composition string

	// compositionSelStart and compositionSelEnd are the IME-side selection within
	// composition, as byte offsets.
	compositionSelStart int
	compositionSelEnd   int

	// dragging reports whether a pointer drag that started on this field is extending
	// the selection.
	dragging bool

	// lastClickTick is the tick of the most recent pointer press on this field. Together
	// with lastClickPos and consecutiveClicks it distinguishes single-, double-, and
	// triple-clicks.
	lastClickTick int64

	// lastClickPos is the byte offset of the most recent pointer press. A press at a
	// different offset restarts the sequence so that it can begin a fresh drag-selection.
	lastClickPos int

	// consecutiveClicks counts pointer presses arriving in quick succession at the same
	// position: 1 for a single click, 2 for a double-click, 3 for a triple-click.
	consecutiveClicks int

	// scrollX is the horizontal scroll offset in pixels, used when the text is wider
	// than the field.
	scrollX int
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
		TextBeforeCaret: t.committedText[:t.selectionStart()],
		TextAfterCaret:  t.committedText[t.selectionEnd():],
	}
}

func (t *textFieldState) onComposition(c *textinput.Composition) {
	t.composition = c.Text()
	t.compositionSelStart, t.compositionSelEnd = c.SelectionRangeInBytes()
}

// onCommit applies a committed IME state. The committed text replaces the selection,
// and the selection collapses to the caret after the committed text.
func (t *textFieldState) onCommit(c *textinput.Commit) {
	before := t.committedText[:t.selectionStart()]
	after := t.committedText[t.selectionEnd():]
	if replacedBefore, replacedAfter := c.IsSurroundingTextReplaced(); replacedBefore || replacedAfter {
		before, after = c.SurroundingText()
	}
	t.committedText = before + c.Text() + after
	t.selectionAnchor = len(before) + len(c.Text())
	t.selectionCaret = t.selectionAnchor
}

// selectionStart returns the smaller byte offset of the selection in the committed text.
func (t *textFieldState) selectionStart() int {
	return min(t.selectionAnchor, t.selectionCaret)
}

// selectionEnd returns the larger byte offset of the selection in the committed text.
func (t *textFieldState) selectionEnd() int {
	return max(t.selectionAnchor, t.selectionCaret)
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

// textForRendering returns the committed text with the active composition spliced in
// at the selection, which the composition visually replaces.
func (t *textFieldState) textForRendering() string {
	if t.composition == "" {
		return t.committedText
	}
	return t.committedText[:t.selectionStart()] + t.composition + t.committedText[t.selectionEnd():]
}

// caretRenderIndex returns the caret position as a byte offset in the text returned
// by textForRendering.
func (t *textFieldState) caretRenderIndex() int {
	if t.composition != "" {
		return t.selectionStart() + t.compositionSelStart
	}
	return t.selectionCaret
}

// setText replaces the committed text and abandons any in-progress composition.
// The caret moves to the end of the text.
func (t *textFieldState) setText(text string) {
	t.composer.Finish()
	t.committedText = text
	t.composition = ""
	t.selectionAnchor = len(text)
	t.selectionCaret = len(text)
	t.dragging = false
	t.scrollX = 0
}

// moveCaretTo moves the caret to the byte offset pos in the committed text, ending
// any IME session. When extend is true, the selection anchor is kept so that the
// selection extends; otherwise the selection collapses to the caret.
func (t *textFieldState) moveCaretTo(pos int, extend bool) {
	pos = min(max(pos, 0), len(t.committedText))
	if pos == t.selectionCaret && (extend || pos == t.selectionAnchor) {
		return
	}
	t.composer.Finish()
	t.selectionCaret = pos
	if !extend {
		t.selectionAnchor = pos
	}
}

// moveCaretLeft moves the caret one rune left, extending the selection when extend
// is true. When a selection exists and extend is false, the selection collapses to
// its start instead.
func (t *textFieldState) moveCaretLeft(extend bool) {
	if !extend && t.selectionAnchor != t.selectionCaret {
		t.moveCaretTo(t.selectionStart(), false)
		return
	}
	if t.selectionCaret == 0 {
		return
	}
	_, size := utf8.DecodeLastRuneInString(t.committedText[:t.selectionCaret])
	t.moveCaretTo(t.selectionCaret-size, extend)
}

// moveCaretRight moves the caret one rune right, extending the selection when extend
// is true. When a selection exists and extend is false, the selection collapses to
// its end instead.
func (t *textFieldState) moveCaretRight(extend bool) {
	if !extend && t.selectionAnchor != t.selectionCaret {
		t.moveCaretTo(t.selectionEnd(), false)
		return
	}
	if t.selectionCaret >= len(t.committedText) {
		return
	}
	_, size := utf8.DecodeRuneInString(t.committedText[t.selectionCaret:])
	t.moveCaretTo(t.selectionCaret+size, extend)
}

// isWordRune reports whether r is part of a word for the purpose of word-wise
// caret movement. Letters, digits, and underscores are word runes; everything
// else (whitespace, punctuation) is a separator.
func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// wordBoundaryLeft returns the byte offset of the start of the word at or before pos:
// it skips separators immediately before pos, then the word runes before that.
func wordBoundaryLeft(s string, pos int) int {
	for pos > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:pos])
		if isWordRune(r) {
			break
		}
		pos -= size
	}
	for pos > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:pos])
		if !isWordRune(r) {
			break
		}
		pos -= size
	}
	return pos
}

// wordBoundaryRight returns the byte offset of the end of the word at or after pos:
// it skips separators immediately after pos, then the word runes after that.
func wordBoundaryRight(s string, pos int) int {
	for pos < len(s) {
		r, size := utf8.DecodeRuneInString(s[pos:])
		if isWordRune(r) {
			break
		}
		pos += size
	}
	for pos < len(s) {
		r, size := utf8.DecodeRuneInString(s[pos:])
		if !isWordRune(r) {
			break
		}
		pos += size
	}
	return pos
}

// wordRangeAt returns the byte range [start, end) of the word at pos: the maximal run
// of word runes (see isWordRune) containing pos, or the maximal run of non-word runes
// when pos is adjacent only to non-word runes. It is used to select a word on a
// double-click.
func wordRangeAt(s string, pos int) (start, end int) {
	pos = min(max(pos, 0), len(s))
	// A click adjacent to a word rune (on either side) selects that word; otherwise it
	// selects the run of non-word runes (such as spaces) under the click.
	var targetWord bool
	if pos < len(s) {
		r, _ := utf8.DecodeRuneInString(s[pos:])
		targetWord = isWordRune(r)
	}
	if !targetWord && pos > 0 {
		r, _ := utf8.DecodeLastRuneInString(s[:pos])
		targetWord = isWordRune(r)
	}
	start = pos
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:start])
		if isWordRune(r) != targetWord {
			break
		}
		start -= size
	}
	end = pos
	for end < len(s) {
		r, size := utf8.DecodeRuneInString(s[end:])
		if isWordRune(r) != targetWord {
			break
		}
		end += size
	}
	return start, end
}

// moveCaretWordLeft moves the caret to the start of the previous word, extending
// the selection when extend is true.
func (t *textFieldState) moveCaretWordLeft(extend bool) {
	t.moveCaretTo(wordBoundaryLeft(t.committedText, t.selectionCaret), extend)
}

// moveCaretWordRight moves the caret to the end of the next word, extending the
// selection when extend is true.
func (t *textFieldState) moveCaretWordRight(extend bool) {
	t.moveCaretTo(wordBoundaryRight(t.committedText, t.selectionCaret), extend)
}

// selectAll selects the whole committed text, ending any IME session.
func (t *textFieldState) selectAll() {
	if t.selectionAnchor == 0 && t.selectionCaret == len(t.committedText) {
		return
	}
	t.composer.Finish()
	t.selectionAnchor = 0
	t.selectionCaret = len(t.committedText)
}

// selectWordAt selects the word at the byte offset pos, ending any IME session.
func (t *textFieldState) selectWordAt(pos int) {
	t.composer.Finish()
	t.selectionAnchor, t.selectionCaret = wordRangeAt(t.committedText, pos)
}

// handleClick processes a pointer press at the byte offset pos; now is the current tick.
// Presses within interval ticks of one another at the same position escalate the
// selection: a single click places the caret (extending the selection when extend is
// true), a double-click selects the word at pos, and a triple-click selects the whole
// text. A press at a different position restarts the sequence as a single click, so that
// it can begin a fresh drag-selection.
func (t *textFieldState) handleClick(pos int, extend bool, now, interval int64) {
	count := 1
	if now-t.lastClickTick <= interval && pos == t.lastClickPos {
		count = min(t.consecutiveClicks+1, 3)
	}
	t.consecutiveClicks = count
	t.lastClickTick = now
	t.lastClickPos = pos

	switch count {
	case 2:
		t.selectWordAt(pos)
		t.dragging = false
	case 3:
		t.selectAll()
		t.dragging = false
	default:
		t.composer.Finish()
		t.moveCaretTo(pos, extend)
		t.dragging = true
	}
}

// deleteSelection removes the selected text, ending any IME session.
// It reports whether a non-empty selection was removed.
func (t *textFieldState) deleteSelection() bool {
	start, end := t.selectionStart(), t.selectionEnd()
	if start == end {
		return false
	}
	t.composer.Finish()
	t.committedText = t.committedText[:start] + t.committedText[end:]
	t.selectionAnchor = start
	t.selectionCaret = start
	return true
}

// deleteBackward removes the selected text, or the rune before the caret when the
// selection is empty.
func (t *textFieldState) deleteBackward() {
	if t.deleteSelection() {
		return
	}
	if t.selectionCaret == 0 {
		return
	}
	t.composer.Finish()
	_, size := utf8.DecodeLastRuneInString(t.committedText[:t.selectionCaret])
	t.committedText = t.committedText[:t.selectionCaret-size] + t.committedText[t.selectionCaret:]
	t.selectionCaret -= size
	t.selectionAnchor = t.selectionCaret
}

// deleteForward removes the selected text, or the rune after the caret when the
// selection is empty.
func (t *textFieldState) deleteForward() {
	if t.deleteSelection() {
		return
	}
	if t.selectionCaret >= len(t.committedText) {
		return
	}
	t.composer.Finish()
	_, size := utf8.DecodeRuneInString(t.committedText[t.selectionCaret:])
	t.committedText = t.committedText[:t.selectionCaret] + t.committedText[t.selectionCaret+size:]
}

// textIndexFromX returns the byte offset of the inter-rune boundary in str closest
// to the x position in pixels, measured from the start of str.
func textIndexFromX(str string, x int) int {
	var idx int
	for idx < len(str) {
		_, size := utf8.DecodeRuneInString(str[idx:])
		w0 := textWidth(str[:idx])
		w1 := textWidth(str[:idx+size])
		if x < (w0+w1)/2 {
			return idx
		}
		idx += size
	}
	return idx
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

// textFieldTextX returns the x position where the text field's text starts, adjusting
// the field's horizontal scroll offset so that the caret stays visible.
func (c *Context) textFieldTextX(f *textFieldState, bounds image.Rectangle, opt option) int {
	padding := c.style().padding
	innerWidth := bounds.Dx() - 2*padding
	w := textWidth(f.textForRendering())
	if w <= innerWidth {
		f.scrollX = 0
		switch {
		case opt&optionAlignCenter != 0:
			return bounds.Min.X + (bounds.Dx()-w)/2
		case opt&optionAlignRight != 0:
			return bounds.Min.X + bounds.Dx() - w - padding
		default:
			return bounds.Min.X + padding
		}
	}
	// Keep the caret within the visible range. One pixel is reserved for the caret itself.
	caretX := textWidth(f.textForRendering()[:f.caretRenderIndex()])
	f.scrollX = min(f.scrollX, caretX)
	f.scrollX = max(f.scrollX, caretX-innerWidth+1)
	f.scrollX = min(f.scrollX, w-innerWidth+1)
	f.scrollX = max(f.scrollX, 0)
	return bounds.Min.X + padding - f.scrollX
}

func (c *Context) textFieldRaw(buf *string, id widgetID, opt option) (EventHandler, error) {
	return c.widget(id, opt|optionHoldFocus, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler

		f := c.currentContainer().textInputTextField(id, true)
		if c.focus == id {
			// While focused, f's committed text is the source of truth.
			textx := c.textFieldTextX(f, bounds, opt)

			// Handle pointing input before running the IME, so that a new IME session
			// starts with the up-to-date selection.
			pt := c.pointingPosition()
			if c.pointing.justPressed() && c.pointingOver(bounds) {
				// End the session first; this commits any in-progress composition, and the
				// caret position is then resolved against the resulting committed text.
				f.composer.Finish()
				idx := textIndexFromX(f.text(), pt.X-textx)
				f.handleClick(idx, ebiten.IsKeyPressed(ebiten.KeyShift), ebiten.Tick(), int64(ebiten.TPS())/2)
			} else if f.dragging {
				if c.pointing.pressed() {
					f.moveCaretTo(textIndexFromX(f.text(), pt.X-textx), true)
				} else {
					f.dragging = false
				}
			}

			texth := lineHeight()
			texty := bounds.Min.Y + (bounds.Dy()-texth)/2
			caretX := textx + textWidth(f.textForRendering()[:f.caretRenderIndex()])
			// The IME takes the caret bounds in the screen coordinate space.
			scale := c.Scale()
			handled, err := f.update(image.Rect(caretX*scale, texty*scale, (caretX+1)*scale, (texty+texth)*scale))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}

			if !handled {
				shift := ebiten.IsKeyPressed(ebiten.KeyShift)
				apple := isApplePlatform()
				// cmd is the primary shortcut modifier (e.g. select all): Command on Apple
				// platforms, Control elsewhere. On Apple platforms Command+Left/Right also
				// jumps to the line ends; elsewhere the line ends are reached with Home/End.
				// word is the word-wise movement modifier: Option (Alt) on Apple platforms,
				// Control elsewhere.
				var cmd, word bool
				if apple {
					cmd = ebiten.IsKeyPressed(ebiten.KeyMeta)
					word = ebiten.IsKeyPressed(ebiten.KeyAlt)
				} else {
					cmd = ebiten.IsKeyPressed(ebiten.KeyControl)
					word = ebiten.IsKeyPressed(ebiten.KeyControl)
				}
				switch {
				case keyRepeated(ebiten.KeyLeft):
					switch {
					case apple && cmd:
						f.moveCaretTo(0, shift)
					case word:
						f.moveCaretWordLeft(shift)
					default:
						f.moveCaretLeft(shift)
					}
				case keyRepeated(ebiten.KeyRight):
					switch {
					case apple && cmd:
						f.moveCaretTo(len(f.text()), shift)
					case word:
						f.moveCaretWordRight(shift)
					default:
						f.moveCaretRight(shift)
					}
				case keyRepeated(ebiten.KeyHome):
					f.moveCaretTo(0, shift)
				case keyRepeated(ebiten.KeyEnd):
					f.moveCaretTo(len(f.text()), shift)
				case inpututil.IsKeyJustPressed(ebiten.KeyA) && cmd:
					f.selectAll()
				case keyRepeated(ebiten.KeyBackspace):
					f.deleteBackward()
				case keyRepeated(ebiten.KeyDelete):
					f.deleteForward()
				case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
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
			texth := lineHeight()
			texty := bounds.Min.Y + (bounds.Dy()-texth)/2
			textx := c.textFieldTextX(f, bounds, opt)
			str := f.textForRendering()
			c.pushClipRect(bounds)
			if f.composition != "" {
				// The composition visually replaces the selection. Highlight the IME-side
				// selection within the composition, and underline the whole composition.
				compositionX := textx + textWidth(f.committedText[:f.selectionStart()])
				if f.compositionSelStart != f.compositionSelEnd {
					x0 := compositionX + textWidth(f.composition[:f.compositionSelStart])
					x1 := compositionX + textWidth(f.composition[:f.compositionSelEnd])
					c.drawRect(image.Rect(x0, texty, x1, texty+texth), c.style().colors[colorSelection])
				}
				c.drawRect(image.Rect(compositionX, texty+texth-1, compositionX+textWidth(f.composition), texty+texth), color)
			} else if f.selectionStart() != f.selectionEnd() {
				x0 := textx + textWidth(str[:f.selectionStart()])
				x1 := textx + textWidth(str[:f.selectionEnd()])
				c.drawRect(image.Rect(x0, texty, x1, texty+texth), c.style().colors[colorSelection])
			}
			c.drawText(str, image.Pt(textx, texty), color)
			caretX := textx + textWidth(str[:f.caretRenderIndex()])
			c.drawRect(image.Rect(caretX, texty, caretX+1, texty+texth), color)
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

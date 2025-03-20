// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const idSeparator = "\x00"

const (
	realFmt   = "%.3g"
	sliderFmt = "%.2f"
)

type option int

const (
	optionAlignCenter option = (1 << iota)
	optionAlignRight
	optionNoInteract
	optionNoFrame
	optionNoResize
	optionNoScroll
	optionNoClose
	optionNoTitle
	optionHoldFocus
	optionAutoSize
	optionPopup
	optionClosed
	optionExpanded
)

func (c *Context) inHoverRoot() bool {
	for i := len(c.containerStack) - 1; i >= 0; i-- {
		if c.containerStack[i] == c.hoverRoot {
			return true
		}
		// only root containers have their `head` field set; stop searching if we've
		// reached the current root container
		if c.containerStack[i].headIdx >= 0 {
			break
		}
	}
	return false
}

func (c *Context) mouseOver(bounds image.Rectangle) bool {
	p := image.Pt(ebiten.CursorPosition())
	return p.In(bounds) && p.In(c.clipRect()) && c.inHoverRoot()
}

func (c *Context) updateControl(id controlID, bounds image.Rectangle, opt option) (wasFocused bool) {
	if id == 0 {
		return false
	}

	mouseover := c.mouseOver(bounds)

	if c.focus == id {
		c.keepFocus = true
	}
	if (opt & optionNoInteract) != 0 {
		return false
	}
	if mouseover && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		c.hover = id
	}

	if c.focus == id {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !mouseover {
			c.setFocus(0)
			wasFocused = true
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && (^opt&optionHoldFocus) != 0 {
			c.setFocus(0)
			wasFocused = true
		}
	}

	if c.hover == id {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			c.setFocus(id)
		} else if !mouseover {
			c.hover = 0
		}
	}

	return
}

func (c *Context) Control(idStr string, f func(bounds image.Rectangle) bool) bool {
	id := c.idFromString(idStr)
	return c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) bool {
		return f(bounds)
	})
}

func (c *Context) control(id controlID, opt option, f func(bounds image.Rectangle, wasFocused bool) bool) bool {
	r := c.layoutNext()
	wasFocused := c.updateControl(id, r, opt)
	return f(r, wasFocused)
}

func (c *Context) Text(text string) {
	color := c.style.colors[ColorText]
	c.GridCell(func() {
		var endIdx, p int
		c.SetGridLayout([]int{-1}, []int{lineHeight()})
		for endIdx < len(text) {
			c.control(0, 0, func(bounds image.Rectangle, wasFocused bool) bool {
				w := 0
				endIdx = p
				startIdx := endIdx
				for endIdx < len(text) && text[endIdx] != '\n' {
					word := p
					for p < len(text) && text[p] != ' ' && text[p] != '\n' {
						p++
					}
					w += textWidth(text[word:p])
					if w > bounds.Dx() && endIdx != startIdx {
						break
					}
					if p < len(text) {
						w += textWidth(string(text[p]))
					}
					endIdx = p
					p++
				}
				c.drawText(text[startIdx:endIdx], bounds.Min, color)
				p = endIdx + 1
				return false
			})
		}
	})
}

func (c *Context) Label(text string) {
	c.control(0, 0, func(bounds image.Rectangle, wasFocused bool) bool {
		c.drawControlText(text, bounds, ColorText, 0)
		return false
	})
}

func (c *Context) button(label string, opt option) (controlID, bool) {
	label, idStr, _ := strings.Cut(label, idSeparator)
	id := c.idFromString(idStr)
	return id, c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) bool {
		var res bool
		// handle click
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
			res = true
		}
		// draw
		c.drawControlFrame(id, bounds, ColorButton, opt)
		if len(label) > 0 {
			c.drawControlText(label, bounds, ColorText, opt)
		}
		return res
	})
}

func (c *Context) Checkbox(label string, state *bool) bool {
	id := c.idFromGlobalUniqueString(fmt.Sprintf("%p", state))

	return c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) bool {
		var res bool
		box := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+bounds.Dy(), bounds.Max.Y)
		c.updateControl(id, bounds, 0)
		// handle click
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
			res = true
			*state = !*state
		}
		// draw
		c.drawControlFrame(id, box, ColorBase, 0)
		if *state {
			c.drawIcon(iconCheck, box, c.style.colors[ColorText])
		}
		bounds = image.Rect(bounds.Min.X+box.Dx(), bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
		c.drawControlText(label, bounds, ColorText, 0)
		return res
	})
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

		if c.focus == id {
			// handle text input
			f := c.textInputTextField(id)
			f.Focus()
			x := bounds.Min.X + c.style.padding + textWidth(*buf)
			y := bounds.Min.Y + lineHeight()
			handled, err := f.HandleInput(x, y)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return false
			}
			if *buf != f.TextForRendering() {
				*buf = f.TextForRendering()
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
			f := c.textInputTextField(id)
			if *buf != f.TextForRendering() {
				f.SetTextAndSelection(*buf, len(*buf), len(*buf))
			}
			if wasFocused {
				res = true
			}
		}

		// draw
		c.drawControlFrame(id, bounds, ColorBase, opt)
		if c.focus == id {
			color := c.style.colors[ColorText]
			textw := textWidth(*buf)
			texth := lineHeight()
			ofx := bounds.Dx() - c.style.padding - textw - 1
			textx := bounds.Min.X + min(ofx, c.style.padding)
			texty := bounds.Min.Y + (bounds.Dy()-texth)/2
			c.pushClipRect(bounds)
			c.drawText(*buf, image.Pt(textx, texty), color)
			c.drawRect(image.Rect(textx+textw, texty, textx+textw+1, texty+texth), color)
			c.popClipRect()
		} else {
			c.drawControlText(*buf, bounds, ColorText, opt)
		}
		return res
	})
}

func (c *Context) SetTextFieldValue(value *string) {
	id := c.idFromGlobalUniqueString(fmt.Sprintf("%p", value))
	f := c.textInputTextField(id)
	f.SetTextAndSelection(*value, 0, 0)
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

func (c *Context) textField(buf *string, opt option) bool {
	id := c.idFromGlobalUniqueString(fmt.Sprintf("%p", buf))
	return c.textFieldRaw(buf, id, opt)
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

func (c *Context) slider(value *float64, low, high, step float64, digits int, opt option) bool {
	last := *value
	v := last
	id := c.idFromGlobalUniqueString(fmt.Sprintf("%p", value))

	// handle text input mode
	if c.numberTextField(&v, id) {
		*value = v
		return false
	}

	// handle normal mode
	return c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) bool {
		var res bool
		// handle input
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, _ := ebiten.CursorPosition()
			v = low + float64(x-bounds.Min.X)*(high-low)/float64(bounds.Dx())
			if step != 0 {
				v = math.Round(v/step) * step
			}
		}
		// clamp and store value, update res
		*value = clamp(v, low, high)
		v = *value
		if last != v {
			res = true
		}

		// draw base
		c.drawControlFrame(id, bounds, ColorBase, opt)
		// draw thumb
		w := c.style.thumbSize
		x := int((v - low) * float64(bounds.Dx()-w) / (high - low))
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawControlFrame(id, thumb, ColorButton, opt)
		// draw text
		text := formatNumber(v, digits)
		c.drawControlText(text, bounds, ColorText, opt)

		return res
	})
}

func (c *Context) number(value *float64, step float64, digits int, opt option) bool {
	id := c.idFromGlobalUniqueString(fmt.Sprintf("%p", value))
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

func (c *Context) header(label string, istreenode bool, opt option, f func()) {
	label, idStr, _ := strings.Cut(label, idSeparator)
	id := c.idFromString(idStr)
	_, toggled := c.toggledIDs[id]
	c.SetGridLayout([]int{-1}, nil)

	var expanded bool
	if (opt & optionExpanded) != 0 {
		expanded = !toggled
	} else {
		expanded = toggled
	}

	if c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) bool {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
			if toggled {
				delete(c.toggledIDs, id)
			} else {
				if c.toggledIDs == nil {
					c.toggledIDs = map[controlID]struct{}{}
				}
				c.toggledIDs[id] = struct{}{}
			}
		}

		// draw
		if istreenode {
			if c.hover == id {
				c.drawFrame(bounds, ColorButtonHover)
			}
		} else {
			c.drawControlFrame(id, bounds, ColorButton, 0)
		}
		var icon icon
		if expanded {
			icon = iconExpanded
		} else {
			icon = iconCollapsed
		}
		c.drawIcon(
			icon,
			image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+bounds.Dy(), bounds.Max.Y),
			c.style.colors[ColorText],
		)
		bounds.Min.X += bounds.Dy() - c.style.padding
		c.drawControlText(label, bounds, ColorText, 0)

		return expanded
	}) {
		f()
	}
}

func (c *Context) treeNode(label string, opt option, f func()) {
	c.header(label, true, opt, func() {
		c.layout().indent += c.style.indent
		defer func() {
			c.layout().indent -= c.style.indent
		}()
		f()
	})
}

// x = x, y = y, w = w, h = h
func (c *Context) scrollbarVertical(cnt *container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.Y - b.Dy()
	if maxscroll > 0 && b.Dy() > 0 {
		// get sizing / positioning
		base := b
		base.Min.X = b.Max.X
		base.Max.X = base.Min.X + c.style.scrollbarSize

		// handle input
		id := c.idFromString("!scrollbar" + "y")
		c.updateControl(id, base, 0)
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.ScrollOffset.Y += c.mouseDelta().Y * cs.Y / base.Dy()
		}
		// clamp scroll to limits
		cnt.layout.ScrollOffset.Y = clamp(cnt.layout.ScrollOffset.Y, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.Y = thumb.Min.Y + max(c.style.thumbSize, base.Dy()*b.Dy()/cs.Y)
		thumb = thumb.Add(image.Pt(0, cnt.layout.ScrollOffset.Y*(base.Dy()-thumb.Dy())/maxscroll))
		c.drawFrame(thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if c.mouseOver(b) {
			c.scrollTarget = cnt
		}
	} else {
		cnt.layout.ScrollOffset.Y = 0
	}
}

// x = y, y = x, w = h, h = w
func (c *Context) scrollbarHorizontal(cnt *container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.X - b.Dx()
	if maxscroll > 0 && b.Dx() > 0 {
		// get sizing / positioning
		base := b
		base.Min.Y = b.Max.Y
		base.Max.Y = base.Min.Y + c.style.scrollbarSize

		// handle input
		id := c.idFromString("!scrollbar" + "x")
		c.updateControl(id, base, 0)
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.ScrollOffset.X += c.mouseDelta().X * cs.X / base.Dx()
		}
		// clamp scroll to limits
		cnt.layout.ScrollOffset.X = clamp(cnt.layout.ScrollOffset.X, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.X = thumb.Min.X + max(c.style.thumbSize, base.Dx()*b.Dx()/cs.X)
		thumb = thumb.Add(image.Pt(cnt.layout.ScrollOffset.X*(base.Dx()-thumb.Dx())/maxscroll, 0))
		c.drawFrame(thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if c.mouseOver(b) {
			c.scrollTarget = cnt
		}
	} else {
		cnt.layout.ScrollOffset.X = 0
	}
}

// if `swap` is true, X = Y, Y = X, W = H, H = W
func (c *Context) scrollbar(cnt *container, b image.Rectangle, cs image.Point, swap bool) {
	if swap {
		c.scrollbarHorizontal(cnt, b, cs)
	} else {
		c.scrollbarVertical(cnt, b, cs)
	}
}

func (c *Context) scrollbars(cnt *container, body image.Rectangle) image.Rectangle {
	sz := c.style.scrollbarSize
	cs := cnt.layout.ContentSize
	cs.X += c.style.padding * 2
	cs.Y += c.style.padding * 2
	c.pushClipRect(body)
	// resize body to make room for scrollbars
	if cs.Y > cnt.layout.BodyBounds.Dy() {
		body.Max.X -= sz
	}
	if cs.X > cnt.layout.BodyBounds.Dx() {
		body.Max.Y -= sz
	}
	// to create a horizontal or vertical scrollbar almost-identical code is
	// used; only the references to `x|y` `w|h` need to be switched
	c.scrollbar(cnt, body, cs, false)
	c.scrollbar(cnt, body, cs, true)
	c.popClipRect()
	return body
}

func (c *Context) pushContainerBodyLayout(cnt *container, body image.Rectangle, opt option) {
	if (^opt & optionNoScroll) != 0 {
		body = c.scrollbars(cnt, body)
	}
	c.pushLayout(body.Inset(c.style.padding), cnt.layout.ScrollOffset)
	cnt.layout.BodyBounds = body
}

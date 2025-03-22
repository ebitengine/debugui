// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
	"math"
	"strings"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type controlID string

const emptyControlID controlID = ""

const idSeparator = "\x00"

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
	p := c.cursorPosition()
	return p.In(bounds) && p.In(c.clipRect()) && c.inHoverRoot()
}

func (c *Context) updateControl(id controlID, bounds image.Rectangle, opt option) (wasFocused bool) {
	if id == emptyControlID {
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
			c.setFocus(emptyControlID)
			wasFocused = true
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && (^opt&optionHoldFocus) != 0 {
			c.setFocus(emptyControlID)
			wasFocused = true
		}
	}

	if c.hover == id {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			c.setFocus(id)
		} else if !mouseover {
			c.hover = emptyControlID
		}
	}

	return
}

func (c *Context) Control(idStr string, f func(bounds image.Rectangle) bool) bool {
	pc := caller()
	var res bool
	c.wrapError(func() error {
		id := c.idFromCaller(pc, idStr)
		var err error
		res, err = c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
			return f(bounds), nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) control(id controlID, opt option, f func(bounds image.Rectangle, wasFocused bool) (bool, error)) (bool, error) {
	r, err := c.layoutNext()
	if err != nil {
		return false, err
	}
	wasFocused := c.updateControl(id, r, opt)
	res, err := f(r, wasFocused)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) Text(text string) {
	c.wrapError(func() error {
		if err := c.gridCell(func() error {
			var endIdx, p int
			c.SetGridLayout([]int{-1}, []int{lineHeight()})
			for endIdx < len(text) {
				if _, err := c.control(emptyControlID, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
					w := 0
					endIdx = p
					startIdx := endIdx
					for endIdx < len(text) && text[endIdx] != '\n' {
						word := p
						for p < len(text) && text[p] != ' ' && text[p] != '\n' {
							p++
						}
						w += textWidth(text[word:p])
						if w > bounds.Dx()-c.style().padding && endIdx != startIdx {
							break
						}
						if p < len(text) {
							w += textWidth(string(text[p]))
						}
						endIdx = p
						p++
					}
					c.drawControlText(text[startIdx:endIdx], bounds, ColorText, 0)
					p = endIdx + 1
					return false, nil
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) button(label string, opt option, callerPC uintptr) (controlID, bool, error) {
	label, idStr, _ := strings.Cut(label, idSeparator)
	id := c.idFromCaller(callerPC, idStr)
	res, err := c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
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
		return res, nil
	})
	if err != nil {
		return emptyControlID, false, err
	}
	return id, res, nil
}

func (c *Context) Checkbox(state *bool, label string) bool {
	var res bool
	c.wrapError(func() error {
		id := c.idFromPointer(unsafe.Pointer(state))
		var err error
		res, err = c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
			var res bool
			box := image.Rect(bounds.Min.X, bounds.Min.Y+(bounds.Dy()-lineHeight())/2, bounds.Min.X+lineHeight(), bounds.Max.Y-(bounds.Dy()-lineHeight())/2)
			c.updateControl(id, bounds, 0)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
				res = true
				*state = !*state
			}
			c.drawControlFrame(id, box, ColorBase, 0)
			if *state {
				c.drawIcon(iconCheck, box, c.style().colors[ColorText])
			}
			if label != "" {
				bounds = image.Rect(bounds.Min.X+lineHeight(), bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
				c.drawControlText(label, bounds, ColorText, 0)
			}
			return res, nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func (c *Context) slider(value *float64, low, high, step float64, digits int, opt option) (bool, error) {
	last := *value
	v := last
	id := c.idFromPointer(unsafe.Pointer(value))

	// handle text input mode
	res, err := c.numberTextField(&v, id)
	if err != nil {
		return false, err
	}
	if res {
		*value = v
		return false, nil
	}

	// handle normal mode
	res, err = c.control(id, opt, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		var res bool
		// handle input
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			v = low + float64(c.cursorPosition().X-bounds.Min.X)*(high-low)/float64(bounds.Dx())
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
		w := c.style().thumbSize
		x := int((v - low) * float64(bounds.Dx()-w) / (high - low))
		thumb := image.Rect(bounds.Min.X+x, bounds.Min.Y, bounds.Min.X+x+w, bounds.Max.Y)
		c.drawControlFrame(id, thumb, ColorButton, opt)
		// draw text
		text := formatNumber(v, digits)
		c.drawControlText(text, bounds, ColorText, opt)

		return res, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Context) header(label string, isTreeNode bool, opt option, callerPC uintptr, f func() error) error {
	label, idStr, _ := strings.Cut(label, idSeparator)
	id := c.idFromCaller(callerPC, idStr)
	c.SetGridLayout(nil, nil)

	var expanded bool
	toggled := c.currentContainer().toggled(id)
	if (opt & optionExpanded) != 0 {
		expanded = !toggled
	} else {
		expanded = toggled
	}

	res, err := c.control(id, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
			c.currentContainer().toggle(id)
		}
		if isTreeNode {
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
			c.style().colors[ColorText],
		)
		bounds.Min.X += bounds.Dy() - c.style().padding
		c.drawControlText(label, bounds, ColorText, 0)

		return expanded, nil
	})
	if err != nil {
		return err
	}
	if res {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) treeNode(label string, opt option, callerPC uintptr, f func()) error {
	if err := c.header(label, true, opt, callerPC, func() (err error) {
		l, err := c.layout()
		if err != nil {
			return err
		}
		l.indent += c.style().indent
		defer func() {
			l, err2 := c.layout()
			if err2 != nil && err == nil {
				err = err2
				return
			}
			l.indent -= c.style().indent
		}()
		f()
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// x = x, y = y, w = w, h = h
func (c *Context) scrollbarVertical(cnt *container, b image.Rectangle, cs image.Point, callerPC uintptr) {
	maxscroll := cs.Y - b.Dy()
	if maxscroll > 0 && b.Dy() > 0 {
		// get sizing / positioning
		base := b
		base.Min.X = b.Max.X
		base.Max.X = base.Min.X + c.style().scrollbarSize

		// handle input
		id := c.idFromCaller(callerPC, "scrollbar-y")
		c.updateControl(id, base, 0)
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.ScrollOffset.Y += c.mouseDelta().Y * cs.Y / base.Dy()
		}
		// clamp scroll to limits
		cnt.layout.ScrollOffset.Y = clamp(cnt.layout.ScrollOffset.Y, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.Y = thumb.Min.Y + max(c.style().thumbSize, base.Dy()*b.Dy()/cs.Y)
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
func (c *Context) scrollbarHorizontal(cnt *container, b image.Rectangle, cs image.Point, callerPC uintptr) {
	maxscroll := cs.X - b.Dx()
	if maxscroll > 0 && b.Dx() > 0 {
		// get sizing / positioning
		base := b
		base.Min.Y = b.Max.Y
		base.Max.Y = base.Min.Y + c.style().scrollbarSize

		// handle input
		id := c.idFromCaller(callerPC, "scrollbar-x")
		c.updateControl(id, base, 0)
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.ScrollOffset.X += c.mouseDelta().X * cs.X / base.Dx()
		}
		// clamp scroll to limits
		cnt.layout.ScrollOffset.X = clamp(cnt.layout.ScrollOffset.X, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.X = thumb.Min.X + max(c.style().thumbSize, base.Dx()*b.Dx()/cs.X)
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
func (c *Context) scrollbar(cnt *container, b image.Rectangle, cs image.Point, swap bool, callerPC uintptr) {
	if swap {
		c.scrollbarHorizontal(cnt, b, cs, callerPC)
	} else {
		c.scrollbarVertical(cnt, b, cs, callerPC)
	}
}

func (c *Context) scrollbars(cnt *container, body image.Rectangle, callerPC uintptr) image.Rectangle {
	sz := c.style().scrollbarSize
	cs := cnt.layout.ContentSize
	cs.X += c.style().padding * 2
	cs.Y += c.style().padding * 2
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
	c.scrollbar(cnt, body, cs, false, callerPC)
	c.scrollbar(cnt, body, cs, true, callerPC)
	c.popClipRect()
	return body
}

func (c *Context) pushContainerBodyLayout(cnt *container, body image.Rectangle, opt option, callerPC uintptr) error {
	if (^opt & optionNoScroll) != 0 {
		body = c.scrollbars(cnt, body, callerPC)
	}
	if err := c.pushLayout(body.Inset(c.style().padding), cnt.layout.ScrollOffset, opt&optionAutoSize != 0); err != nil {
		return err
	}
	cnt.layout.BodyBounds = body
	return nil
}

// SetScale sets the scale of the UI.
//
// The scale affects the rendering result of the UI.
//
// The default scale is 1.
func (c *Context) SetScale(scale int) {
	if scale < 1 {
		panic("debugui: scale must be >= 1")
	}
	c.scaleMinus1 = scale - 1
}

// Scale returns the scale of the UI.
func (c *Context) Scale() int {
	return c.scaleMinus1 + 1
}

func (c *Context) style() *style {
	return &defaultStyle
}

func (c *Context) isCapturingInput() bool {
	if c.err != nil {
		return false
	}

	return c.hoverRoot != nil || c.focus != emptyControlID
}

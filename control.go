// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (c *Context) drawFrame(rect image.Rectangle, colorid int) {
	c.drawRect(rect, c.style.colors[colorid])
	if colorid == ColorScrollBase || colorid == ColorScrollThumb || colorid == ColorTitleBG {
		return
	}
	// draw border
	if c.style.colors[ColorBorder].A != 0 {
		c.drawBox(rect.Inset(-1), c.style.colors[ColorBorder])
	}
}

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

func (c *Context) drawControlFrame(id controlID, rect image.Rectangle, colorid int, opt option) {
	if (opt & optionNoFrame) != 0 {
		return
	}
	if c.focus == id {
		colorid += 2
	} else if c.hover == id {
		colorid++
	}
	c.drawFrame(rect, colorid)
}

func (c *Context) drawControlText(str string, rect image.Rectangle, colorid int, opt option) {
	var pos image.Point
	tw := textWidth(str)
	c.pushClipRect(rect)
	pos.Y = rect.Min.Y + (rect.Dy()-lineHeight())/2
	if (opt & optionAlignCenter) != 0 {
		pos.X = rect.Min.X + (rect.Dx()-tw)/2
	} else if (opt & optionAlignRight) != 0 {
		pos.X = rect.Min.X + rect.Dx() - tw - c.style.padding
	} else {
		pos.X = rect.Min.X + c.style.padding
	}
	c.drawText(str, pos, c.style.colors[colorid])
	c.popClipRect()
}

func (c *Context) mouseOver(rect image.Rectangle) bool {
	p := image.Pt(ebiten.CursorPosition())
	return p.In(rect) && p.In(c.clipRect()) && c.inHoverRoot()
}

func (c *Context) updateControl(id controlID, rect image.Rectangle, opt option) {
	if id == 0 {
		return
	}

	mouseover := c.mouseOver(rect)

	if c.focus == id {
		c.keepFocus = true
	}
	if (opt & optionNoInteract) != 0 {
		return
	}
	if mouseover && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		c.hover = id
	}

	if c.focus == id {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !mouseover {
			c.setFocus(0)
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && (^opt&optionHoldFocus) != 0 {
			c.setFocus(0)
		}
	}

	if c.hover == id {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			c.setFocus(id)
		} else if !mouseover {
			c.hover = 0
		}
	}
}

func (c *Context) Control(idStr string, f func(bounds image.Rectangle) Response) Response {
	id := c.pushID([]byte(idStr))
	defer c.popID()
	return c.control(id, 0, f)
}

func (c *Context) control(id controlID, opt option, f func(bounds image.Rectangle) Response) Response {
	r := c.layoutNext()
	c.updateControl(id, r, opt)
	return f(r)
}

func (c *Context) Text(text string) {
	color := c.style.colors[ColorText]
	c.LayoutColumn(func() {
		var endIdx, p int
		c.SetLayoutRow([]int{-1}, lineHeight())
		for endIdx < len(text) {
			c.control(0, 0, func(bounds image.Rectangle) Response {
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
				return 0
			})
		}
	})
}

func (c *Context) Label(text string) {
	c.control(0, 0, func(bounds image.Rectangle) Response {
		c.drawControlText(text, bounds, ColorText, 0)
		return 0
	})
}

func (c *Context) button(label string, idStr string, opt option) Response {
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(label) > 0 {
		id = c.pushID([]byte(label))
		defer c.popID()
	}
	return c.control(id, opt, func(bounds image.Rectangle) Response {
		var res Response
		// handle click
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
			res |= ResponseSubmit
		}
		// draw
		c.drawControlFrame(id, bounds, ColorButton, opt)
		if len(label) > 0 {
			c.drawControlText(label, bounds, ColorText, opt)
		}
		return res
	})
}

func (c *Context) Checkbox(label string, state *bool) Response {
	id := c.pushID(ptrToBytes(unsafe.Pointer(state)))
	defer c.popID()

	return c.control(id, 0, func(bounds image.Rectangle) Response {
		var res Response
		box := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+bounds.Dy(), bounds.Max.Y)
		c.updateControl(id, bounds, 0)
		// handle click
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.focus == id {
			res |= ResponseChange
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

func (c *Context) textField(id controlID) *textinput.Field {
	if id == 0 {
		return nil
	}
	if _, ok := c.textFields[id]; !ok {
		if c.textFields == nil {
			c.textFields = make(map[controlID]*textinput.Field)
		}
		c.textFields[id] = &textinput.Field{}
	}
	return c.textFields[id]
}

func (c *Context) textBoxRaw(buf *string, id controlID, opt option) Response {
	return c.control(id, opt|optionHoldFocus, func(bounds image.Rectangle) Response {
		var res Response

		if c.focus == id {
			// handle text input
			f := c.textField(id)
			f.Focus()
			x := bounds.Min.X + c.style.padding + textWidth(*buf)
			y := bounds.Min.Y + lineHeight()
			handled, err := f.HandleInput(x, y)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 0
			}
			if *buf != f.TextForRendering() {
				*buf = f.TextForRendering()
				res |= ResponseChange
			}

			if !handled {
				// handle backspace
				if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(*buf) > 0 {
					_, size := utf8.DecodeLastRuneInString(*buf)
					*buf = (*buf)[:len(*buf)-size]
					f.SetTextAndSelection(*buf, len(*buf), len(*buf))
					res |= ResponseChange
				}

				// handle return
				if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
					c.setFocus(0)
					res |= ResponseSubmit
					f.SetTextAndSelection("", 0, 0)
				}
			}
		} else {
			f := c.textField(id)
			if *buf != f.TextForRendering() {
				f.SetTextAndSelection(*buf, len(*buf), len(*buf))
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

func (c *Context) numberTextBox(value *float64, id controlID) bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && ebiten.IsKeyPressed(ebiten.KeyShift) &&
		c.hover == id {
		c.numberEdit = id
		c.numberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.numberEdit == id {
		res := c.textBoxRaw(&c.numberEditBuf, id, 0)
		if (res&ResponseSubmit) != 0 || c.focus != id {
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

func (c *Context) textBox(buf *string, opt option) Response {
	id := c.pushID(ptrToBytes(unsafe.Pointer(buf)))
	defer c.popID()

	return c.textBoxRaw(buf, id, opt)
}

func formatNumber(v float64, digits int) string {
	return fmt.Sprintf("%."+strconv.Itoa(digits)+"f", v)
}

func (c *Context) slider(value *float64, low, high, step float64, digits int, opt option) Response {
	last := *value
	v := last
	id := c.pushID(ptrToBytes(unsafe.Pointer(value)))
	defer c.popID()

	// handle text input mode
	if c.numberTextBox(&v, id) {
		return 0
	}

	// handle normal mode
	return c.control(id, opt, func(bounds image.Rectangle) Response {
		var res Response
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
			res |= ResponseChange
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

func (c *Context) number(value *float64, step float64, digits int, opt option) Response {
	id := c.pushID(ptrToBytes(unsafe.Pointer(value)))
	defer c.popID()
	last := *value

	// handle text input mode
	if c.numberTextBox(value, id) {
		return 0
	}

	// handle normal mode
	return c.control(id, opt, func(bounds image.Rectangle) Response {
		var res Response
		// handle input
		if c.focus == id && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			*value += float64(c.mouseDelta().X) * step
		}
		// set flag if value changed
		if *value != last {
			res |= ResponseChange
		}

		// draw base
		c.drawControlFrame(id, bounds, ColorBase, opt)
		// draw text
		text := formatNumber(*value, digits)
		c.drawControlText(text, bounds, ColorText, opt)

		return res
	})
}

func (c *Context) header(label string, idStr string, istreenode bool, opt option) Response {
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(label) > 0 {
		id = c.pushID([]byte(label))
		defer c.popID()
	}

	_, toggled := c.toggledIDs[id]
	c.SetLayoutRow([]int{-1}, 0)

	var expanded bool
	if (opt & optionExpanded) != 0 {
		expanded = !toggled
	} else {
		expanded = toggled
	}

	return c.control(id, 0, func(bounds image.Rectangle) Response {
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

		if expanded {
			return ResponseActive
		}
		return 0
	})
}

func (c *Context) treeNode(label string, idStr string, opt option, f func(res Response)) {
	res := c.header(label, idStr, true, opt)
	if res&ResponseActive == 0 {
		return
	}
	c.layout().indent += c.style.indent
	defer func() {
		c.layout().indent -= c.style.indent
	}()
	f(res)
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
		id := c.idFromBytes([]byte("!scrollbar" + "y"))
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
		id := c.idFromBytes([]byte("!scrollbar" + "x"))
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

func (c *Context) window(title string, idStr string, rect image.Rectangle, opt option, f func(res Response, layout ContainerLayout)) {
	var id controlID
	if len(idStr) > 0 {
		id = c.pushID([]byte(idStr))
		defer c.popID()
	} else if len(title) > 0 {
		id = c.pushID([]byte(title))
		defer c.popID()
	}

	cnt := c.container(id, opt)
	if cnt == nil || !cnt.open {
		return
	}
	// This is popped at endRootContainer.
	// TODO: This is tricky. Refactor this.

	if cnt.layout.Bounds.Dx() == 0 {
		cnt.layout.Bounds = rect
	}

	c.containerStack = append(c.containerStack, cnt)
	defer c.popContainer()

	// push container to roots list and push head command
	c.rootList = append(c.rootList, cnt)
	cnt.headIdx = c.pushJump(-1)
	defer func() {
		// push tail 'goto' jump command and set head 'skip' command. the final steps
		// on initing these are done in End
		cnt := c.currentContainer()
		cnt.tailIdx = c.pushJump(-1)
		c.commandList[cnt.headIdx].jump.dstIdx = len(c.commandList) //- 1
	}()

	// set as hover root if the mouse is overlapping this container and it has a
	// higher zindex than the current hover root
	if image.Pt(ebiten.CursorPosition()).In(cnt.layout.Bounds) && (c.nextHoverRoot == nil || cnt.zIndex > c.nextHoverRoot.zIndex) {
		c.nextHoverRoot = cnt
	}

	// clipping is reset here in case a root-container is made within
	// another root-containers's begin/end block; this prevents the inner
	// root-container being clipped to the outer
	c.clipStack = append(c.clipStack, unclippedRect)
	defer c.popClipRect()

	body := cnt.layout.Bounds
	rect = body

	// draw frame
	collapsed := cnt.collapsed
	if (^opt&optionNoFrame) != 0 && !collapsed {
		c.drawFrame(rect, ColorWindowBG)
	}

	// do title bar
	if (^opt & optionNoTitle) != 0 {
		tr := rect
		tr.Max.Y = tr.Min.Y + c.style.titleHeight
		c.drawFrame(tr, ColorTitleBG)

		// do title text
		if (^opt & optionNoTitle) != 0 {
			id := c.idFromBytes([]byte("!title"))
			r := image.Rect(tr.Min.X+tr.Dy(), tr.Min.Y, tr.Max.X, tr.Max.Y)
			c.updateControl(id, r, opt)
			c.drawControlText(title, r, ColorTitleText, opt)
			if id == c.focus && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				cnt.layout.Bounds = cnt.layout.Bounds.Add(c.mouseDelta())
			}
			body.Min.Y += tr.Dy()
		}

		// do `close` button
		if (^opt & optionNoClose) != 0 {
			id := c.idFromBytes([]byte("!close"))
			r := image.Rect(tr.Min.X, tr.Min.Y, tr.Min.X+tr.Dy(), tr.Max.Y)
			icon := iconExpanded
			if collapsed {
				icon = iconCollapsed
			}
			c.drawIcon(icon, r, c.style.colors[ColorTitleText])
			c.updateControl(id, r, opt)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && id == c.focus {
				cnt.collapsed = !cnt.collapsed
			}
		}
	}

	if collapsed {
		return
	}

	c.pushContainerBodyLayout(cnt, body, opt)
	defer c.popLayout()

	// do `resize` handle
	if (^opt & optionNoResize) != 0 {
		sz := c.style.titleHeight
		id := c.idFromBytes([]byte("!resize"))
		r := image.Rect(rect.Max.X-sz, rect.Max.Y-sz, rect.Max.X, rect.Max.Y)
		c.updateControl(id, r, opt)
		if id == c.focus && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cnt.layout.Bounds.Max.X = cnt.layout.Bounds.Min.X + max(96, cnt.layout.Bounds.Dx()+c.mouseDelta().X)
			cnt.layout.Bounds.Max.Y = cnt.layout.Bounds.Min.Y + max(64, cnt.layout.Bounds.Dy()+c.mouseDelta().Y)
		}
	}

	// resize to content size
	if (opt & optionAutoSize) != 0 {
		r := c.layout().body
		cnt.layout.Bounds.Max.X = cnt.layout.Bounds.Min.X + cnt.layout.ContentSize.X + (cnt.layout.Bounds.Dx() - r.Dx())
		cnt.layout.Bounds.Max.Y = cnt.layout.Bounds.Min.Y + cnt.layout.ContentSize.Y + (cnt.layout.Bounds.Dy() - r.Dy())
	}

	// close if this is a popup window and elsewhere was clicked
	if (opt&optionPopup) != 0 && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.hoverRoot != cnt {
		cnt.open = false
	}

	c.pushClipRect(cnt.layout.BodyBounds)
	defer c.popClipRect()

	f(ResponseActive, c.currentContainer().layout)
}

func (c *Context) OpenPopup(name string) {
	cnt := c.Container(name)
	// set as hover root so popup isn't closed in begin_window_ex()
	c.nextHoverRoot = cnt
	c.hoverRoot = c.nextHoverRoot
	// position at mouse cursor, open and bring-to-front
	pt := image.Pt(ebiten.CursorPosition())
	cnt.layout.Bounds = image.Rectangle{
		Min: pt,
		Max: pt.Add(image.Pt(1, 1)),
	}
	cnt.open = true
	c.bringToFront(cnt)
}

func (c *Context) Popup(name string, f func(res Response, layout ContainerLayout)) {
	opt := optionPopup | optionAutoSize | optionNoResize | optionNoScroll | optionNoTitle | optionClosed
	c.window(name, "", image.Rectangle{}, opt, f)
}

func (c *Context) panel(name string, opt option, f func(layout ContainerLayout)) {
	id := c.pushID([]byte(name))
	defer c.popID()

	cnt := c.container(id, opt)
	cnt.layout.Bounds = c.layoutNext()
	if (^opt & optionNoFrame) != 0 {
		c.drawFrame(cnt.layout.Bounds, ColorPanelBG)
	}

	c.containerStack = append(c.containerStack, cnt)
	defer c.popContainer()

	c.pushContainerBodyLayout(cnt, cnt.layout.Bounds, opt)
	defer c.popLayout()

	c.pushClipRect(cnt.layout.BodyBounds)
	defer c.popClipRect()

	f(c.currentContainer().layout)
}

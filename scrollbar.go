// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// x = x, y = y, w = w, h = h
func (c *Context) scrollbarVertical(cnt *container, body image.Rectangle, cs image.Point) {
	maxscroll := cs.Y - body.Dy()
	if maxscroll > 0 && body.Dy() > 0 {
		// get sizing / positioning
		base := body
		base.Min.X = body.Max.X
		base.Max.X = base.Min.X + c.style().scrollbarSize

		// handle input
		id := c.idStack.push(idPartFromString("scrollbar-y"))
		_ = c.widgetWithBounds(id, 0, base, func(bounds image.Rectangle, wasFocused bool) EventHandler {
			if c.focus == id && c.pointing.pressed() {
				cnt.layout.ScrollOffset.Y += c.pointingDelta().Y * cs.Y / bounds.Dy()
			}
			// clamp scroll to limits
			cnt.layout.ScrollOffset.Y = clamp(cnt.layout.ScrollOffset.Y, 0, maxscroll)

			// set this as the scroll_target (will get scrolled on mousewheel)
			// if the pointing device is over it
			if c.pointingOver(body) {
				c.scrollTarget = cnt
			}
			return nil
		}, func(bounds image.Rectangle) {
			c.drawFrame(bounds, colorScrollBase)
			thumb := bounds
			thumb.Max.Y = thumb.Min.Y + max(c.style().thumbSize, bounds.Dy()*body.Dy()/cs.Y)
			thumb = thumb.Add(image.Pt(0, cnt.layout.ScrollOffset.Y*(bounds.Dy()-thumb.Dy())/maxscroll))
			c.drawFrame(thumb, colorScrollThumb)
		})
	} else {
		cnt.layout.ScrollOffset.Y = 0
	}
}

// x = y, y = x, w = h, h = w
func (c *Context) scrollbarHorizontal(cnt *container, body image.Rectangle, cs image.Point) {
	maxscroll := cs.X - body.Dx()
	if maxscroll > 0 && body.Dx() > 0 {
		// get sizing / positioning
		base := body
		base.Min.Y = body.Max.Y
		base.Max.Y = base.Min.Y + c.style().scrollbarSize

		// handle input
		id := c.idStack.push(idPartFromString("scrollbar-x"))
		_ = c.widgetWithBounds(id, 0, base, func(bounds image.Rectangle, wasFocused bool) EventHandler {
			if c.focus == id && c.pointing.pressed() {
				cnt.layout.ScrollOffset.X += c.pointingDelta().X * cs.X / bounds.Dx()
			}
			// clamp scroll to limits
			cnt.layout.ScrollOffset.X = clamp(cnt.layout.ScrollOffset.X, 0, maxscroll)

			// set this as the scroll_target (will get scrolled on mousewheel)
			// if the pointing device is over it
			if c.pointingOver(body) {
				c.scrollTarget = cnt
			}
			return nil
		}, func(bounds image.Rectangle) {
			c.drawFrame(bounds, colorScrollBase)
			thumb := bounds
			thumb.Max.X = thumb.Min.X + max(c.style().thumbSize, bounds.Dx()*body.Dx()/cs.X)
			thumb = thumb.Add(image.Pt(cnt.layout.ScrollOffset.X*(bounds.Dx()-thumb.Dx())/maxscroll, 0))
			c.drawFrame(thumb, colorScrollThumb)
		})
	} else {
		cnt.layout.ScrollOffset.X = 0
	}
}

// if `swap` is true, X = Y, Y = X, W = H, H = W
func (c *Context) scrollbar(cnt *container, body image.Rectangle, cs image.Point, swap bool) {
	if swap {
		c.scrollbarHorizontal(cnt, body, cs)
	} else {
		c.scrollbarVertical(cnt, body, cs)
	}
}

func (c *Context) scrollbars(cnt *container, body image.Rectangle) image.Rectangle {
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
	c.scrollbar(cnt, body, cs, false)
	c.scrollbar(cnt, body, cs, true)
	c.popClipRect()
	return body
}

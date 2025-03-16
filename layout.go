// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import "image"

func (c *Context) pushLayout(body image.Rectangle, scroll image.Point) {
	c.layoutStack = append(c.layoutStack, layout{
		body: body.Sub(scroll),
		max:  image.Pt(-0x1000000, -0x1000000),
	})
	c.SetGridLayout(nil, nil)
}

func (c *Context) popLayout() {
	cnt := c.currentContainer()
	layout := c.layout()
	cnt.layout.ContentSize.X = layout.max.X - layout.body.Min.X
	cnt.layout.ContentSize.Y = layout.max.Y - layout.body.Min.Y
	c.layoutStack = c.layoutStack[:len(c.layoutStack)-1]
}

func (c *Context) LayoutColumn(f func()) {
	c.control(0, 0, func(bounds image.Rectangle) Response {
		c.pushLayout(bounds, image.Pt(0, 0))
		defer c.popLayout()
		f()
		b := &c.layoutStack[len(c.layoutStack)-1]
		// inherit position/next_row/max from child layout if they are greater
		a := &c.layoutStack[len(c.layoutStack)-2]
		a.position.X = max(a.position.X, b.position.X+b.body.Min.X-a.body.Min.X)
		a.nextRowY = max(a.nextRowY, b.nextRowY+b.body.Min.Y-a.body.Min.Y)
		a.max.X = max(a.max.X, b.max.X)
		a.max.Y = max(a.max.Y, b.max.Y)
		return 0
	})
}

func (c *Context) SetGridLayout(widths []int, heights []int) {
	layout := c.layout()

	if len(layout.widths) < len(widths) {
		layout.widths = append(layout.widths, make([]int, len(widths)-len(layout.widths))...)
	}
	copy(layout.widths, widths)
	layout.widths = layout.widths[:len(widths)]
	if len(layout.heights) == 0 {
		layout.heights = append(layout.heights, 0) // TODO: This should be -1?
	}

	if len(layout.heights) < len(heights) {
		layout.heights = append(layout.heights, make([]int, len(heights)-len(layout.heights))...)
	}
	copy(layout.heights, heights)
	layout.heights = layout.heights[:len(heights)]
	if len(layout.widths) == 0 {
		layout.widths = append(layout.widths, 0)
	}

	layout.position = image.Pt(layout.indent, layout.nextRowY)
	layout.itemIndex = 0
}

func (c *Context) layoutNext() image.Rectangle {
	layout := c.layout()

	// If the item reaches the end of the row, start a new row with the same rule.
	if layout.itemIndex == len(layout.widths)*len(layout.heights) {
		c.SetGridLayout(layout.widths, layout.heights)
	} else if layout.itemIndex%len(layout.widths) == 0 {
		layout.position = image.Pt(layout.indent, layout.nextRowY)
	}

	// position
	r := image.Rect(layout.position.X, layout.position.Y, layout.position.X, layout.position.Y)

	// size
	if len(layout.widths) > 0 {
		r.Max.X = r.Min.X + layout.widths[layout.itemIndex%len(layout.widths)]
	}
	if len(layout.heights) > 0 {
		r.Max.Y = r.Min.Y + layout.heights[layout.itemIndex/len(layout.widths)]
	}
	if r.Dx() == 0 {
		r.Max.X = r.Min.X + c.style.size.X + c.style.padding*2
	}
	if r.Dy() == 0 {
		r.Max.Y = r.Min.Y + c.style.size.Y + c.style.padding*2
	}
	if r.Dx() < 0 {
		r.Max.X += layout.body.Dx() - r.Min.X + 1
	}
	if r.Dy() < 0 {
		r.Max.Y += layout.body.Dy() - r.Min.Y + 1
	}

	layout.itemIndex++
	// update position
	layout.position.X += r.Dx() + c.style.spacing
	layout.nextRowY = max(layout.nextRowY, r.Max.Y+c.style.spacing)

	// apply body offset
	r = r.Add(layout.body.Min)

	// update max position
	layout.max.X = max(layout.max.X, r.Max.X)
	layout.max.Y = max(layout.max.Y, r.Max.Y)

	return r
}

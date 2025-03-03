// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import "image"

func (c *Context) pushLayout(body image.Rectangle, scroll image.Point) {
	// push()
	c.layoutStack = append(c.layoutStack, layout{
		body: body.Sub(scroll),
		max:  image.Pt(-0x1000000, -0x1000000),
	})
	c.SetLayoutRow([]int{0}, 0)
}

func (c *Context) LayoutColumn(f func()) {
	c.control(0, 0, func(r image.Rectangle) Response {
		c.pushLayout(r, image.Pt(0, 0))
		defer c.popLayout()
		f()
		b := &c.layoutStack[len(c.layoutStack)-1]
		// inherit position/next_row/max from child layout if they are greater
		a := &c.layoutStack[len(c.layoutStack)-2]
		a.position.X = max(a.position.X, b.position.X+b.body.Min.X-a.body.Min.X)
		a.nextRow = max(a.nextRow, b.nextRow+b.body.Min.Y-a.body.Min.Y)
		a.max.X = max(a.max.X, b.max.X)
		a.max.Y = max(a.max.Y, b.max.Y)
		return 0
	})
}

func (c *Context) SetLayoutRow(widths []int, height int) {
	layout := c.layout()

	if len(layout.widths) < len(widths) {
		layout.widths = append(layout.widths, make([]int, len(widths)-len(layout.widths))...)
	}
	copy(layout.widths, widths)
	layout.widths = layout.widths[:len(widths)]

	layout.position = image.Pt(layout.indent, layout.nextRow)
	layout.height = height
	layout.itemIndex = 0
}

func (c *Context) layoutNext() image.Rectangle {
	layout := c.layout()

	// handle next row
	if layout.itemIndex == len(layout.widths) {
		c.SetLayoutRow(layout.widths, layout.height)
	}

	// position
	res := image.Rect(layout.position.X, layout.position.Y, layout.position.X, layout.position.Y)

	// size
	if len(layout.widths) > 0 {
		res.Max.X = res.Min.X + layout.widths[layout.itemIndex]
	}
	res.Max.Y = res.Min.Y + layout.height
	if res.Dx() == 0 {
		res.Max.X = res.Min.X + c.style.size.X + c.style.padding*2
	}
	if res.Dy() == 0 {
		res.Max.Y = res.Min.Y + c.style.size.Y + c.style.padding*2
	}
	if res.Dx() < 0 {
		res.Max.X += layout.body.Dx() - res.Min.X + 1
	}
	if res.Dy() < 0 {
		res.Max.Y += layout.body.Dy() - res.Min.Y + 1
	}

	layout.itemIndex++

	// update position
	layout.position.X += res.Dx() + c.style.spacing
	layout.nextRow = max(layout.nextRow, res.Max.Y+c.style.spacing)

	// apply body offset
	res = res.Add(layout.body.Min)

	// update max position
	layout.max.X = max(layout.max.X, res.Max.X)
	layout.max.Y = max(layout.max.Y, res.Max.Y)

	c.lastRect = res
	return c.lastRect
}

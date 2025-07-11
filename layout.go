// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"errors"
	"image"
)

type layout struct {
	body      image.Rectangle
	position  image.Point
	max       image.Point
	widths    []int
	heights   []int
	itemIndex int
	nextRowY  int
	indent    int
}

func (l *layout) widthInPixels(style *style) int {
	return l.sizeInPixels(l.widths, l.itemIndex%len(l.widths), 8, style.defaultWidth+style.padding*2, l.body.Dx()-l.indent, style)
}

func (l *layout) heightInPixels(style *style) int {
	return l.sizeInPixels(l.heights, l.itemIndex/len(l.widths), 6, style.defaultHeight, l.body.Dy(), style)
}

func (l *layout) sizeInPixels(sizes []int, index int, minSize, defaultSize int, entireSize int, style *style) int {
	s := sizes[index]
	if s > 0 {
		return s
	}
	if s == 0 {
		return defaultSize
	}

	remain := entireSize - (len(sizes)-1)*style.spacing
	var denom int
	for _, s := range sizes {
		if s > 0 {
			remain -= s
		}
		if s == 0 {
			remain -= defaultSize
		}
		if s < 0 {
			denom += -s
		}
	}
	return max(minSize, int(float64(remain)*float64(-s)/float64(denom)))
}

func (c *Context) pushLayout(body image.Rectangle, scroll image.Point, autoResize bool) error {
	c.layoutStack = append(c.layoutStack, layout{
		body:    body.Sub(scroll),
		max:     image.Pt(-0x1000000, -0x1000000),
		widths:  []int{0},
		heights: []int{0},
	})
	if autoResize {
		if err := c.setGridLayout([]int{0}, nil); err != nil {
			return err
		}
	} else {
		if err := c.setGridLayout(nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) popLayout() error {
	cnt := c.currentContainer()
	layout, err := c.layout()
	if err != nil {
		return err
	}
	cnt.layout.ContentSize.X = layout.max.X - layout.body.Min.X
	cnt.layout.ContentSize.Y = layout.max.Y - layout.body.Min.Y
	c.layoutStack = c.layoutStack[:len(c.layoutStack)-1]
	return err
}

// GridCell creates a grid cell with the content defined by the function f.
func (c *Context) GridCell(f func(bounds image.Rectangle)) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if _, err := c.widget(widgetID{}, 0, f, nil, nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (c *Context) layout() (*layout, error) {
	if len(c.layoutStack) == 0 {
		return nil, errors.New("debugui: layout stack is empty; perhaps a window is absent")
	}
	return &c.layoutStack[len(c.layoutStack)-1], nil
}

// SetGridLayout sets the grid layout.
// widths and heights are the sizes of each column and row.
//
// If a size is 0, the default size is used.
// If a size is negative, the size represents the ratio of the remaining space.
// For example, if widths is []int{100, -1}, the first column is 100 pixels and the second column takes the remaining space.
// If widths is []int{100, -1, -2}, the first column is 100 pixels, the second column takes 1/3 of the remaining space, and the third column takes 2/3 of the remaining space.
//
// If widths is nil, one column width with -1 is used for windows, and 0 (the default size) is used for popups.
// If heights is nil, 0 (the default size) is used.
//
// When the number of items exceeds the number of grid cells, a new row starts with the same grid layout.
func (c *Context) SetGridLayout(widths []int, heights []int) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if err := c.setGridLayout(widths, heights); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (c *Context) setGridLayout(widths []int, heights []int) error {
	layout, err := c.layout()
	if err != nil {
		return err
	}

	if len(layout.widths) < len(widths) {
		layout.widths = append(layout.widths, make([]int, len(widths)-len(layout.widths))...)
	}
	copy(layout.widths, widths)
	layout.widths = layout.widths[:len(widths)]
	if len(layout.widths) == 0 {
		layout.widths = append(layout.widths, -1)
	}

	if len(layout.heights) < len(heights) {
		layout.heights = append(layout.heights, make([]int, len(heights)-len(layout.heights))...)
	}
	copy(layout.heights, heights)
	layout.heights = layout.heights[:len(heights)]
	if len(layout.heights) == 0 {
		layout.heights = append(layout.heights, 0) // TODO: This should be -1?
	}

	layout.position = image.Pt(layout.indent, layout.nextRowY)
	layout.itemIndex = 0
	return nil
}

func (c *Context) layoutNext() (image.Rectangle, error) {
	layout, err := c.layout()
	if err != nil {
		return image.Rectangle{}, err
	}

	if len(layout.widths) == 0 {
		panic("not reached")
	}
	if len(layout.heights) == 0 {
		panic("not reached")
	}

	// If the item reaches the end of the row, start a new row with the same rule.
	if layout.itemIndex == len(layout.widths)*len(layout.heights) {
		c.SetGridLayout(layout.widths, layout.heights)
	} else if layout.itemIndex%len(layout.widths) == 0 {
		layout.position = image.Pt(layout.indent, layout.nextRowY)
	}

	// position
	r := image.Rect(layout.position.X, layout.position.Y, layout.position.X, layout.position.Y)

	// size
	r.Max.X = r.Min.X + layout.widthInPixels(c.style())
	r.Max.Y = r.Min.Y + layout.heightInPixels(c.style())

	layout.itemIndex++
	// update position
	layout.position.X += r.Dx() + c.style().spacing
	layout.nextRowY = max(layout.nextRowY, r.Max.Y+c.style().spacing)

	// apply body offset
	r = r.Add(layout.body.Min)

	// update max position
	layout.max.X = max(layout.max.X, r.Max.X)
	layout.max.Y = max(layout.max.Y, r.Max.Y)

	return r, nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "image"

// Header creates a header widget with the given label.
//
// initialExpansion specifies whether the header is initially expanded.
// f is called to render the content of the header.
// The content is only rendered when the header is expanded.
//
// A Header widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) Header(label string, initialExpansion bool, f func()) {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		var opt option
		if initialExpansion {
			opt |= optionExpanded
		}
		if err := c.header(label, false, opt, id, func() error {
			f()
			return nil
		}); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

// TreeNode creates a tree node widget with the given label.
//
// A TreeNode widget is uniquely determined by its call location.
// Function calls made in different locations will create different widgets.
// If you want to generate different widgets with the same function call in a loop (such as a for loop), use [IDScope].
func (c *Context) TreeNode(label string, f func()) {
	pc := caller()
	id := c.idStack.push(idPartFromCaller(pc))
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if err := c.treeNode(label, 0, id, f); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (c *Context) header(label string, isTreeNode bool, opt option, id widgetID, f func() error) error {
	c.SetGridLayout(nil, nil)

	var expanded bool
	toggled := c.currentContainer().toggled(id)
	if (opt & optionExpanded) != 0 {
		expanded = !toggled
	} else {
		expanded = toggled
	}

	e, err := c.widget(id, 0, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		if c.pointing.justPressed() && c.focus == id {
			c.currentContainer().toggle(id)
		}
		if expanded {
			return &eventHandler{}
		}
		return nil
	}, func(bounds image.Rectangle) {
		if isTreeNode {
			if c.hover == id {
				c.drawFrame(bounds, colorButtonHover)
			}
		} else {
			c.drawWidgetFrame(id, bounds, colorButton, 0)
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
			c.style().colors[colorText],
		)
		bounds.Min.X += bounds.Dy() - c.style().padding
		c.drawWidgetText(label, bounds, colorText, 0)
	})
	if err != nil {
		return err
	}
	if e != nil {
		e.On(func() {
			if err := f(); err != nil && c.err == nil {
				c.err = err
			}
		})
	}
	return nil
}

func (c *Context) treeNode(label string, opt option, id widgetID, f func()) error {
	if err := c.header(label, true, opt, id, func() (err error) {
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

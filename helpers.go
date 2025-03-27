// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

func clamp[T int | float64](x, a, b T) T {
	return min(b, max(a, x))
}

func (c *Context) bringToFront(cnt *container) {
	c.lastZIndex++
	cnt.zIndex = c.lastZIndex
}

func (c *Context) Focus() {
	c.setFocus(c.lastID)
}

func (c *Context) setFocus(id widgetID) {
	c.focus = id
	c.keepFocus = true
}

func (c *Context) addUsedContainer(id widgetID) {
	if c.usedContainers == nil {
		c.usedContainers = map[widgetID]struct{}{}
	}
	c.usedContainers[id] = struct{}{}
}

func (c *Context) wrapError(f func() error) {
	if c.err != nil {
		return
	}
	c.err = f()
}

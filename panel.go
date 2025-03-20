// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

func (c *Context) Panel(name string, f func(layout ContainerLayout)) {
	c.panel(name, 0, f)
}

func (c *Context) panel(name string, opt option, f func(layout ContainerLayout)) {
	id := c.idFromGlobalUniqueString(name)

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

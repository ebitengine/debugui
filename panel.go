// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

func (c *Context) Panel(name string, f func(layout ContainerLayout)) {
	pc := caller()
	c.wrapError(func() error {
		if err := c.panel(name, 0, pc, f); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) panel(name string, opt option, callerPC uintptr, f func(layout ContainerLayout)) (err error) {
	id := c.idFromGlobalUniqueString(name)

	cnt := c.container(id, opt)
	l, err := c.layoutNext()
	if err != nil {
		return err
	}
	cnt.layout.Bounds = l
	if (^opt & optionNoFrame) != 0 {
		c.drawFrame(cnt.layout.Bounds, ColorPanelBG)
	}

	c.containerStack = append(c.containerStack, cnt)
	defer c.popContainer()

	c.pushContainerBodyLayout(cnt, cnt.layout.Bounds, opt, callerPC)
	defer func() {
		if err2 := c.popLayout(); err2 != nil && err == nil {
			err = err2
		}
	}()

	c.pushClipRect(cnt.layout.BodyBounds)
	defer c.popClipRect()

	f(c.currentContainer().layout)
	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

func (c *Context) Panel(name string, f func(layout ContainerLayout)) {
	c.wrapError(func() error {
		if err := c.panel(name, 0, f); err != nil {
			return err
		}
		return nil
	})
}

func (c *Context) panel(name string, opt option, f func(layout ContainerLayout)) (err error) {
	id := c.idFromString(name)
	c.idScopeFromControlID(id, func() {
		err = c.doPanel(opt, id, f)
	})
	return
}

func (c *Context) doPanel(opt option, id controlID, f func(layout ContainerLayout)) (err error) {
	cnt := c.container(id, opt)
	l, err := c.layoutNext()
	if err != nil {
		return err
	}
	cnt.layout.Bounds = l
	if (^opt & optionNoFrame) != 0 {
		c.drawFrame(cnt.layout.Bounds, colorPanelBG)
	}

	c.pushContainer(cnt)
	defer c.popContainer()

	if err := c.pushContainerBodyLayout(cnt, cnt.layout.Bounds, opt); err != nil {
		return err
	}
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

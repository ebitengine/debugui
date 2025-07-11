// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

// Panel creates a new panel with the contents defined by the function f.
// Panel can have scroll bars, and the contents of the panel can be scrolled.
func (c *Context) Panel(f func(layout ContainerLayout)) {
	pc := caller()
	idPart := idPartFromCaller(pc)
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		if err := c.panel(0, idPart, f); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (c *Context) panel(opt option, idPart string, f func(layout ContainerLayout)) (err error) {
	c.idScopeFromIDPart(idPart, func(id widgetID) {
		err = c.doPanel(opt, id, f)
	})
	return
}

func (c *Context) doPanel(opt option, id widgetID, f func(layout ContainerLayout)) (err error) {
	cnt := c.container(id, opt)
	l, err := c.layoutNext()
	if err != nil {
		return err
	}
	cnt.layout.Bounds = l
	if (^opt & optionNoFrame) != 0 {
		c.drawFrame(cnt.layout.Bounds, colorPanelBG)
	}

	c.pushContainer(cnt, false)
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

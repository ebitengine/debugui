// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

const EmptyWidgetID = emptyWidgetID

func (c *Context) IDFromCaller() WidgetID {
	pc := caller()
	return c.idFromCaller(pc)
}

func (d *DebugUI) ContainerCounter() int {
	return len(d.ctx.idToContainer)
}

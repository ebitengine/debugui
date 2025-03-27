// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

type ControlID = controlID

const EmptyControlID = emptyControlID

func (c *Context) IDFromCaller() ControlID {
	pc := caller()
	return c.idFromCaller(pc)
}

func (d *DebugUI) ContainerCounter() int {
	return len(d.ctx.idToContainer)
}

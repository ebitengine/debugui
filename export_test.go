// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

type ControlID = controlID

const EmptyControlID = emptyControlID

func (c *Context) ButtonID(label string) ControlID {
	pc := caller()
	var id controlID
	c.wrapError(func() error {
		var err error
		id, _, err = c.button(label, optionAlignCenter, pc)
		if err != nil {
			return err
		}
		return nil
	})
	return id
}

func (d *DebugUI) ContainerCounter() int {
	return len(d.ctx.idToContainer)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

type ControlID = controlID

func (c *Context) ButtonID(label string) ControlID {
	id, _ := c.button(label, optionAlignCenter)
	return id
}

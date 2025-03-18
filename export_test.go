// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "strings"

type ControlID = controlID

func (c *Context) ButtonID(label string) ControlID {
	label, idStr, _ := strings.Cut(label, idSeparator)
	id, _ := c.button(label, idStr, optionAlignCenter)
	return id
}

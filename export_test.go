// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import "github.com/ebitengine/debugui/internal/caller"

type ControlID = controlID

func (c *Context) ButtonID(label string) ControlID {
	pc := caller.Caller()
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

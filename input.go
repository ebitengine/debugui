// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

func (c *Context) inputScroll(x, y int) {
	c.scrollDelta.X += x
	c.scrollDelta.Y += y
}

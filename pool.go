// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

func (c *Context) poolInit(items []poolItem, id controlID) int {
	f := c.tick
	n := -1
	for i := 0; i < len(items); i++ {
		if items[i].lastUpdate < f {
			f = items[i].lastUpdate
			n = i
		}
	}
	items[n].id = id
	c.poolUpdate(items, n)
	return n
}

// returns the index of an ID in the pool. returns -1 if it is not found
func (c *Context) poolGet(items []poolItem, id controlID) int {
	for i := 0; i < len(items); i++ {
		if items[i].id == id {
			return i
		}
	}
	return -1
}

func (c *Context) poolUpdate(items []poolItem, idx int) {
	items[idx].lastUpdate = c.tick
}

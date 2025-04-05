// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"fmt"
	"runtime"
	"slices"
)

// caller returns a program counter of the caller.
func caller() uintptr {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return 0
	}
	return pc
}

// IDScope creates a new scope for widget IDs.
// IDScope creates a unique scope based on the caller's position and the given name string.
//
// IDScope is useful when you want to create multiple widgets at the same position e.g. in a for loop.
func (c *Context) IDScope(name string, f func()) {
	pc := caller()
	c.idStack = append(c.idStack, widgetID(fmt.Sprintf("caller:%d", pc)))
	c.idStack = append(c.idStack, widgetID(fmt.Sprintf("string:%q", name)))
	defer func() {
		c.idStack = slices.Delete(c.idStack, len(c.idStack)-2, len(c.idStack))
	}()
	f()
}

func (c *Context) idScopeFromID(id widgetID, f func()) {
	c.idStack = append(c.idStack, id)
	defer func() {
		c.idStack = slices.Delete(c.idStack, len(c.idStack)-1, len(c.idStack))
	}()
	f()
}

func (c *Context) idScopeToWidgetID() widgetID {
	var newID widgetID
	for _, id := range c.idStack {
		if len(newID) > 0 {
			newID += ":"
		}
		newID += "[" + id + "]"
	}
	return newID
}

func (c *Context) idFromString(str string) widgetID {
	newID := c.idScopeToWidgetID()
	if len(newID) > 0 {
		newID += ":"
	}
	newID += widgetID(fmt.Sprintf("string:%q", str))
	return newID
}

// idFromCaller returns a hash value based on the caller's file and line number.
func (c *Context) idFromCaller(callerPC uintptr) widgetID {
	newID := c.idScopeToWidgetID()
	if len(newID) > 0 {
		newID += ":"
	}
	newID += widgetID(fmt.Sprintf("caller:%d", callerPC))
	return newID
}

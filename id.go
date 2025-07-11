// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"fmt"
	"runtime"
)

// caller returns a program counter of the caller.
func caller() uintptr {
	var pcs [1]uintptr
	n := runtime.Callers(3, pcs[:])
	if n == 0 {
		return 0
	}
	return pcs[0]
}

// Loop creates a loop to iterate by the given count.
// Loop creates a unique ID scope for each iteration.
func (c *Context) Loop(count int, f func(i int)) {
	pc := caller()
	c.idStack = c.idStack.push(idPartFromCaller(pc))
	defer func() {
		c.idStack = c.idStack.pop()
	}()
	for i := range count {
		c.idStack = c.idStack.push(idPartFromInt(i))
		f(i)
		c.idStack = c.idStack.pop()
	}
}

// IDScope creates a new scope for widget IDs.
// IDScope creates a unique scope based on the caller's position and the given name string.
//
// IDScope is useful when you want to create multiple widgets at the same position e.g. in a for loop.
//
// IDScope is a low level API. For a simple loop, use [Loop] instead.
func (c *Context) IDScope(name string, f func()) {
	pc := caller()
	c.idStack = c.idStack.push(idPartFromCaller(pc))
	c.idStack = c.idStack.push(idPartFromString(name))
	defer func() {
		c.idStack = c.idStack.pop().pop()
	}()
	f()
}

func (c *Context) idScopeFromIDPart(idPart string, f func(id widgetID)) {
	c.idStack = c.idStack.push(idPart)
	defer func() {
		c.idStack = c.idStack.pop()
	}()
	f(c.idStack)
}

func idPartFromString(str string) string {
	return theStringIDCache.get(str)
}

func idPartFromInt(i int) string {
	return theIntIDCache.get(i)
}

func idPartFromCaller(callerPC uintptr) string {
	return theCallerIDCache.get(callerPC)
}

type idCache[T comparable] struct {
	values map[T]string
	prefix string
}

func (c *idCache[T]) get(id T) string {
	if idStr, ok := c.values[id]; ok {
		return idStr
	}
	if c.values == nil {
		c.values = map[T]string{}
	}
	idStr := fmt.Sprintf("%s:%v", c.prefix, id)
	c.values[id] = idStr
	return idStr
}

var (
	theStringIDCache = idCache[string]{prefix: "string"}
	theIntIDCache    = idCache[int]{prefix: "number"}
	theCallerIDCache = idCache[uintptr]{prefix: "caller"}
)

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"errors"
	"fmt"
	"image"
	"runtime"
	"slices"
	"sort"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func clamp[T int | float64](x, a, b T) T {
	return min(b, max(a, x))
}

// caller returns a program counter of the caller.
func caller() uintptr {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return 0
	}
	return pc
}

func (c *Context) idFromGlobalUniquePointer(pointer unsafe.Pointer) controlID {
	return controlID(fmt.Sprintf("!pointer:%p", pointer))
}

func (c *Context) idFromGlobalUniqueString(str string) controlID {
	return controlID(fmt.Sprintf("!string:%s", str))
}

// idFromCaller returns a hash value based on the caller's file and line number.
func (c *Context) idFromCaller(callerPC uintptr, str string) controlID {
	if len(str) > 0 {
		return controlID(fmt.Sprintf("!caller:%d:%s", callerPC, str))
	}
	return controlID(fmt.Sprintf("!caller:%d", callerPC))
}

func (c *Context) popContainer() {
	c.containerStack = c.containerStack[:len(c.containerStack)-1]
}

func (c *Context) currentContainer() *container {
	return c.containerStack[len(c.containerStack)-1]
}

func (c *Context) SetScroll(scroll image.Point) {
	c.currentContainer().layout.ScrollOffset = scroll
}

func (c *Context) container(id controlID, opt option) *container {
	// try to get existing container from pool
	if container, ok := c.idToContainer[id]; ok {
		if !container.open && (^opt&optionClosed) == 0 {
			delete(c.idToContainer, id)
		}
		return container
	}

	if (opt & optionClosed) != 0 {
		return nil
	}

	// container not found in pool: init new container
	if c.idToContainer == nil {
		c.idToContainer = map[controlID]*container{}
	}
	cnt := &container{}
	c.idToContainer[id] = cnt
	cnt.headIdx = -1
	cnt.tailIdx = -1
	cnt.open = true
	c.bringToFront(cnt)
	return cnt
}

func (c *Context) bringToFront(cnt *container) {
	c.lastZIndex++
	cnt.zIndex = c.lastZIndex
}

func (c *Context) Focus() {
	c.setFocus(c.lastID)
}

func (c *Context) setFocus(id controlID) {
	c.focus = id
	c.keepFocus = true
}

func (c *Context) update(f func(ctx *Context) error) (err error) {
	if c.err != nil {
		return c.err
	}

	c.begin()
	defer func() {
		if err2 := c.end(); err2 != nil && err == nil {
			err = err2
		}
	}()

	if err := f(c); err != nil {
		return err
	}
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *Context) begin() {
	c.commandList = slices.Delete(c.commandList, 0, len(c.commandList))
	c.rootList = slices.Delete(c.rootList, 0, len(c.rootList))
	c.scrollTarget = nil
	c.hoverRoot = c.nextHoverRoot
	c.nextHoverRoot = nil

	c.tick++
}

func (c *Context) mouseDelta() image.Point {
	return c.cursorPosition().Sub(c.lastMousePos)
}

func (c *Context) cursorPosition() image.Point {
	p := image.Pt(ebiten.CursorPosition())
	p.X /= c.Scale()
	p.Y /= c.Scale()
	return p
}

func (c *Context) end() error {
	// check stacks
	if len(c.containerStack) > 0 {
		return errors.New("debugui: container stack not empty")
	}
	if len(c.clipStack) > 0 {
		return errors.New("debugui: clip stack not empty")
	}
	if len(c.layoutStack) > 0 {
		return errors.New("debugui: layout stack not empty")
	}

	// handle scroll input
	if c.scrollTarget != nil {
		wx, wy := ebiten.Wheel()
		c.scrollTarget.layout.ScrollOffset.X += int(wx * -30)
		c.scrollTarget.layout.ScrollOffset.Y += int(wy * -30)
	}

	// unset focus if focus id was not touched this frame
	if !c.keepFocus {
		c.focus = emptyControlID
	}
	c.keepFocus = false

	// bring hover root to front if mouse was pressed
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.nextHoverRoot != nil &&
		c.nextHoverRoot.zIndex < c.lastZIndex &&
		c.nextHoverRoot.zIndex >= 0 {
		c.bringToFront(c.nextHoverRoot)
	}

	// reset input state
	c.lastMousePos = c.cursorPosition()

	// sort root containers by zindex
	sort.SliceStable(c.rootList, func(i, j int) bool {
		return c.rootList[i].zIndex < c.rootList[j].zIndex
	})

	// set root container jump commands
	for i := range c.rootList {
		cnt := c.rootList[i]
		// if this is the first container then make the first command jump to it.
		// otherwise set the previous container's tail to jump to this one
		if i == 0 {
			cmd := c.commandList[0]
			if cmd.typ != commandJump {
				panic("debugui: expected jump command")
			}
			cmd.jump.dstIdx = cnt.headIdx + 1
			if cnt.headIdx >= len(c.commandList) {
				panic("debugui: invalid head index")
			}
		} else {
			prev := c.rootList[i-1]
			c.commandList[prev.tailIdx].jump.dstIdx = cnt.headIdx + 1
		}
		// make the last container's tail jump to the end of command list
		if i == len(c.rootList)-1 {
			c.commandList[cnt.tailIdx].jump.dstIdx = len(c.commandList)
		}
	}

	return nil
}

func (c *Context) wrapError(f func() error) {
	if c.err != nil {
		return
	}
	c.err = f()
}

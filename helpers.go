// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
	"sort"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func clamp[T int | float64](x, a, b T) T {
	return min(b, max(a, x))
}

func fnv1a(init controlID, data []byte) controlID {
	h := init
	for i := 0; i < len(data); i++ {
		h = (h ^ controlID(data[i])) * 1099511628211
	}
	return h
}

func ptrToBytes(ptr unsafe.Pointer) []byte {
	slice := unsafe.Slice((*byte)(unsafe.Pointer(&ptr)), unsafe.Sizeof(ptr))

	// `slice` points to `ptr`, which is currently allocated on the stack.
	// after this function returns, `slice` will point to freed memory, so
	// we need to copy it to the heap for this to be safe
	heapSlice := make([]byte, len(slice))
	copy(heapSlice, slice)
	return heapSlice
}

// idFromBytes returns a hash value based on the data and the last ID on the stack.
func (c *Context) idFromBytes(data []byte) controlID {
	if len(data) == 0 {
		return 0
	}

	const (
		// hashInitial is the initial value for the FNV-1a hash.
		// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
		hashInitial = 14695981039346656037
	)

	var init controlID = hashInitial
	if len(c.idStack) > 0 {
		init = c.idStack[len(c.idStack)-1]
	}
	id := fnv1a(init, data)
	c.lastID = id
	return id
}

func (c *Context) pushID(data []byte) controlID {
	// push()
	id := c.idFromBytes(data)
	c.idStack = append(c.idStack, id)
	return id
}

func (c *Context) popID() {
	c.idStack = c.idStack[:len(c.idStack)-1]
}

func (c *Context) pushClipRect(rect image.Rectangle) {
	last := c.clipRect()
	// push()
	c.clipStack = append(c.clipStack, rect.Intersect(last))
}

func (c *Context) popClipRect() {
	c.clipStack = c.clipStack[:len(c.clipStack)-1]
}

func (c *Context) clipRect() image.Rectangle {
	return c.clipStack[len(c.clipStack)-1]
}

func (c *Context) checkClip(r image.Rectangle) int {
	cr := c.clipRect()
	if !r.Overlaps(cr) {
		return clipAll
	}
	if r.In(cr) {
		return 0
	}
	return clipPart
}

func (c *Context) layout() *layout {
	return &c.layoutStack[len(c.layoutStack)-1]
}

func (c *Context) popContainer() {
	c.containerStack = c.containerStack[:len(c.containerStack)-1]
}

func (c *Context) popLayout() {
	cnt := c.currentContainer()
	layout := c.layout()
	cnt.layout.ContentSize.X = layout.max.X - layout.body.Min.X
	cnt.layout.ContentSize.Y = layout.max.Y - layout.body.Min.Y
	c.layoutStack = c.layoutStack[:len(c.layoutStack)-1]
}

func (c *Context) currentContainer() *container {
	return c.containerStack[len(c.containerStack)-1]
}

func (c *Context) SetScroll(scroll image.Point) {
	c.currentContainer().layout.Scroll = scroll
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
	*cnt = container{}
	cnt.headIdx = -1
	cnt.tailIdx = -1
	cnt.open = true
	c.bringToFront(cnt)
	return cnt
}

func (c *Context) Container(name string) *container {
	id := c.idFromBytes([]byte(name))
	return c.container(id, 0)
}

func (c *Context) bringToFront(cnt *container) {
	c.lastZIndex++
	cnt.zIndex = c.lastZIndex
}

func (c *Context) SetFocus() {
	c.setFocus(c.lastID)
}

func (c *Context) setFocus(id controlID) {
	c.focus = id
	c.keepFocus = true
}

func (c *Context) update(f func(ctx *Context)) {
	c.begin()
	defer c.end()
	f(c)
}

func (c *Context) begin() {
	c.updateInput()

	c.commandList = c.commandList[:0]
	c.rootList = c.rootList[:0]
	c.scrollTarget = nil
	c.hoverRoot = c.nextHoverRoot
	c.nextHoverRoot = nil

	c.tick++
}

func (c *Context) mouseDelta() image.Point {
	p := image.Pt(ebiten.CursorPosition())
	return p.Sub(c.lastMousePos)
}

func (c *Context) end() {
	// check stacks
	if len(c.containerStack) > 0 {
		panic("container stack not empty")
	}
	if len(c.clipStack) > 0 {
		panic("clip stack not empty")
	}
	if len(c.idStack) > 0 {
		panic("id stack not empty")
	}
	if len(c.layoutStack) > 0 {
		panic("layout stack not empty")
	}

	// handle scroll input
	if c.scrollTarget != nil {
		c.scrollTarget.layout.Scroll.X += c.scrollDelta.X
		c.scrollTarget.layout.Scroll.Y += c.scrollDelta.Y
	}

	// unset focus if focus id was not touched this frame
	if !c.keepFocus {
		c.focus = 0
	}
	c.keepFocus = false

	// bring hover root to front if mouse was pressed
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && c.nextHoverRoot != nil &&
		c.nextHoverRoot.zIndex < c.lastZIndex &&
		c.nextHoverRoot.zIndex >= 0 {
		c.bringToFront(c.nextHoverRoot)
	}

	// reset input state
	c.scrollDelta = image.Pt(0, 0)
	c.lastMousePos = image.Pt(ebiten.CursorPosition())

	// sort root containers by zindex
	sort.SliceStable(c.rootList, func(i, j int) bool {
		return c.rootList[i].zIndex < c.rootList[j].zIndex
	})

	// set root container jump commands
	for i := 0; i < len(c.rootList); i++ {
		cnt := c.rootList[i]
		// if this is the first container then make the first command jump to it.
		// otherwise set the previous container's tail to jump to this one
		if i == 0 {
			cmd := c.commandList[0]
			if cmd.typ != commandJump {
				panic("expected jump command")
			}
			cmd.jump.dstIdx = cnt.headIdx + 1
			if cnt.headIdx >= len(c.commandList) {
				panic("invalid head index")
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
}

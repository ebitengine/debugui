// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

type controlID uint64

type baseCommand struct {
	typ int
}

type jumpCommand struct {
	dstIdx int
}

type clipCommand struct {
	rect image.Rectangle
}

type rectCommand struct {
	rect  image.Rectangle
	color color.Color
}

type textCommand struct {
	pos   image.Point
	color color.Color
	str   string
}

type iconCommand struct {
	rect  image.Rectangle
	icon  icon
	color color.Color
}

type drawCommand struct {
	f func(screen *ebiten.Image)
}

type layout struct {
	body      image.Rectangle
	position  image.Point
	height    int
	max       image.Point
	widths    []int
	itemIndex int
	nextRowY  int
	indent    int
}

type command struct {
	typ  int
	idx  int
	base baseCommand // type 0 (TODO)
	jump jumpCommand // type 1
	clip clipCommand // type 2
	rect rectCommand // type 3
	text textCommand // type 4
	icon iconCommand // type 5
	draw drawCommand // type 6
}

type container struct {
	layout    Layout
	headIdx   int
	tailIdx   int
	zIndex    int
	open      bool
	collapsed bool
}

type Layout struct {
	Rect        image.Rectangle
	Body        image.Rectangle
	ContentSize image.Point
	Scroll      image.Point
}

type style struct {
	size          image.Point
	padding       int
	spacing       int
	indent        int
	titleHeight   int
	scrollbarSize int
	thumbSize     int
	colors        [ColorMax + 1]color.RGBA
}

type Context struct {
	// core state

	style         *style
	hover         controlID
	focus         controlID
	lastID        controlID
	lastRect      image.Rectangle
	lastZIndex    int
	keepFocus     bool
	tick          int
	hoverRoot     *container
	nextHoverRoot *container
	scrollTarget  *container
	numberEditBuf string
	numberEdit    controlID

	// stacks

	commandList    []*command
	rootList       []*container
	containerStack []*container
	clipStack      []image.Rectangle
	idStack        []controlID
	layoutStack    []layout

	// retained state pools

	idToContainer map[controlID]*container
	toggledIDs    map[controlID]struct{}

	// input state

	lastMousePos image.Point

	textFields map[controlID]*textinput.Field
}

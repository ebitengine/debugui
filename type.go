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
	layout    ContainerLayout
	headIdx   int
	tailIdx   int
	zIndex    int
	open      bool
	collapsed bool
}

// ContainerLayout represents the layout of a container control.
type ContainerLayout struct {
	// Bounds is the bounds of the control.
	Bounds image.Rectangle

	// BodyBounds is the bounds of the body area of the container.
	BodyBounds image.Rectangle

	// ContentSize is the size of the content.
	// ContentSize can be larger than Bounds or BodyBounds. In this case, the control should be scrollable.
	ContentSize image.Point

	// ScrollOffset is the offset of the scroll.
	ScrollOffset image.Point
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
	layoutStack    []layout

	// retained state pools

	idToContainer map[controlID]*container
	toggledIDs    map[controlID]struct{}

	// input state

	lastMousePos image.Point

	textFields map[controlID]*textinput.Field
}

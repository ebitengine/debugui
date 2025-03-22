// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
)

type Context struct {
	scaleMinus1   int
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

	commandList    []*command
	rootList       []*container
	containerStack []*container
	clipStack      []image.Rectangle
	layoutStack    []layout

	idToContainer map[controlID]*container

	lastMousePos image.Point

	err error
}

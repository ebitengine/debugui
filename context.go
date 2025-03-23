// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
)

type Context struct {
	scaleMinus1     int
	hover           controlID
	focus           controlID
	lastID          controlID
	lastTextFieldID controlID
	lastZIndex      int
	keepFocus       bool
	hoverRoot       *container
	nextHoverRoot   *container
	scrollTarget    *container
	numberEditBuf   string
	numberEdit      controlID

	idStack        []controlID
	commandList    []*command
	rootList       []*container
	containerStack []*container
	usedContainers map[controlID]struct{}
	clipStack      []image.Rectangle
	layoutStack    []layout

	idToContainer map[controlID]*container

	lastMousePos image.Point

	err error
}

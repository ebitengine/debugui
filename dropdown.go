// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"time"
)

// Dropdown creates a dropdown menu widget that allows users to select from a list of options.
// selectedIndex is a pointer to the currently selected option index (0-based).
// options is a slice of strings representing the available choices.
// Returns an EventHandler that triggers when the selection changes.
func (c *Context) Dropdown(selectedIndex *int, options []string) EventHandler {
	pc := caller()
	idPart := idPartFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.dropdown(selectedIndex, options, idPart)
	})
}

func (c *Context) dropdown(selectedIndex *int, options []string, idPart string) (EventHandler, error) {
	if selectedIndex == nil || len(options) == 0 {
		return &nullEventHandler{}, nil
	}
	if *selectedIndex < 0 || *selectedIndex >= len(options) {
		*selectedIndex = 0
	}
	last := *selectedIndex

	buttonBounds, err := c.layoutNext()
	if err != nil {
		return nil, err
	}

	id := c.idStack.push(idPart)
	dropdownContainer := c.container(id, 0)

	if dropdownContainer.dropdownCloseTimer.expired() {
		dropdownContainer.open = false
		dropdownContainer.dropdownCloseTimer.stop()
	}

	if dropdownContainer.layout.Bounds.Empty() {
		dropdownContainer.open = false
	}

	dropdownContainer.layout.Bounds = c.dropdownBounds(buttonBounds, len(options))

	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		windowOptions := optionNoResize | optionNoTitle

		if err := c.window("", image.Rectangle{}, windowOptions, idPart, func(layout ContainerLayout) {
			if cnt := c.container(id, 0); cnt != nil {
				if cnt.open {
					c.bringToFront(cnt)
				}
			}
			c.SetGridLayout([]int{-1}, nil)

			c.Loop(len(options), func(i int) {
				option := options[i]
				c.Button(option).On(func() {
					*selectedIndex = i
					if cnt := c.container(id, 0); cnt != nil {
						cnt.dropdownCloseTimer.start(c.now, 100*time.Millisecond, 0)
					}
				})
			})
		}); err != nil {
			return nil, err
		}
		return nil, nil
	})

	e := c.widgetWithBounds(id, optionAlignCenter, buttonBounds, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler

		dropdownContainer := c.container(id, 0)
		// Manual "click outside to close" and dropdown toggle, trying to do this in the container.go had lots of issues
		if dropdownContainer.open && c.pointingJustPressed() {
			clickPos := c.pointingPosition()
			clickInButton := clickPos.In(bounds)
			clickInDropdown := clickPos.In(dropdownContainer.layout.Bounds)

			if !clickInButton && !clickInDropdown {
				// Only close immediately if there's no close delay active
				if !dropdownContainer.dropdownCloseTimer.active() {
					dropdownContainer.open = false
				}
			}
		}

		if c.pointingJustPressed() && c.focus == id {
			if dropdownContainer.open {
				// Close the dropdown immediately and cancel any pending delay
				dropdownContainer.open = false
				dropdownContainer.dropdownCloseTimer.stop()
			} else {
				// Open the dropdown and cancel any pending close delay
				dropdownContainer.open = true
				dropdownContainer.dropdownCloseTimer.stop()
			}
		}
		if last != *selectedIndex {
			e = &eventHandler{}
		}

		return e
	}, func(bounds image.Rectangle) {
		if !c.currentContainer().layout.BodyBounds.Overlaps(bounds) {
			return
		}
		c.drawWidgetFrame(id, bounds, colorButton, optionAlignCenter)

		arrowWidth := bounds.Dy()
		textBounds := bounds
		textBounds.Max.X -= arrowWidth
		c.drawWidgetText(options[*selectedIndex], textBounds, colorText, optionAlignCenter)

		arrowBounds := image.Rect(bounds.Max.X-arrowWidth, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
		icon := iconDown
		if c.container(id, 0).open {
			icon = iconUp
		}
		c.drawIcon(icon, arrowBounds, c.style().colors[colorText])
	})
	return e, nil
}

func (c *Context) dropdownBounds(buttonBounds image.Rectangle, optionCount int) image.Rectangle {
	st := c.style()
	height := min(optionCount*(st.defaultHeight+st.spacing)-st.spacing+st.padding*2, st.defaultHeight*12)
	width := buttonBounds.Dx()
	pos := image.Pt(buttonBounds.Min.X, buttonBounds.Max.Y)

	if c.screenWidth > 0 {
		screenWidth := c.screenWidth / c.Scale()
		width = min(width, screenWidth)
		pos.X = clamp(pos.X, 0, screenWidth-width)
	}
	if c.screenHeight > 0 {
		screenHeight := c.screenHeight / c.Scale()
		above := clamp(buttonBounds.Min.Y, 0, screenHeight)
		below := screenHeight - clamp(buttonBounds.Max.Y, 0, screenHeight)
		if height > below && above > below {
			height = min(height, above)
			pos.Y = above - height
		} else {
			height = min(height, below)
			pos.Y = screenHeight - below
		}
	}
	return image.Rectangle{
		Min: pos,
		Max: pos.Add(image.Pt(width, height)),
	}
}

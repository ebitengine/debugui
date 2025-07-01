// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
)

// DropdownID is the ID of a dropdown menu container.
type DropdownID widgetID

// Dropdown creates a dropdown menu widget that allows users to select from a list of options.
// selectedIndex is a pointer to the currently selected option index (0-based).
// options is a slice of strings representing the available choices.
// Returns an EventHandler that triggers when the selection changes.
func (c *Context) Dropdown(selectedIndex *int, options []string) EventHandler {
	pc := caller()
	id := c.idFromCaller(pc)
	return c.wrapEventHandlerAndError(func() (EventHandler, error) {
		return c.dropdown(selectedIndex, options, id)
	})
}

func (c *Context) dropdown(selectedIndex *int, options []string, id widgetID) (EventHandler, error) {
	if selectedIndex == nil || len(options) == 0 {
		// If no options or selectedIndex is nil, return a null event handler
		return &nullEventHandler{}, nil
	}
	// Clamp selectedIndex to valid range to prevent out-of-bounds access
	if *selectedIndex < 0 || *selectedIndex >= len(options) {
		*selectedIndex = 0
	}
	last := *selectedIndex

	dropdownID := DropdownID(c.idFromString("dropdown:" + string(id)))

	// Ensure dropdown container always exists (create it if needed)

	dropdownContainer := c.container(widgetID(dropdownID), 0)

	// Start with the dropdown closed

	if dropdownContainer.layout.Bounds.Empty() {
		dropdownContainer.open = false
	}

	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		windowOptions := optionDropdown | optionNoResize | optionNoTitle

		if err := c.window("", image.Rectangle{}, windowOptions, widgetID(dropdownID), func(layout ContainerLayout) {
			// Ensure dropdown container reference is fresh for each render
			if cnt := c.container(widgetID(dropdownID), 0); cnt != nil {
				if cnt.open {
					c.bringToFront(cnt)
				}
			}

			// full width dropdown
			c.SetGridLayout([]int{-1}, nil)

			// Render each dropdown option as a clickable button
			for i, option := range options {
				c.IDScope(option, func() {
					isSelected := i == *selectedIndex

					// Highlight the currently selected option
					var buttonColor int
					if isSelected {
						buttonColor = colorButtonFocus
					} else {
						buttonColor = colorButton
					}

					// Create clickable button widget for this option
					pc := caller()
					buttonID := c.idFromCaller(pc)
					var wasPressed bool

					_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
						e, err := c.widget(buttonID, optionAlignCenter, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
							var e EventHandler

							if c.pointing.justPressed() && c.focus == buttonID {
								// Handle option selection
								wasPressed = true
								e = &eventHandler{}
							}

							return e
						}, func(bounds image.Rectangle) {
							// Draw the option button with appropriate styling
							c.drawWidgetFrame(buttonID, bounds, buttonColor, optionAlignCenter)
							if len(option) > 0 {
								c.drawWidgetText(option, bounds, colorText, optionAlignCenter)
							}
						})
						return e, err
					})

					// Handle option selection: update index and close dropdown
					if wasPressed {
						*selectedIndex = i
						if cnt := c.container(widgetID(dropdownID), 0); cnt != nil {
							cnt.open = false // Close the dropdown when an option is selected
						}
					}
				})
			}
		}); err != nil {
			return nil, err
		}
		return nil, nil
	})

	// Create the main dropdown button that toggles the menu
	return c.widget(id, optionAlignCenter, nil, func(bounds image.Rectangle, wasFocused bool) EventHandler {
		var e EventHandler

		dropdownContainer := c.container(widgetID(dropdownID), 0)
		// Manual "click outside to close" and dropdown toggle, trying to do this in the container.go had lots of issues
		if dropdownContainer.open && c.pointing.justPressed() {
			clickPos := c.pointingPosition()
			clickInButton := clickPos.In(bounds)
			clickInDropdown := clickPos.In(dropdownContainer.layout.Bounds)

			if !clickInButton && !clickInDropdown {
				dropdownContainer.open = false
			}
		}

		// Toggle dropdown when button is clicked
		if c.pointing.justPressed() && c.focus == id {
			// Check if dropdown container exists and its state

			isOpen := dropdownContainer.open

			if isOpen {
				// Close the dropdown
				dropdownContainer.open = false
			} else {
				// Store the current state before opening, made in some desperate attempts to avoid feedback loops
				wasClosedBefore := !dropdownContainer.open

				// Open the dropdown
				dropdownContainer.open = true

				// Position dropdown directly below button with proper width
				if wasClosedBefore {
					dropdownPos := image.Pt(bounds.Min.X, bounds.Max.Y)
					buttonWidth := bounds.Dx()
					optionHeight := c.style().defaultHeight + c.style().padding
					estimatedHeight := len(options) * optionHeight
					dropdownContainer.layout.Bounds = image.Rectangle{
						Min: dropdownPos,
						Max: dropdownPos.Add(image.Pt(buttonWidth, estimatedHeight)),
					}
				}
			}
		}

		// Generate event if user changed selection
		if last != *selectedIndex {
			e = &eventHandler{}
		}

		return e
	}, func(bounds image.Rectangle) {
		// Draw the dropdown button appearance
		c.drawWidgetFrame(id, bounds, colorButton, optionAlignCenter)

		// Show currently selected text (reserve space for arrow - use widget height for square arrow area)
		arrowWidth := bounds.Dy()
		textBounds := bounds
		textBounds.Max.X -= arrowWidth
		c.drawWidgetText(options[*selectedIndex], textBounds, colorText, optionAlignCenter)

		// Draw dropdown arrow indicator (up/down based on current state)
		arrowBounds := image.Rect(bounds.Max.X-arrowWidth, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
		icon := iconDown
		if c.container(widgetID(dropdownID), 0).open {
			icon = iconUp
		}
		c.drawIcon(icon, arrowBounds, c.style().colors[colorText])
	})
}

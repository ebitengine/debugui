// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image/color"
)

type style struct {
	defaultWidth  int
	padding       int
	spacing       int
	indent        int
	titleHeight   int
	scrollbarSize int
	thumbSize     int
	colors        [ColorMax + 1]color.RGBA
}

const (
	ColorText = iota
	ColorBorder
	ColorWindowBG
	ColorTitleBG
	ColorTitleText
	ColorPanelBG
	ColorButton
	ColorButtonHover
	ColorButtonFocus
	ColorBase
	ColorBaseHover
	ColorBaseFocus
	ColorScrollBase
	ColorScrollThumb
	ColorMax = ColorScrollThumb
)

var defaultStyle style = style{
	defaultWidth:  60,
	padding:       5,
	spacing:       4,
	indent:        24,
	titleHeight:   24,
	scrollbarSize: 12,
	thumbSize:     8,
	colors: [...]color.RGBA{
		{230, 230, 230, 255}, // MU_COLOR_TEXT
		{25, 25, 25, 255},    // MU_COLOR_BORDER
		{50, 50, 50, 255},    // MU_COLOR_WINDOWBG
		{25, 25, 25, 255},    // MU_COLOR_TITLEBG
		{240, 240, 240, 255}, // MU_COLOR_TITLETEXT
		{0, 0, 0, 0},         // MU_COLOR_PANELBG
		{75, 75, 75, 255},    // MU_COLOR_BUTTON
		{95, 95, 95, 255},    // MU_COLOR_BUTTONHOVER
		{115, 115, 115, 255}, // MU_COLOR_BUTTONFOCUS
		{30, 30, 30, 255},    // MU_COLOR_BASE
		{35, 35, 35, 255},    // MU_COLOR_BASEHOVER
		{40, 40, 40, 255},    // MU_COLOR_BASEFOCUS
		{43, 43, 43, 255},    // MU_COLOR_SCROLLBASE
		{30, 30, 30, 255},    // MU_COLOR_SCROLLTHUMB
	},
}

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
		ColorText:        {230, 230, 230, 255},
		ColorBorder:      {25, 25, 25, 255},
		ColorWindowBG:    {50, 50, 50, 255},
		ColorTitleBG:     {25, 25, 25, 255},
		ColorTitleText:   {240, 240, 240, 255},
		ColorPanelBG:     {0, 0, 0, 0},
		ColorButton:      {75, 75, 75, 255},
		ColorButtonHover: {95, 95, 95, 255},
		ColorButtonFocus: {115, 115, 115, 255},
		ColorBase:        {30, 30, 30, 255},
		ColorBaseHover:   {35, 35, 35, 255},
		ColorBaseFocus:   {40, 40, 40, 255},
		ColorScrollBase:  {43, 43, 43, 255},
		ColorScrollThumb: {30, 30, 30, 255},
	},
}

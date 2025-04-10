// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image/color"
)

type style struct {
	defaultWidth  int
	defaultHeight int
	padding       int
	spacing       int
	indent        int
	titleHeight   int
	scrollbarSize int
	thumbSize     int
	colors        [colorCount]color.RGBA
}

const (
	colorText = iota
	colorBorder
	colorWindowBG
	colorTitleBG
	colorTitleBGTransparent
	colorTitleText
	colorPanelBG
	colorButton
	colorButtonHover
	colorButtonFocus
	colorBase
	colorBaseHover
	colorBaseFocus
	colorScrollBase
	colorScrollThumb
	colorCount
)

var defaultStyle style = style{
	defaultWidth:  60,
	defaultHeight: 18,
	padding:       5,
	spacing:       4,
	indent:        lineHeight(),
	titleHeight:   24,
	scrollbarSize: 12,
	thumbSize:     8,
	colors: [...]color.RGBA{
		colorText:               {230, 230, 230, 255},
		colorBorder:             {60, 60, 60, 255},
		colorWindowBG:           {45, 45, 45, 230},
		colorTitleBG:            {30, 30, 30, 255},
		colorTitleBGTransparent: {20, 20, 20, 204},
		colorTitleText:          {240, 240, 240, 255},
		colorPanelBG:            {0, 0, 0, 0},
		colorButton:             {75, 75, 75, 255},
		colorButtonHover:        {95, 95, 95, 255},
		colorButtonFocus:        {115, 115, 115, 255},
		colorBase:               {30, 30, 30, 255},
		colorBaseHover:          {35, 35, 35, 255},
		colorBaseFocus:          {40, 40, 40, 255},
		colorScrollBase:         {43, 43, 43, 255},
		colorScrollThumb:        {30, 30, 30, 255},
	},
}

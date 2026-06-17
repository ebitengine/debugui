// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image/color"
)

// Style holds layout metrics and colors for the debug UI.
// Use [DefaultStyle] as a base, then override fields, or load from JSON with [LoadStyleFile].
type Style struct {
	DefaultWidth  int `json:"defaultWidth"`
	DefaultHeight int `json:"defaultHeight"`
	Padding       int `json:"padding"`
	Spacing       int `json:"spacing"`
	Indent        int `json:"indent"`
	TitleHeight   int `json:"titleHeight"`
	ScrollbarSize int `json:"scrollbarSize"`
	ThumbSize     int `json:"thumbSize"`
	Colors        StyleColors `json:"colors"`
}

// StyleColors holds widget and chrome colors. Hover/focus variants for buttons,
// slider thumbs, and slider tracks are separate entries so themes can tune them independently.
type StyleColors struct {
	Text               color.RGBA `json:"text"`
	Border             color.RGBA `json:"border"`
	WindowBG           color.RGBA `json:"windowBG"`
	TitleBG            color.RGBA `json:"titleBG"`
	TitleBGTransparent color.RGBA `json:"titleBGTransparent"`
	TitleText          color.RGBA `json:"titleText"`
	PanelBG            color.RGBA `json:"panelBG"`
	Button             color.RGBA `json:"button"`
	ButtonHover        color.RGBA `json:"buttonHover"`
	ButtonFocus        color.RGBA `json:"buttonFocus"`
	SliderThumb        color.RGBA `json:"sliderThumb"`
	SliderThumbHover   color.RGBA `json:"sliderThumbHover"`
	SliderThumbFocus   color.RGBA `json:"sliderThumbFocus"`
	Base               color.RGBA `json:"base"`
	BaseHover          color.RGBA `json:"baseHover"`
	BaseFocus          color.RGBA `json:"baseFocus"`
	ScrollBase         color.RGBA `json:"scrollBase"`
	ScrollThumb        color.RGBA `json:"scrollThumb"`
	Selection          color.RGBA `json:"selection"`
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
	colorSliderThumb
	colorSliderThumbHover
	colorSliderThumbFocus
	colorBase
	colorBaseHover
	colorBaseFocus
	colorScrollBase
	colorScrollThumb
	colorSelection
	colorCount
)

var builtinStyle Style

func init() {
	builtinStyle = DefaultStyle()
}

// DefaultStyle returns the built-in dark theme. Indent is derived from the UI font line height.
func DefaultStyle() Style {
	return Style{
		DefaultWidth:  60,
		DefaultHeight: 18,
		Padding:       5,
		Spacing:       4,
		Indent:        lineHeight(),
		TitleHeight:   24,
		ScrollbarSize: 12,
		ThumbSize:     8,
		Colors:        defaultStyleColors(),
	}
}

func defaultStyleColors() StyleColors {
	return StyleColors{
		Text:               color.RGBA{R: 230, G: 230, B: 230, A: 255},
		Border:             color.RGBA{R: 60, G: 60, B: 60, A: 255},
		WindowBG:           color.RGBA{R: 45, G: 45, B: 45, A: 230},
		TitleBG:            color.RGBA{R: 30, G: 30, B: 30, A: 255},
		TitleBGTransparent: color.RGBA{R: 20, G: 20, B: 20, A: 204},
		TitleText:          color.RGBA{R: 240, G: 240, B: 240, A: 255},
		PanelBG:            color.RGBA{A: 0},
		Button:             color.RGBA{R: 75, G: 75, B: 75, A: 255},
		ButtonHover:        color.RGBA{R: 95, G: 95, B: 95, A: 255},
		ButtonFocus:        color.RGBA{R: 115, G: 115, B: 115, A: 255},
		SliderThumb:        color.RGBA{R: 75, G: 75, B: 75, A: 255},
		SliderThumbHover:   color.RGBA{R: 95, G: 95, B: 95, A: 255},
		SliderThumbFocus:   color.RGBA{R: 115, G: 115, B: 115, A: 255},
		Base:               color.RGBA{R: 30, G: 30, B: 30, A: 255},
		BaseHover:          color.RGBA{R: 35, G: 35, B: 35, A: 255},
		BaseFocus:          color.RGBA{R: 40, G: 40, B: 40, A: 255},
		ScrollBase:         color.RGBA{R: 43, G: 43, B: 43, A: 255},
		ScrollThumb:        color.RGBA{R: 30, G: 30, B: 30, A: 255},
		Selection:          color.RGBA{R: 70, G: 100, B: 140, A: 255},
	}
}

func (s *Style) widgetColor(id int) color.RGBA {
	switch id {
	case colorText:
		return s.Colors.Text
	case colorBorder:
		return s.Colors.Border
	case colorWindowBG:
		return s.Colors.WindowBG
	case colorTitleBG:
		return s.Colors.TitleBG
	case colorTitleBGTransparent:
		return s.Colors.TitleBGTransparent
	case colorTitleText:
		return s.Colors.TitleText
	case colorPanelBG:
		return s.Colors.PanelBG
	case colorButton:
		return s.Colors.Button
	case colorButtonHover:
		return s.Colors.ButtonHover
	case colorButtonFocus:
		return s.Colors.ButtonFocus
	case colorSliderThumb:
		return s.Colors.SliderThumb
	case colorSliderThumbHover:
		return s.Colors.SliderThumbHover
	case colorSliderThumbFocus:
		return s.Colors.SliderThumbFocus
	case colorBase:
		return s.Colors.Base
	case colorBaseHover:
		return s.Colors.BaseHover
	case colorBaseFocus:
		return s.Colors.BaseFocus
	case colorScrollBase:
		return s.Colors.ScrollBase
	case colorScrollThumb:
		return s.Colors.ScrollThumb
	case colorSelection:
		return s.Colors.Selection
	default:
		return color.RGBA{}
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"fmt"
	"image/color"
	"strings"
)

// ThemeLight is a light palette intended for UIs drawn over pale or white backgrounds.
func ThemeLight() Style {
	s := DefaultStyle()
	s.Colors = StyleColors{
		Text:               color.RGBA{R: 30, G: 30, B: 30, A: 255},
		Border:             color.RGBA{R: 140, G: 140, B: 150, A: 255},
		WindowBG:           color.RGBA{R: 245, G: 245, B: 248, A: 240},
		TitleBG:            color.RGBA{R: 228, G: 228, B: 235, A: 255},
		TitleBGTransparent: color.RGBA{R: 220, G: 220, B: 228, A: 220},
		TitleText:          color.RGBA{R: 20, G: 20, B: 24, A: 255},
		PanelBG:            color.RGBA{A: 0},
		Button:             color.RGBA{R: 210, G: 210, B: 218, A: 255},
		ButtonHover:        color.RGBA{R: 195, G: 195, B: 208, A: 255},
		ButtonFocus:        color.RGBA{R: 175, G: 180, B: 200, A: 255},
		SliderThumb:        color.RGBA{R: 200, G: 200, B: 212, A: 255},
		SliderThumbHover:   color.RGBA{R: 185, G: 185, B: 200, A: 255},
		SliderThumbFocus:   color.RGBA{R: 165, G: 170, B: 195, A: 255},
		Base:               color.RGBA{R: 236, G: 236, B: 242, A: 255},
		BaseHover:          color.RGBA{R: 228, G: 228, B: 236, A: 255},
		BaseFocus:          color.RGBA{R: 218, G: 220, B: 232, A: 255},
		ScrollBase:         color.RGBA{R: 220, G: 220, B: 228, A: 255},
		ScrollThumb:        color.RGBA{R: 180, G: 180, B: 195, A: 255},
		Selection:          color.RGBA{R: 100, G: 150, B: 210, A: 180},
	}
	return s
}

// ThemeOption pairs a [BuiltInTheme] key with a short label for menus and settings UIs.
type ThemeOption struct {
	Key   string
	Label string
}

// BuiltInThemeMenu returns the predefined themes in display order.
// Each [ThemeOption.Key] is valid for [BuiltInTheme].
func BuiltInThemeMenu() []ThemeOption {
	return []ThemeOption{
		{Key: "dark", Label: "Dark"},
		{Key: "light", Label: "Light"},
	}
}

// BuiltInTheme returns a copy of a named packaged theme.
// Recognized names: "dark", "default" (same as [DefaultStyle]), "light" ([ThemeLight]).
func BuiltInTheme(name string) (Style, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "dark", "default":
		return DefaultStyle(), nil
	case "light":
		return ThemeLight(), nil
	default:
		return Style{}, fmt.Errorf("debugui: unknown built-in theme %q", name)
	}
}

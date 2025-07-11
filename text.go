// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"iter"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

func removeSpaceAtLineTail(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

func lines(text string, width int) iter.Seq[string] {
	return func(yield func(string) bool) {
		var line string
		var word string
		state := -1
		for len(text) > 0 {
			cluster, nextText, boundaries, nextState := uniseg.StepString(text, state)
			switch m := boundaries & uniseg.MaskLine; m {
			default:
				word += cluster
			case uniseg.LineCanBreak, uniseg.LineMustBreak:
				if line == "" {
					line += word + cluster
				} else {
					if l := removeSpaceAtLineTail(line + word + cluster); textWidth(l) > width {
						if !yield(removeSpaceAtLineTail(line)) {
							return
						}
						line = word + cluster
					} else {
						line += word + cluster
					}
				}
				word = ""
				if m == uniseg.LineMustBreak {
					if !yield(removeSpaceAtLineTail(line)) {
						return
					}
					line = ""
				}
			}
			state = nextState
			text = nextText
		}

		line += word
		if len(line) > 0 {
			if !yield(removeSpaceAtLineTail(line)) {
				return
			}
		}
	}
}

// Text creates a text label.
func (c *Context) Text(text string) {
	c.GridCell(func(bounds image.Rectangle) {
		for line := range lines(text, bounds.Dx()-c.style().padding) {
			_, _ = c.widget(widgetID{}, 0, nil, nil, func(bounds image.Rectangle) {
				c.drawWidgetText(line, bounds, colorText, 0)
			})
		}
	})
}

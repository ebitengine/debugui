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
					l := removeSpaceAtLineTail(line + word + cluster)
					if textWidth(l) > width {
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
	c.wrapError(func() error {
		if err := c.gridCell(func(bounds image.Rectangle) error {
			c.SetGridLayout([]int{-1}, []int{lineHeight()})
			for line := range lines(text, bounds.Dx()-c.style().padding) {
				if _, err := c.widget(emptyWidgetID, 0, func(bounds image.Rectangle, wasFocused bool) (bool, error) {
					c.drawWidgetText(line, bounds, colorText, 0)
					return false, nil
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
}

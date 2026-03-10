// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

import (
	"image"
	"iter"
	"strings"
	"unicode"
)

func removeSpaceAtLineTail(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

func sanitizeUTF8(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(r)
	}
	return b.String()
}

func (c *Context) lines(text string, width int) iter.Seq[string] {
	return func(yield func(string) bool) {
		seg := c.pushSegmenter()
		defer c.popSegmenter()

		if err := seg.InitWithString(text); err != nil {
			text = sanitizeUTF8(text)
			if err := seg.InitWithString(text); err != nil {
				panic("debugui: segmenter.InitWithString failed even after sanitizing: " + err.Error())
			}
		}

		var line string
		it := seg.LineIterator()
		for it.Next() {
			l := it.Line()
			segment := text[l.OffsetInBytes : l.OffsetInBytes+l.LengthInBytes]

			if line == "" {
				line = segment
			} else {
				if trimmed := removeSpaceAtLineTail(line + segment); textWidth(trimmed) > width {
					if !yield(removeSpaceAtLineTail(line)) {
						return
					}
					line = segment
				} else {
					line += segment
				}
			}

			if l.IsMandatoryBreak {
				if !yield(removeSpaceAtLineTail(line)) {
					return
				}
				line = ""
			}
		}

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
		for line := range c.lines(text, bounds.Dx()-c.style().padding) {
			_, _ = c.widget(widgetID{}, 0, nil, nil, func(bounds image.Rectangle) {
				c.drawWidgetText(line, bounds, colorText, 0)
			})
		}
	})
}

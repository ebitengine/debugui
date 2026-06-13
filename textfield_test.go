// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui_test

import (
	"testing"

	"github.com/ebitengine/debugui"
)

func TestTextFieldStateCaretMovement(t *testing.T) {
	// "aあb" consists of runes of 1, 3, and 1 bytes.
	f := debugui.NewTextFieldState("aあb")

	if start, end := f.Selection(); start != 5 || end != 5 {
		t.Errorf("Selection() = %d, %d; want 5, 5", start, end)
	}

	f.MoveCaretLeft(false)
	if start, end := f.Selection(); start != 4 || end != 4 {
		t.Errorf("Selection() = %d, %d; want 4, 4", start, end)
	}
	f.MoveCaretLeft(false)
	if start, end := f.Selection(); start != 1 || end != 1 {
		t.Errorf("Selection() = %d, %d; want 1, 1", start, end)
	}
	f.MoveCaretLeft(false)
	f.MoveCaretLeft(false)
	if start, end := f.Selection(); start != 0 || end != 0 {
		t.Errorf("Selection() = %d, %d; want 0, 0", start, end)
	}

	f.MoveCaretRight(false)
	if start, end := f.Selection(); start != 1 || end != 1 {
		t.Errorf("Selection() = %d, %d; want 1, 1", start, end)
	}
	f.MoveCaretRight(false)
	if start, end := f.Selection(); start != 4 || end != 4 {
		t.Errorf("Selection() = %d, %d; want 4, 4", start, end)
	}

	f.MoveCaretTo(100, false)
	if start, end := f.Selection(); start != 5 || end != 5 {
		t.Errorf("Selection() = %d, %d; want 5, 5", start, end)
	}
	f.MoveCaretTo(-1, false)
	if start, end := f.Selection(); start != 0 || end != 0 {
		t.Errorf("Selection() = %d, %d; want 0, 0", start, end)
	}
}

func TestTextFieldStateSelection(t *testing.T) {
	f := debugui.NewTextFieldState("hello")

	f.MoveCaretTo(1, false)
	f.MoveCaretTo(4, true)
	if start, end := f.Selection(); start != 1 || end != 4 {
		t.Errorf("Selection() = %d, %d; want 1, 4", start, end)
	}

	// Moving left without extending collapses the selection to its start.
	f.MoveCaretLeft(false)
	if start, end := f.Selection(); start != 1 || end != 1 {
		t.Errorf("Selection() = %d, %d; want 1, 1", start, end)
	}

	// A selection extended backward has its caret before its anchor.
	f.MoveCaretTo(4, false)
	f.MoveCaretTo(1, true)
	if start, end := f.Selection(); start != 1 || end != 4 {
		t.Errorf("Selection() = %d, %d; want 1, 4", start, end)
	}

	// Moving right without extending collapses the selection to its end.
	f.MoveCaretRight(false)
	if start, end := f.Selection(); start != 4 || end != 4 {
		t.Errorf("Selection() = %d, %d; want 4, 4", start, end)
	}

	// Extending shrinks the selection when the caret moves back toward the anchor.
	f.MoveCaretTo(0, false)
	f.MoveCaretTo(3, true)
	f.MoveCaretLeft(true)
	if start, end := f.Selection(); start != 0 || end != 2 {
		t.Errorf("Selection() = %d, %d; want 0, 2", start, end)
	}

	f.SelectAll()
	if start, end := f.Selection(); start != 0 || end != 5 {
		t.Errorf("Selection() = %d, %d; want 0, 5", start, end)
	}
}

func TestTextFieldStateDeleteSelection(t *testing.T) {
	f := debugui.NewTextFieldState("hello world")
	f.MoveCaretTo(5, false)
	f.MoveCaretTo(11, true)
	f.DeleteBackward()
	if got, want := f.Text(), "hello"; got != want {
		t.Errorf("Text() = %q; want %q", got, want)
	}
	if start, end := f.Selection(); start != 5 || end != 5 {
		t.Errorf("Selection() = %d, %d; want 5, 5", start, end)
	}

	f = debugui.NewTextFieldState("hello world")
	f.MoveCaretTo(6, false)
	f.MoveCaretTo(0, true)
	f.DeleteForward()
	if got, want := f.Text(), "world"; got != want {
		t.Errorf("Text() = %q; want %q", got, want)
	}
	if start, end := f.Selection(); start != 0 || end != 0 {
		t.Errorf("Selection() = %d, %d; want 0, 0", start, end)
	}

	f = debugui.NewTextFieldState("hello")
	f.SelectAll()
	f.DeleteBackward()
	if got, want := f.Text(), ""; got != want {
		t.Errorf("Text() = %q; want %q", got, want)
	}
}

func TestTextFieldStateDeleteRune(t *testing.T) {
	// "aあb" consists of runes of 1, 3, and 1 bytes.
	f := debugui.NewTextFieldState("aあb")
	f.MoveCaretTo(4, false)
	f.DeleteBackward()
	if got, want := f.Text(), "ab"; got != want {
		t.Errorf("Text() = %q; want %q", got, want)
	}
	if start, end := f.Selection(); start != 1 || end != 1 {
		t.Errorf("Selection() = %d, %d; want 1, 1", start, end)
	}

	f.DeleteForward()
	if got, want := f.Text(), "a"; got != want {
		t.Errorf("Text() = %q; want %q", got, want)
	}
	if start, end := f.Selection(); start != 1 || end != 1 {
		t.Errorf("Selection() = %d, %d; want 1, 1", start, end)
	}

	// Deleting at the boundaries is a no-op.
	f.DeleteForward()
	f.MoveCaretTo(0, false)
	f.DeleteBackward()
	if got, want := f.Text(), "a"; got != want {
		t.Errorf("Text() = %q; want %q", got, want)
	}
}

func TestTextFieldStateWordMovement(t *testing.T) {
	// Byte offsets:        0         1         2
	//                      0123456789012345678901234
	f := debugui.NewTextFieldState("foo, bar_baz  qux")

	// From the end, word-left lands at the start of each word, skipping separators.
	f.MoveCaretWordLeft(false)
	if start, end := f.Selection(); start != 14 || end != 14 {
		t.Errorf("Selection() = %d, %d; want 14, 14 (start of \"qux\")", start, end)
	}
	// "bar_baz" is a single word because '_' is a word rune.
	f.MoveCaretWordLeft(false)
	if start, end := f.Selection(); start != 5 || end != 5 {
		t.Errorf("Selection() = %d, %d; want 5, 5 (start of \"bar_baz\")", start, end)
	}
	f.MoveCaretWordLeft(false)
	if start, end := f.Selection(); start != 0 || end != 0 {
		t.Errorf("Selection() = %d, %d; want 0, 0 (start of \"foo\")", start, end)
	}
	// Word-left at the start is a no-op.
	f.MoveCaretWordLeft(false)
	if start, end := f.Selection(); start != 0 || end != 0 {
		t.Errorf("Selection() = %d, %d; want 0, 0", start, end)
	}

	// From the start, word-right lands at the end of each word.
	f.MoveCaretWordRight(false)
	if start, end := f.Selection(); start != 3 || end != 3 {
		t.Errorf("Selection() = %d, %d; want 3, 3 (end of \"foo\")", start, end)
	}
	f.MoveCaretWordRight(false)
	if start, end := f.Selection(); start != 12 || end != 12 {
		t.Errorf("Selection() = %d, %d; want 12, 12 (end of \"bar_baz\")", start, end)
	}
	f.MoveCaretWordRight(false)
	if start, end := f.Selection(); start != 17 || end != 17 {
		t.Errorf("Selection() = %d, %d; want 17, 17 (end of \"qux\")", start, end)
	}

	// Word movement extends the selection when requested.
	f.MoveCaretTo(0, false)
	f.MoveCaretWordRight(true)
	if start, end := f.Selection(); start != 0 || end != 3 {
		t.Errorf("Selection() = %d, %d; want 0, 3", start, end)
	}
}

func TestWordRangeAt(t *testing.T) {
	// Byte offsets:  0         1
	//                0123456789012345
	const s = "foo, bar_baz  qux"

	tests := []struct {
		pos        int
		start, end int
	}{
		// Inside a word selects the whole word.
		{pos: 0, start: 0, end: 3},    // start of "foo"
		{pos: 1, start: 0, end: 3},    // middle of "foo"
		{pos: 3, start: 0, end: 3},    // end of "foo", adjacent to ','
		{pos: 14, start: 14, end: 17}, // start of "qux"
		{pos: 17, start: 14, end: 17}, // end of the string, after "qux"
		// '_' is a word rune, so "bar_baz" is one word.
		{pos: 8, start: 5, end: 12},
		// A click on a run of non-word runes selects that run; punctuation and spaces
		// are both non-word, so the comma and the following space select together.
		{pos: 4, start: 3, end: 5},
		{pos: 13, start: 12, end: 14}, // the two spaces before "qux"
	}
	for _, tc := range tests {
		if start, end := debugui.WordRangeAt(s, tc.pos); start != tc.start || end != tc.end {
			t.Errorf("WordRangeAt(%q, %d) = %d, %d; want %d, %d", s, tc.pos, start, end, tc.start, tc.end)
		}
	}

	// A position past either end clamps into range.
	if start, end := debugui.WordRangeAt(s, -1); start != 0 || end != 3 {
		t.Errorf("WordRangeAt(%q, -1) = %d, %d; want 0, 3", s, start, end)
	}
	if start, end := debugui.WordRangeAt(s, 100); start != 14 || end != 17 {
		t.Errorf("WordRangeAt(%q, 100) = %d, %d; want 14, 17", s, start, end)
	}

	// An empty string has an empty word range.
	if start, end := debugui.WordRangeAt("", 0); start != 0 || end != 0 {
		t.Errorf("WordRangeAt(%q, 0) = %d, %d; want 0, 0", "", start, end)
	}
}

func TestTextFieldStateMultiClick(t *testing.T) {
	const interval = 30

	f := debugui.NewTextFieldState("hello world")

	// A single click places the caret without selecting.
	f.HandleClick(2, false, 100, interval)
	if start, end := f.Selection(); start != 2 || end != 2 {
		t.Errorf("after single click: Selection() = %d, %d; want 2, 2", start, end)
	}

	// A second click within the interval is a double-click and selects the word.
	f.HandleClick(2, false, 110, interval)
	if start, end := f.Selection(); start != 0 || end != 5 {
		t.Errorf("after double-click: Selection() = %d, %d; want 0, 5 (\"hello\")", start, end)
	}

	// A third click within the interval is a triple-click and selects the whole text.
	f.HandleClick(2, false, 120, interval)
	if start, end := f.Selection(); start != 0 || end != 11 {
		t.Errorf("after triple-click: Selection() = %d, %d; want 0, 11", start, end)
	}

	// A fourth click within the interval keeps the triple-click selection.
	f.HandleClick(2, false, 130, interval)
	if start, end := f.Selection(); start != 0 || end != 11 {
		t.Errorf("after fourth click: Selection() = %d, %d; want 0, 11", start, end)
	}

	// A click after the interval elapses restarts the sequence as a single click.
	f.HandleClick(8, false, 200, interval)
	if start, end := f.Selection(); start != 8 || end != 8 {
		t.Errorf("after interval elapses: Selection() = %d, %d; want 8, 8", start, end)
	}
	// And the next quick click is a double-click again, selecting the word under it.
	f.HandleClick(8, false, 210, interval)
	if start, end := f.Selection(); start != 6 || end != 11 {
		t.Errorf("after double-click: Selection() = %d, %d; want 6, 11 (\"world\")", start, end)
	}
}

func TestTextIndexFromX(t *testing.T) {
	if got, want := debugui.TextIndexFromX("abc", -1), 0; got != want {
		t.Errorf("TextIndexFromX(%q, -1) = %d; want %d", "abc", got, want)
	}
	if got, want := debugui.TextIndexFromX("abc", debugui.TextWidth("abc")+10), 3; got != want {
		t.Errorf("TextIndexFromX(%q, %d) = %d; want %d", "abc", debugui.TextWidth("abc")+10, got, want)
	}
	// A position right on an inter-rune boundary resolves to that boundary.
	if got, want := debugui.TextIndexFromX("abc", debugui.TextWidth("ab")), 2; got != want {
		t.Errorf("TextIndexFromX(%q, %d) = %d; want %d", "abc", debugui.TextWidth("ab"), got, want)
	}
	// A position past the midpoint of a rune resolves to the next boundary.
	if got, want := debugui.TextIndexFromX("aあb", debugui.TextWidth("a")+debugui.TextWidth("あ")*3/4), 4; got != want {
		t.Errorf("TextIndexFromX(%q, ...) = %d; want %d", "aあb", got, want)
	}
}

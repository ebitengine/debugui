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

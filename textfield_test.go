// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package debugui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

func TestFocusTextInputFieldInitializesTextAndCaretAtEnd(t *testing.T) {
	var f textinput.Field
	t.Cleanup(f.Blur)

	// This is the original regression: focusing a field with preloaded text must
	// also initialize the internal selection so the next typed character appends.
	focusTextInputField(&f, "hello")

	if got, want := f.Text(), "hello"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
	if !f.IsFocused() {
		t.Fatal("field is not focused")
	}

	start, end := f.Selection()
	if got, want := start, len("hello"); got != want {
		t.Fatalf("selection start = %d, want %d", got, want)
	}
	if got, want := end, len("hello"); got != want {
		t.Fatalf("selection end = %d, want %d", got, want)
	}
}

func TestSetTextFieldValueMovesCaretToEnd(t *testing.T) {
	var c Context
	cnt := &container{}
	c.containerStack = []*container{cnt}

	id := widgetID{}.push("field")
	c.currentID = id
	cnt.textInputTextField(id, true)

	// SetTextFieldValue is used to replace the visible contents, so it must leave
	// the hidden caret state at the end of the new text as well.
	c.SetTextFieldValue("loaded")

	f := cnt.textInputTextField(id, false)
	if f == nil {
		t.Fatal("field was not created")
	}
	if got, want := f.Text(), "loaded"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}

	start, end := f.Selection()
	if got, want := start, len("loaded"); got != want {
		t.Fatalf("selection start = %d, want %d", got, want)
	}
	if got, want := end, len("loaded"); got != want {
		t.Fatalf("selection end = %d, want %d", got, want)
	}
}

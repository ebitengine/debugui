// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui_test

import (
	"errors"
	"image"
	"testing"

	"github.com/ebitengine/debugui"
)

func TestMultipleButtonsInForLoop(t *testing.T) {
	var d debugui.DebugUI
	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			var id debugui.ControlID
			for range 10 {
				id2 := ctx.ButtonID("a")
				if id2 == 0 {
					t.Errorf("Caller() returned 0")
					continue
				}
				if id == 0 {
					id = id2
					continue
				}
				if id != id2 {
					t.Errorf("Caller() returned different values: %d and %d", id, id2)
				}
			}
			if id == 0 {
				t.Errorf("Caller() returned 0")
			}
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMultipleButtonsOnOneLine(t *testing.T) {
	var d debugui.DebugUI
	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			idA1 := ctx.ButtonID("a")
			idA2 := ctx.ButtonID("a")
			if idA1 == idA2 {
				t.Errorf("Button() returned the same value twice: %d", idA1)
			}
			idB1, idB2 := ctx.ButtonID("b"), ctx.ButtonID("b")
			if idB1 == idB2 {
				t.Errorf("Button() returned the same value twice: %d", idB1)
			}
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestError(t *testing.T) {
	e := errors.New("test")
	var d debugui.DebugUI
	if got, want := d.Update(func(ctx *debugui.Context) error {
		return e
	}), e; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

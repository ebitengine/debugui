// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui_test

import (
	"errors"
	"image"
	"testing"

	"github.com/ebitengine/debugui"
)

func TestMultipleIDPartFromCallersInForLoop(t *testing.T) {
	var d debugui.DebugUI
	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			var idPart string
			for range 10 {
				idPart2 := debugui.IDPartFromCaller()
				if idPart2 == "" {
					t.Errorf("IDPartFromCaller() returned an empty string")
					continue
				}
				if idPart == "" {
					idPart = idPart2
					continue
				}
				if idPart != idPart2 {
					t.Errorf("IDPartFromCaller() returned different values: %q and %q", idPart, idPart2)
				}
			}
			if idPart == "" {
				t.Errorf("IDPartFromCaller() returned an empty string")
			}
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMultipleIDPartFromCallersOnOneLine(t *testing.T) {
	var d debugui.DebugUI
	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			idPartA1 := debugui.IDPartFromCaller()
			idPartA2 := debugui.IDPartFromCaller()
			if idPartA1 == idPartA2 {
				t.Errorf("IDPartFromCaller() returned the same value twice: %q", idPartA1)
			}
			idPartB1, idPartB2 := debugui.IDPartFromCaller(), debugui.IDPartFromCaller()
			if idPartB1 == idPartB2 {
				t.Errorf("IDPartFromCaller() returned the same value twice: %q", idPartB1)
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
	_, err := d.Update(func(ctx *debugui.Context) error {
		return e
	})
	if got, want := err, e; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestUpdateWithoutWindow(t *testing.T) {
	var d debugui.DebugUI
	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.SetGridLayout(nil, nil)
		return nil
	}); err == nil {
		t.Errorf("Update() returned nil, want error")
	}
}

func TestUnusedContainer(t *testing.T) {
	var d debugui.DebugUI
	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window1", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 1; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window1", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		ctx.Window("Window2", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 2; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window1", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 1; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if _, err := d.Update(func(ctx *debugui.Context) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 0; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if _, err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window2", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 1; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

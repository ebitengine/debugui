// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui_test

import (
	"errors"
	"image"
	"testing"

	"github.com/ebitengine/debugui"
)

func TestMultipleIDFromCallersInForLoop(t *testing.T) {
	var d debugui.DebugUI
	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			var id debugui.WidgetID
			for range 10 {
				id2 := ctx.IDFromCaller()
				if id2 == debugui.EmptyWidgetID {
					t.Errorf("IDFromCaller() returned 0")
					continue
				}
				if id == debugui.EmptyWidgetID {
					id = id2
					continue
				}
				if id != id2 {
					t.Errorf("IDFromCaller) returned different values: %q and %q", id, id2)
				}
			}
			if id == debugui.EmptyWidgetID {
				t.Errorf("IDFromCaller() returned 0")
			}
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMultipleIDFromCallersOnOneLine(t *testing.T) {
	var d debugui.DebugUI
	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			idA1 := ctx.IDFromCaller()
			idA2 := ctx.IDFromCaller()
			if idA1 == idA2 {
				t.Errorf("IDFromCaller() returned the same value twice: %q", idA1)
			}
			idB1, idB2 := ctx.IDFromCaller(), ctx.IDFromCaller()
			if idB1 == idB2 {
				t.Errorf("IDFromCaller() returned the same value twice: %q", idB1)
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

func TestUpdateWithoutWindow(t *testing.T) {
	var d debugui.DebugUI
	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.SetGridLayout(nil, nil)
		return nil
	}); err == nil {
		t.Errorf("Update() returned nil, want error")
	}
}

func TestUnusedContainer(t *testing.T) {
	var d debugui.DebugUI
	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window1", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 1; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if err := d.Update(func(ctx *debugui.Context) error {
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

	if err := d.Update(func(ctx *debugui.Context) error {
		ctx.Window("Window1", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 1; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if err := d.Update(func(ctx *debugui.Context) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := d.ContainerCounter(), 0; got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}

	if err := d.Update(func(ctx *debugui.Context) error {
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

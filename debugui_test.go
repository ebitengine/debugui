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

func TestCollapsedWindowDoesNotCoverWindowsBelow(t *testing.T) {
	const (
		bottom = 0
		top    = 1
	)
	var d debugui.DebugUI
	update := func() {
		t.Helper()
		if _, err := d.Update(func(ctx *debugui.Context) error {
			ctx.Window("Bottom", image.Rect(50, 70, 260, 180), func(layout debugui.ContainerLayout) {})
			ctx.Window("Top", image.Rect(0, 0, 320, 240), func(layout debugui.ContainerLayout) {})
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	update()

	// The title bar of Top is 24 pixels tall by default.
	onTopTitleBar := image.Pt(10, 10)
	onBottomBelowTopTitleBar := image.Pt(100, 100)
	onTopBodyOnly := image.Pt(10, 200)
	outside := image.Pt(400, 400)

	testCases := []struct {
		name  string
		pt    image.Point
		want  int
		setup func()
	}{
		{name: "expanded title bar", pt: onTopTitleBar, want: top},
		{name: "expanded body over bottom", pt: onBottomBelowTopTitleBar, want: top},
		{name: "expanded body only", pt: onTopBodyOnly, want: top},
		{name: "outside", pt: outside, want: -1},
		{name: "collapsed title bar", pt: onTopTitleBar, want: top, setup: func() { d.SetRootContainerCollapsed(top, true) }},
		{name: "collapsed title bar bottom edge", pt: image.Pt(10, 23), want: top},
		{name: "collapsed just below title bar", pt: image.Pt(10, 24), want: -1},
		{name: "collapsed body over bottom", pt: onBottomBelowTopTitleBar, want: bottom},
		{name: "collapsed body only", pt: onTopBodyOnly, want: -1},
		{name: "collapsed and moved, old title bar", pt: onTopTitleBar, want: -1, setup: func() { d.MoveRootContainer(top, image.Pt(0, 150)) }},
		{name: "collapsed and moved, new title bar", pt: image.Pt(10, 160), want: top},
		{name: "collapsed and moved, over bottom", pt: onBottomBelowTopTitleBar, want: bottom},
		{name: "expanded again", pt: image.Pt(10, 200), want: top, setup: func() { d.SetRootContainerCollapsed(top, false) }},
	}
	for _, tc := range testCases {
		if tc.setup != nil {
			tc.setup()
			update()
		}
		if got := d.RootContainerIndexAt(tc.pt); got != tc.want {
			t.Errorf("%s: RootContainerIndexAt(%v): got: %d, want: %d", tc.name, tc.pt, got, tc.want)
		}
	}
}

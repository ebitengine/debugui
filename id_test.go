// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui_test

import (
	"image"
	"testing"

	"github.com/ebitengine/debugui"
)

func TestMultipleButtonsInForLoop(t *testing.T) {
	var d debugui.DebugUI
	d.Update(func(ctx *debugui.Context) {
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
	})
}

func TestMultipleButtonsOnOneLine(t *testing.T) {
	var d debugui.DebugUI
	d.Update(func(ctx *debugui.Context) {
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
	})
}

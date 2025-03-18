// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui_test

import (
	"image"
	"testing"

	"github.com/ebitengine/debugui"
)

func TestMultipleButtonsInForLoop(t *testing.T) {
	d := debugui.New()
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
	d := debugui.New()
	d.Update(func(ctx *debugui.Context) {
		ctx.Window("Window", image.Rect(0, 0, 100, 100), func(layout debugui.ContainerLayout) {
			idA := ctx.ButtonID("a")
			idB := ctx.ButtonID("b")
			if idA == idB {
				t.Errorf("Button() returned the same value twice: %d", idA)
			}
			idC, idD := ctx.ButtonID("c"), ctx.ButtonID("d")
			if idC == idD {
				t.Errorf("Button() returned the same value twice: %d", idC)
			}
		})
	})
}

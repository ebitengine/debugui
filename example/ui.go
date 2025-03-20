// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ebitengine/debugui"
)

func (g *Game) writeLog(text string) {
	if len(g.logBuf) > 0 {
		g.logBuf += "\n"
	}
	g.logBuf += text
	g.logUpdated = true
}

func (g *Game) testWindow(ctx *debugui.Context) {
	ctx.Window("Demo Window", image.Rect(40, 40, 340, 500), func(layout debugui.ContainerLayout) {
		ctx.Header("Window Info", false, func() {
			ctx.SetGridLayout([]int{54, -1}, nil)
			ctx.Label("Position:")
			ctx.Label(fmt.Sprintf("%d, %d", layout.Bounds.Min.X, layout.Bounds.Min.Y))
			ctx.Label("Size:")
			ctx.Label(fmt.Sprintf("%d, %d", layout.Bounds.Dx(), layout.Bounds.Dy()))
		})
		ctx.Header("Test Buttons", true, func() {
			ctx.SetGridLayout([]int{100, -1, -1}, nil)
			ctx.Label("Test buttons 1:")
			if ctx.Button("Button 1") {
				g.writeLog("Pressed button 1")
			}
			if ctx.Button("Button 2") {
				g.writeLog("Pressed button 2")
			}
			ctx.Label("Test buttons 2:")
			if ctx.Button("Button 3") {
				g.writeLog("Pressed button 3")
			}
			if ctx.Button("Popup") {
				ctx.OpenPopup("Test Popup")
			}
			ctx.Popup("Test Popup", func(layout debugui.ContainerLayout) {
				ctx.Button("Hello")
				ctx.Button("World")
				if ctx.Button("Close") {
					ctx.ClosePopup("Test Popup")
				}
			})
		})
		ctx.Header("Tree and Text", true, func() {
			ctx.SetGridLayout([]int{140, -1}, nil)
			ctx.GridCell(func() {
				ctx.TreeNode("Test 1", func() {
					ctx.TreeNode("Test 1a", func() {
						ctx.Label("Hello")
						ctx.Label("World")
					})
					ctx.TreeNode("Test 1b", func() {
						if ctx.Button("Button 1") {
							g.writeLog("Pressed button 1")
						}
						if ctx.Button("Button 2") {
							g.writeLog("Pressed button 2")
						}
					})
				})
				ctx.TreeNode("Test 2", func() {
					ctx.SetGridLayout([]int{54, 54}, nil)
					if ctx.Button("Button 3") {
						g.writeLog("Pressed button 3")
					}
					if ctx.Button("Button 4") {
						g.writeLog("Pressed button 4")
					}
					if ctx.Button("Button 5") {
						g.writeLog("Pressed button 5")
					}
					if ctx.Button("Button 6") {
						g.writeLog("Pressed button 6")
					}
				})
				ctx.TreeNode("Test 3", func() {
					ctx.Checkbox("Checkbox 1", &g.checks[0])
					ctx.Checkbox("Checkbox 2", &g.checks[1])
					ctx.Checkbox("Checkbox 3", &g.checks[2])
				})
			})

			ctx.Text("Lorem ipsum dolor sit amet, consectetur adipiscing " +
				"elit. Maecenas lacinia, sem eu lacinia molestie, mi risus faucibus " +
				"ipsum, eu varius magna felis a nulla.")
		})
		ctx.Header("Background Color", true, func() {
			ctx.SetGridLayout([]int{-1, 78}, []int{74})
			ctx.GridCell(func() {
				ctx.SetGridLayout([]int{46, -1}, nil)
				ctx.Label("Red:")
				ctx.Slider(&g.bg[0], 0, 255, 1, 0)
				ctx.Label("Green:")
				ctx.Slider(&g.bg[1], 0, 255, 1, 0)
				ctx.Label("Blue:")
				ctx.Slider(&g.bg[2], 0, 255, 1, 0)
			})
			ctx.Control("", func(bounds image.Rectangle) bool {
				ctx.DrawControl(func(screen *ebiten.Image) {
					vector.DrawFilledRect(
						screen,
						float32(bounds.Min.X),
						float32(bounds.Min.Y),
						float32(bounds.Dx()),
						float32(bounds.Dy()),
						color.RGBA{byte(g.bg[0]), byte(g.bg[1]), byte(g.bg[2]), 255},
						false)
					txt := fmt.Sprintf("#%02X%02X%02X", int(g.bg[0]), int(g.bg[1]), int(g.bg[2]))
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64((bounds.Min.X+bounds.Max.X)/2), float64((bounds.Min.Y+bounds.Max.Y)/2))
					op.PrimaryAlign = text.AlignCenter
					op.SecondaryAlign = text.AlignCenter
					debugui.DrawText(screen, txt, op)
				})
				return false
			})
		})
		ctx.Header("Number", true, func() {
			ctx.SetGridLayout([]int{-1}, nil)
			ctx.Number(&g.num1, 0.1, 2)
			ctx.Slider(&g.num2, 0, 10, 0.1, 2)
		})
	})
}

func (g *Game) logWindow(ctx *debugui.Context) {
	ctx.Window("Log Window", image.Rect(350, 40, 650, 290), func(layout debugui.ContainerLayout) {
		ctx.SetGridLayout([]int{-1}, []int{-1, 0})
		ctx.Panel("Log Output", func(layout debugui.ContainerLayout) {
			ctx.SetGridLayout([]int{-1}, []int{-1})
			ctx.Text(g.logBuf)
			if g.logUpdated {
				ctx.SetScroll(image.Pt(layout.ScrollOffset.X, layout.ContentSize.Y))
				g.logUpdated = false
			}
		})
		ctx.GridCell(func() {
			var submitted bool
			ctx.SetGridLayout([]int{-1, 70}, nil)
			if ctx.TextField(&g.logSubmitBuf) {
				if ebiten.IsKeyPressed(ebiten.KeyEnter) {
					submitted = true
				}
			}
			if ctx.Button("Submit") {
				submitted = true
			}
			if submitted {
				g.writeLog(g.logSubmitBuf)
				g.logSubmitBuf = ""
				ctx.SetTextFieldValue(&g.logSubmitBuf)
			}
		})
	})
}

func (g *Game) buttonWindows(ctx *debugui.Context) {
	ctx.Window("Button Windows", image.Rect(350, 300, 650, 500), func(layout debugui.ContainerLayout) {
		ctx.SetGridLayout([]int{-1, -1, -1, -1}, nil)
		for i := 0; i < 100; i++ {
			if ctx.Button("Button\x00" + fmt.Sprintf("%d", i)) {
				g.writeLog(fmt.Sprintf("Pressed button %d in Button Window", i))
			}
		}
	})
}

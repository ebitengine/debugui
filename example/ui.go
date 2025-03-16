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
	ctx.Window("Demo Window", image.Rect(40, 40, 340, 500), func(res debugui.Response, layout debugui.ContainerLayout) {
		// window info
		if ctx.Header("Window Info", false) != 0 {
			ctx.SetGridLayout([]int{54, -1}, nil)
			ctx.Label("Position:")
			ctx.Label(fmt.Sprintf("%d, %d", layout.Bounds.Min.X, layout.Bounds.Min.Y))
			ctx.Label("Size:")
			ctx.Label(fmt.Sprintf("%d, %d", layout.Bounds.Dx(), layout.Bounds.Dy()))
		}

		// labels + buttons
		if ctx.Header("Test Buttons", true) != 0 {
			ctx.SetGridLayout([]int{100, -110, -1}, nil)
			ctx.Label("Test buttons 1:")
			if ctx.Button("Button 1") != 0 {
				g.writeLog("Pressed button 1")
			}
			if ctx.Button("Button 2") != 0 {
				g.writeLog("Pressed button 2")
			}
			ctx.Label("Test buttons 2:")
			if ctx.Button("Button 3") != 0 {
				g.writeLog("Pressed button 3")
			}
			if ctx.Button("Popup") != 0 {
				ctx.OpenPopup("Test Popup")
			}
			ctx.Popup("Test Popup", func(res debugui.Response, layout debugui.ContainerLayout) {
				ctx.Button("Hello")
				ctx.Button("World")
			})
		}

		// tree
		if ctx.Header("Tree and Text", true) != 0 {
			ctx.SetGridLayout([]int{140, -1}, nil)
			ctx.Division(func() {
				ctx.TreeNode("Test 1", func(res debugui.Response) {
					ctx.TreeNode("Test 1a", func(res debugui.Response) {
						ctx.Label("Hello")
						ctx.Label("World")
					})
					ctx.TreeNode("Test 1b", func(res debugui.Response) {
						if ctx.Button("Button 1") != 0 {
							g.writeLog("Pressed button 1")
						}
						if ctx.Button("Button 2") != 0 {
							g.writeLog("Pressed button 2")
						}
					})
				})
				ctx.TreeNode("Test 2", func(res debugui.Response) {
					ctx.SetGridLayout([]int{54, 54}, nil)
					if ctx.Button("Button 3") != 0 {
						g.writeLog("Pressed button 3")
					}
					if ctx.Button("Button 4") != 0 {
						g.writeLog("Pressed button 4")
					}
					if ctx.Button("Button 5") != 0 {
						g.writeLog("Pressed button 5")
					}
					if ctx.Button("Button 6") != 0 {
						g.writeLog("Pressed button 6")
					}
				})
				ctx.TreeNode("Test 3", func(res debugui.Response) {
					ctx.Checkbox("Checkbox 1", &g.checks[0])
					ctx.Checkbox("Checkbox 2", &g.checks[1])
					ctx.Checkbox("Checkbox 3", &g.checks[2])
				})
			})

			ctx.Text("Lorem ipsum dolor sit amet, consectetur adipiscing " +
				"elit. Maecenas lacinia, sem eu lacinia molestie, mi risus faucibus " +
				"ipsum, eu varius magna felis a nulla.")
		}

		// background color sliders
		if ctx.Header("Background Color", true) != 0 {
			ctx.SetGridLayout([]int{-78, -1}, []int{74})
			// sliders
			ctx.Division(func() {
				ctx.SetGridLayout([]int{46, -1}, nil)
				ctx.Label("Red:")
				ctx.Slider(&g.bg[0], 0, 255, 1, 0)
				ctx.Label("Green:")
				ctx.Slider(&g.bg[1], 0, 255, 1, 0)
				ctx.Label("Blue:")
				ctx.Slider(&g.bg[2], 0, 255, 1, 0)
			})
			// color preview
			ctx.Control("", func(bounds image.Rectangle) debugui.Response {
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
				return 0
			})
		}

		// Number
		if ctx.Header("Number", true) != 0 {
			ctx.SetGridLayout([]int{-1}, nil)
			ctx.Number(&g.num1, 0.1, 2)
			ctx.Slider(&g.num2, 0, 10, 0.1, 2)
		}
	})
}

func (g *Game) logWindow(ctx *debugui.Context) {
	ctx.Window("Log Window", image.Rect(350, 40, 650, 290), func(res debugui.Response, layout debugui.ContainerLayout) {
		// output text panel
		ctx.SetGridLayout([]int{-1}, []int{-25, 0})
		ctx.Panel("Log Output", func(layout debugui.ContainerLayout) {
			ctx.SetGridLayout([]int{-1}, []int{-1})
			ctx.Text(g.logBuf)
			if g.logUpdated {
				ctx.SetScroll(image.Pt(layout.ScrollOffset.X, layout.ContentSize.Y))
				g.logUpdated = false
			}
		})
		ctx.Division(func() {
			// input textbox + submit button
			var submitted bool
			ctx.SetGridLayout([]int{-70, -1}, nil)
			if ctx.TextBox(&g.logSubmitBuf)&debugui.ResponseSubmit != 0 {
				ctx.SetFocus()
				submitted = true
			}
			if ctx.Button("Submit") != 0 {
				submitted = true
			}
			if submitted {
				g.writeLog(g.logSubmitBuf)
				g.logSubmitBuf = ""
			}
		})
	})
}

func (g *Game) buttonWindows(ctx *debugui.Context) {
	ctx.Window("Button Windows", image.Rect(350, 300, 650, 500), func(res debugui.Response, layout debugui.ContainerLayout) {
		ctx.SetGridLayout([]int{100, 100, 100, 100}, nil)
		for i := 0; i < 100; i++ {
			if ctx.Button("Button\x00"+fmt.Sprintf("%d", i)) != 0 {
				g.writeLog(fmt.Sprintf("Pressed button %d in Button Window", i))
			}
		}
	})
}

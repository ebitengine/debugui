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
			ctx.SetGridLayout([]int{-1, -1}, nil)
			ctx.Text("Position:")
			ctx.Text(fmt.Sprintf("%d, %d", layout.Bounds.Min.X, layout.Bounds.Min.Y))
			ctx.Text("Size:")
			ctx.Text(fmt.Sprintf("%d, %d", layout.Bounds.Dx(), layout.Bounds.Dy()))
		})
		ctx.Header("Game Config", true, func() {
			ctx.Checkbox(&g.hiRes, "Hi-Res").On(func() {
				if g.hiRes {
					ctx.SetScale(2)
				} else {
					ctx.SetScale(1)
				}
				g.resetPosition()
			})
		})
		ctx.Header("Test Buttons", true, func() {
			ctx.SetGridLayout([]int{-2, -1, -1}, nil)
			ctx.Text("Test buttons 1:")
			ctx.Button("Button 1").On(func() {
				g.writeLog("Pressed button 1")
			})
			ctx.Button("Button 2").On(func() {
				g.writeLog("Pressed button 2")
			})
			ctx.Text("Test buttons 2:")
			ctx.Button("Button 3").On(func() {
				g.writeLog("Pressed button 3")
			})
			popupID := ctx.Popup(func(layout debugui.ContainerLayout, id debugui.PopupID) {
				ctx.Button("Hello")
				ctx.Button("World")
				ctx.Button("Close").On(func() {
					ctx.ClosePopup(id)
				})
			})
			ctx.Button("Popup").On(func() {
				ctx.OpenPopup(popupID)
			})
		})
		ctx.Header("Tree and Text", true, func() {
			ctx.SetGridLayout([]int{-1, -1}, nil)
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.TreeNode("Test 1", func() {
					ctx.TreeNode("Test 1a", func() {
						ctx.Text("Hello")
						ctx.Text("World")
					})
					ctx.TreeNode("Test 1b", func() {
						ctx.Button("Button 1").On(func() {
							g.writeLog("Pressed button 1")
						})
						ctx.Button("Button 2").On(func() {
							g.writeLog("Pressed button 2")
						})
					})
				})
				ctx.TreeNode("Test 2", func() {
					ctx.SetGridLayout([]int{-1, -1}, nil)
					ctx.Button("Button 3").On(func() {
						g.writeLog("Pressed button 3")
					})
					ctx.Button("Button 4").On(func() {
						g.writeLog("Pressed button 4")
					})
					ctx.Button("Button 5").On(func() {
						g.writeLog("Pressed button 5")
					})
					ctx.Button("Button 6").On(func() {
						g.writeLog("Pressed button 6")
					})
				})
				ctx.TreeNode("Test 3", func() {
					ctx.Checkbox(&g.checks[0], "Checkbox 1")
					ctx.Checkbox(&g.checks[1], "Checkbox 2")
					ctx.Checkbox(&g.checks[2], "Checkbox 3")
				})
			})

			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.Header("English Text", false, func() {
					ctx.Text("Lorem ipsum dolor sit amet, consectetur adipiscing " +
						"elit. Maecenas lacinia, sem eu lacinia molestie, mi risus faucibus " +
						"ipsum, eu varius magna felis a nulla.")
				})
				ctx.Header("Japanese Text", false, func() {
					ctx.Text("隴西の李徴は博学才穎、天宝の末年、若くして名を虎榜に連ね、" +
						"ついで江南尉に補せられたが、性、狷介、自ら恃むところ頗る厚く、" +
						"賤吏に甘んずるを潔しとしなかった。")
				})
			})
		})
		ctx.Header("Color", true, func() {
			ctx.SetGridLayout([]int{-3, -1}, []int{54})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-1, -3}, nil)
				ctx.Text("Red:")
				ctx.Slider(&g.bg[0], 0, 255, 1)
				ctx.Text("Green:")
				ctx.Slider(&g.bg[1], 0, 255, 1)
				ctx.Text("Blue:")
				ctx.Slider(&g.bg[2], 0, 255, 1)
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
					scale := ctx.Scale()
					vector.DrawFilledRect(
						screen,
						float32(bounds.Min.X*scale),
						float32(bounds.Min.Y*scale),
						float32(bounds.Dx()*scale),
						float32(bounds.Dy()*scale),
						color.RGBA{byte(g.bg[0]), byte(g.bg[1]), byte(g.bg[2]), 255},
						false)
					txt := fmt.Sprintf("#%02X%02X%02X", int(g.bg[0]), int(g.bg[1]), int(g.bg[2]))
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64((bounds.Min.X+bounds.Max.X)/2), float64((bounds.Min.Y+bounds.Max.Y)/2))
					op.GeoM.Scale(float64(scale), float64(scale))
					op.PrimaryAlign = text.AlignCenter
					op.SecondaryAlign = text.AlignCenter
					debugui.DrawText(screen, txt, op)
				})
			})
		})
		ctx.Header("Number", true, func() {
			ctx.NumberField(&g.num1, 1)
			ctx.Slider(&g.num2, 0, 1000, 10)
			ctx.NumberFieldF(&g.num3, 0.1, 2)
			ctx.SliderF(&g.num4, 0, 10, 0.1, 2)
		})
		ctx.Header("Licenses", false, func() {
			ctx.Text(`The photograph by Chris Nokleberg is licensed under the Creative Commons Attribution 4.0 License

The Go Gopher by Renee French is licensed under the Creative Commons Attribution 4.0 License.`)
		})
	})
}

func (g *Game) logWindow(ctx *debugui.Context) {
	ctx.Window("Log Window", image.Rect(350, 40, 650, 290), func(layout debugui.ContainerLayout) {
		ctx.SetGridLayout([]int{-1}, []int{-1, 0})
		ctx.Panel(func(layout debugui.ContainerLayout) {
			ctx.SetGridLayout([]int{-1}, []int{-1})
			ctx.Text(g.logBuf)
			if g.logUpdated {
				ctx.SetScroll(image.Pt(layout.ScrollOffset.X, layout.ContentSize.Y))
				g.logUpdated = false
			}
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			var submit bool
			ctx.SetGridLayout([]int{-3, -1}, nil)
			ctx.TextField(&g.logSubmitBuf).On(func() {
				if g.logSubmitBuf != "" && ebiten.IsKeyPressed(ebiten.KeyEnter) {
					submit = true
				}
			})
			textFieldID := ctx.CurrentWidgetID()
			ctx.Button("Submit").On(func() {
				if g.logSubmitBuf != "" {
					submit = true
				}
			})
			if submit {
				g.writeLog(g.logSubmitBuf)
				g.logSubmitBuf = ""
				ctx.SetTextFieldValue(textFieldID, g.logSubmitBuf)
			}
		})
	})
}

func (g *Game) buttonWindows(ctx *debugui.Context) {
	ctx.Window("Button Windows", image.Rect(350, 300, 650, 500), func(layout debugui.ContainerLayout) {
		ctx.SetGridLayout([]int{-1, -1, -1, -1}, nil)
		for i := 0; i < 100; i++ {
			ctx.IDScope(fmt.Sprintf("%d", i), func() {
				ctx.Button("Button").On(func() {
					g.writeLog(fmt.Sprintf("Pressed button %d in Button Window", i))
				})
			})
		}
	})
}

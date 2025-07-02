// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed gophers.jpg
var gophersJPG []byte

type Game struct {
	gopherImage       *ebiten.Image
	x                 int
	y                 int
	vx                int
	vy                int
	hiRes             bool
	needResetPosition bool
	screenWidth       int
	screenHeight      int

	debugUI             debugui.DebugUI
	inputCapturingState debugui.InputCapturingState

	logBuf       string
	logSubmitBuf string
	logUpdated   bool
	bg           [3]int
	checks       [3]bool
	num1_1       int
	num1_2       int
	num2         int
	num3_1       float64
	num3_2       float64
	num4         float64
	num5         int

	selectedOption, anotherSelectedOption   int
	dropdownOptions, anotherDropdownOptions []string
}

func NewGame() (*Game, error) {
	img, _, err := image.Decode(bytes.NewReader(gophersJPG))
	if err != nil {
		return nil, err
	}

	g := &Game{
		gopherImage:       ebiten.NewImageFromImage(img),
		vx:                2,
		vy:                2,
		bg:                [3]int{90, 95, 100},
		checks:            [3]bool{true, false, true},
		needResetPosition: true,
	}

	return g, nil
}

func (g *Game) resetPosition() {
	sW, sH := g.screenWidth, g.screenHeight
	if sW == 0 || sH == 0 {
		return
	}
	imgW, imgH := g.gopherImage.Bounds().Dx(), g.gopherImage.Bounds().Dy()
	g.x = rand.IntN(sW - imgW)
	g.y = rand.IntN(sH - imgH)
}

func (g *Game) Update() error {
	if g.needResetPosition {
		g.resetPosition()
		g.needResetPosition = false
	}

	sW, sH := g.screenWidth, g.screenHeight
	imgW, imgH := g.gopherImage.Bounds().Dx(), g.gopherImage.Bounds().Dy()
	g.x += g.vx
	g.y += g.vy
	if g.x < 0 || sW-imgW <= g.x {
		g.vx *= -1
	}
	if g.y < 0 || sH-imgH <= g.y {
		g.vy *= -1
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	inputCaptured, err := g.debugUI.Update(func(ctx *debugui.Context) error {
		g.testWindow(ctx)
		g.logWindow(ctx)
		g.buttonWindows(ctx)
		g.dropdownWindows(ctx)
		return nil
	})
	if err != nil {
		return err
	}
	g.inputCapturingState = inputCaptured
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x40, 0x40, 0x80, 0xff})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.x), float64(g.y))
	screen.DrawImage(g.gopherImage, op)

	var msgs []string
	if g.inputCapturingState&debugui.InputCapturingStateHover != 0 {
		msgs = append(msgs, "Hovering")
	}
	if g.inputCapturingState&debugui.InputCapturingStateFocus != 0 {
		msgs = append(msgs, "Focusing")
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Input Capturing State: %s", strings.Join(msgs, ", ")))

	g.debugUI.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	scale := 1
	if g.hiRes {
		scale = 2
	}
	sw, sh := outsideWidth*scale, outsideHeight*scale
	if sw != g.screenWidth || sh != g.screenHeight {
		g.screenWidth = sw
		g.screenHeight = sh
		g.needResetPosition = true
	}
	return sw, sh
}

func main() {
	ebiten.SetWindowTitle("Ebitengine DebugUI Demo")
	ebiten.SetWindowSize(960, 640)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	g, err := NewGame()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

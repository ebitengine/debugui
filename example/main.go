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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/ebitengine/debugui"
)

//go:embed gophers.jpg
var gophersJPG []byte

type Game struct {
	gopherImage *ebiten.Image
	x           int
	y           int
	vx          int
	vy          int
	hiRes       bool

	debugUI             debugui.DebugUI
	inputCapturingState debugui.InputCapturingState

	logBuf       string
	logSubmitBuf string
	logUpdated   bool
	bg           [3]int
	checks       [3]bool
	num1         int
	num2         int
	num3         float64
	num4         float64
}

func NewGame() (*Game, error) {
	img, _, err := image.Decode(bytes.NewReader(gophersJPG))
	if err != nil {
		return nil, err
	}

	g := &Game{
		gopherImage: ebiten.NewImageFromImage(img),
		vx:          2,
		vy:          2,
		bg:          [3]int{90, 95, 100},
		checks:      [3]bool{true, false, true},
	}
	g.resetPosition()

	return g, nil
}

func (g *Game) resetPosition() {
	sW, sH := g.screenSize()
	imgW, imgH := g.gopherImage.Bounds().Dx(), g.gopherImage.Bounds().Dy()
	g.x = rand.IntN(sW - imgW)
	g.y = rand.IntN(sH - imgH)
}

func (g *Game) Update() error {
	sW, sH := g.screenSize()
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
	if g.inputCapturingState&debugui.InputCapturingStateHovering != 0 {
		msgs = append(msgs, "Hovering")
	}
	if g.inputCapturingState&debugui.InputCapturingStateFocusing != 0 {
		msgs = append(msgs, "Focusing")
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Input Capturing State: %s", strings.Join(msgs, ", ")))

	g.debugUI.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenSize()
}

func (g *Game) screenSize() (int, int) {
	scale := 1
	if g.hiRes {
		scale = 2
	}
	return 1280 * scale, 960 * scale
}

func main() {
	ebiten.SetWindowTitle("Ebitengine DebugUI Demo")
	ebiten.SetWindowSize(1280, 960)
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

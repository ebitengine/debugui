// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	debugUI *debugui.DebugUI

	// Dropdown demo data
	resolutionOptions  []string
	selectedResolution int

	qualityOptions  []string
	selectedQuality int

	nameOptions  []string
	selectedName int

	// Log for events
	logBuffer []string
}

func NewGame() *Game {
	return &Game{
		debugUI: &debugui.DebugUI{},

		resolutionOptions:  []string{"720p", "1080p", "1440p", "4K", "Can it run Crysis?"},
		selectedResolution: 1, // Default to 1080p

		qualityOptions:  []string{"Low", "Medium", "High", "Ultra"},
		selectedQuality: 2, // Default to High

		nameOptions:  []string{"Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Gopher", "Hajime", "Isabella", "Jack"},
		selectedName: 6, // Default to Gopher

		logBuffer: []string{},
	}
}

func (g *Game) addLog(message string) {
	g.logBuffer = append(g.logBuffer, message)
	// Keep only last 10 messages
	if len(g.logBuffer) > 10 {
		g.logBuffer = g.logBuffer[1:]
	}
}

func (g *Game) Update() error {
	_, err := g.debugUI.Update(func(ctx *debugui.Context) error {
		// Main demo window
		ctx.Window("Simple Dropdown Demo", image.Rect(50, 50, 400, 280), func(layout debugui.ContainerLayout) {
			ctx.Header("Settings", true, func() {
				ctx.SetGridLayout([]int{100, -1}, nil)

				// Resolution dropdown
				ctx.Text("Resolution:")
				ctx.Dropdown(&g.selectedResolution, g.resolutionOptions).On(func() {
					g.addLog(fmt.Sprintf("Resolution: %s", g.resolutionOptions[g.selectedResolution]))
				})

				// Quality dropdown
				ctx.Text("Quality:")
				ctx.Dropdown(&g.selectedQuality, g.qualityOptions).On(func() {
					g.addLog(fmt.Sprintf("Quality: %s", g.qualityOptions[g.selectedQuality]))
				})
				ctx.Text("Reset Quality:")
				ctx.Button("Reset").On(func() { // used to debug the dropdown above, clicking low also clicks on this
					g.selectedQuality = 2 // Reset to High
				})
				// Name dropdown
				ctx.Text("Name:")
				ctx.Dropdown(&g.selectedName, g.nameOptions).On(func() {
					g.addLog(fmt.Sprintf("Name: %s", g.nameOptions[g.selectedName]))
				})
			})

			ctx.Header("Current Selection", false, func() {
				ctx.Text(fmt.Sprintf("Resolution: %s", g.resolutionOptions[g.selectedResolution]))
				ctx.Text(fmt.Sprintf("Quality: %s", g.qualityOptions[g.selectedQuality]))
				ctx.Text(fmt.Sprintf("Name: %s", g.nameOptions[g.selectedName]))
			})

			ctx.Header("Actions", false, func() {
				ctx.SetGridLayout([]int{-1, -1}, nil)
				ctx.Button("Reset Settings").On(func() {
					g.selectedResolution = 1
					g.selectedQuality = 2
					g.selectedName = 6
					g.addLog("Settings reset to defaults")
				})
				ctx.Button("Clear Log").On(func() {
					g.logBuffer = []string{}
				})
			})
		})

		return nil
	})

	return err
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.debugUI.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowTitle("Simple Dropdown Demo")
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

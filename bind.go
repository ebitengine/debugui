// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"sync"

	"github.com/hajimehoshi/bitmapfont/v3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var fontFace = text.NewGoXFace(bitmapfont.Face)

func DrawText(dst *ebiten.Image, str string, op *text.DrawOptions) {
	text.Draw(dst, str, fontFace, op)
}

func textWidth(str string) int {
	return int(text.Advance(str, fontFace))
}

func lineHeight() int {
	return int(fontFace.Metrics().HAscent + fontFace.Metrics().HDescent + fontFace.Metrics().HLineGap)
}

var (
	//go:embed icon/*.png
	iconFS  embed.FS
	iconMap = map[icon]*ebiten.Image{}
	iconM   sync.Mutex
)

func iconImage(icon icon) *ebiten.Image {
	iconM.Lock()
	defer iconM.Unlock()

	if img, ok := iconMap[icon]; ok {
		return img
	}

	var name string
	switch icon {
	case iconCheck:
		name = "check.png"
	case iconClose:
		name = "close.png"
	case iconCollapsed:
		name = "collapsed.png"
	case iconExpanded:
		name = "expanded.png"
	default:
		return nil
	}
	b, err := iconFS.ReadFile("icon/" + name)
	if err != nil {
		panic(fmt.Sprintf("debugui: %v", err))
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		panic(fmt.Sprintf("debugui: %v", err))
	}
	iconMap[icon] = ebiten.NewImageFromImage(img)
	return iconMap[icon]
}

func (c *Context) draw(screen *ebiten.Image) {
	target := screen
	var cmd *command
	for c.nextCommand(&cmd) {
		switch cmd.typ {
		case commandRect:
			vector.DrawFilledRect(
				target,
				float32(cmd.rect.rect.Min.X),
				float32(cmd.rect.rect.Min.Y),
				float32(cmd.rect.rect.Dx()),
				float32(cmd.rect.rect.Dy()),
				cmd.rect.color,
				false,
			)
		case commandText:
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(cmd.text.pos.X), float64(cmd.text.pos.Y))
			op.ColorScale.ScaleWithColor(cmd.text.color)
			text.Draw(target, cmd.text.str, fontFace, op)
		case commandIcon:
			img := iconImage(cmd.icon.icon)
			if img == nil {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			x := cmd.icon.rect.Min.X + (cmd.icon.rect.Dx()-img.Bounds().Dx())/2
			y := cmd.icon.rect.Min.Y + (cmd.icon.rect.Dy()-img.Bounds().Dy())/2
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorScale.ScaleWithColor(cmd.icon.color)
			target.DrawImage(img, op)
		case commandDraw:
			cmd.draw.f(target)
		case commandClip:
			target = screen.SubImage(cmd.clip.rect).(*ebiten.Image)
		}
	}
}

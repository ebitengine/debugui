// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package debugui

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	clipPart = 1 + iota
	clipAll
)

var (
	unclippedRect = image.Rect(0, 0, 0x1000000, 0x1000000)
)

var fontFace = text.NewGoXFace(bitmapfont.Face)

// DrawText draws the text on the destination image with the given options,
// in the same way as debugui's widget text drawing.
//
// Note that you have to specify the scale at the options when the context scale is not 1.
func DrawText(dst *ebiten.Image, str string, options *text.DrawOptions) {
	text.Draw(dst, str, fontFace, options)
}

func textWidth(str string) int {
	return int(text.Advance(str, fontFace))
}

func lineHeight() int {
	return int(fontFace.Metrics().HAscent + fontFace.Metrics().HDescent + fontFace.Metrics().HLineGap)
}

type icon int

const (
	iconCheck icon = iota + 1
	iconCollapsed
	iconExpanded
	iconDown
	iconUp
)

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
	case iconCollapsed:
		name = "collapsed.png"
	case iconExpanded:
		name = "expanded.png"
	case iconDown:
		name = "down.png"
	case iconUp:
		name = "up.png"
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
	if c.err != nil {
		return
	}

	target := screen
	scale := c.Scale()
	for cmd := range c.commands() {
		switch cmd.typ {
		case commandRect:
			vector.DrawFilledRect(
				target,
				float32(cmd.rect.rect.Min.X*scale),
				float32(cmd.rect.rect.Min.Y*scale),
				float32(cmd.rect.rect.Dx()*scale),
				float32(cmd.rect.rect.Dy()*scale),
				cmd.rect.color,
				false,
			)
		case commandText:
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(cmd.text.pos.X), float64(cmd.text.pos.Y))
			op.GeoM.Scale(float64(scale), float64(scale))
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
			op.GeoM.Scale(float64(scale), float64(scale))
			op.ColorScale.ScaleWithColor(cmd.icon.color)
			target.DrawImage(img, op)
		case commandDraw:
			cmd.draw.f(target)
		case commandClip:
			r := cmd.clip.rect
			r.Min.X *= scale
			r.Min.Y *= scale
			r.Max.X *= scale
			r.Max.Y *= scale
			target = screen.SubImage(r).(*ebiten.Image)
		}
	}
}

func (c *Context) drawRect(rect image.Rectangle, color color.Color) {
	rect2 := rect.Intersect(c.clipRect())
	if rect2.Dx() > 0 && rect2.Dy() > 0 {
		cmd := c.appendCommand(commandRect)
		cmd.rect.rect = rect2
		cmd.rect.color = color
	}
}

func (c *Context) drawBox(rect image.Rectangle, color color.Color) {
	c.drawRect(image.Rect(rect.Min.X+1, rect.Min.Y, rect.Max.X-1, rect.Min.Y+1), color)
	c.drawRect(image.Rect(rect.Min.X+1, rect.Max.Y-1, rect.Max.X-1, rect.Max.Y), color)
	c.drawRect(image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+1, rect.Max.Y), color)
	c.drawRect(image.Rect(rect.Max.X-1, rect.Min.Y, rect.Max.X, rect.Max.Y), color)
}

func (c *Context) drawText(str string, pos image.Point, color color.Color) {
	rect := image.Rect(pos.X, pos.Y, pos.X+textWidth(str), pos.Y+lineHeight())
	clipped := c.checkClip(rect)
	if clipped == clipAll {
		return
	}
	if clipped == clipPart {
		c.setClip(c.clipRect())
	}
	// add command
	cmd := c.appendCommand(commandText)
	cmd.text.str = str
	cmd.text.pos = pos
	cmd.text.color = color
	// reset clipping if it was set
	if clipped != 0 {
		c.setClip(unclippedRect)
	}
}

func (c *Context) drawIcon(icon icon, rect image.Rectangle, color color.Color) {
	// do clip command if the rect isn't fully contained within the cliprect
	clipped := c.checkClip(rect)
	if clipped == clipAll {
		return
	}
	if clipped == clipPart {
		c.setClip(c.clipRect())
	}
	// do icon command
	cmd := c.appendCommand(commandIcon)
	cmd.icon.icon = icon
	cmd.icon.rect = rect
	cmd.icon.color = color
	// reset clipping if it was set
	if clipped != 0 {
		c.setClip(unclippedRect)
	}
}

// DrawOnlyWidget adds a widget that only draws the given function without user interaction.
func (c *Context) DrawOnlyWidget(f func(screen *ebiten.Image)) {
	_ = c.wrapEventHandlerAndError(func() (EventHandler, error) {
		_, _ = c.widget(widgetID{}, 0, nil, nil, func(bounds image.Rectangle) {
			c.setClip(c.clipRect())
			defer c.setClip(unclippedRect)
			cmd := c.appendCommand(commandDraw)
			cmd.draw.f = f
		})
		return nil, nil
	})
}

func (c *Context) drawFrame(rect image.Rectangle, colorid int) {
	c.drawRect(rect, c.style().colors[colorid])
	if colorid == colorScrollBase || colorid == colorTitleBG {
		return
	}
	if c.style().colors[colorBorder].A != 0 {
		c.drawBox(rect.Inset(-1), c.style().colors[colorBorder])
	}
}

func (c *Context) drawWidgetFrame(id widgetID, rect image.Rectangle, colorid int, opt option) {
	if (opt & optionNoFrame) != 0 {
		return
	}
	if c.focus == id {
		colorid += 2
	} else if c.hover == id {
		colorid++
	}
	c.drawFrame(rect, colorid)
}

func (c *Context) drawWidgetText(str string, rect image.Rectangle, colorid int, opt option) {
	var pos image.Point
	tw := textWidth(str)
	c.pushClipRect(rect)
	pos.Y = rect.Min.Y + (rect.Dy()-lineHeight())/2
	if (opt & optionAlignCenter) != 0 {
		pos.X = rect.Min.X + (rect.Dx()-tw)/2
	} else if (opt & optionAlignRight) != 0 {
		pos.X = rect.Min.X + rect.Dx() - tw - c.style().padding
	} else {
		pos.X = rect.Min.X + c.style().padding
	}
	c.drawText(str, pos, c.style().colors[colorid])
	c.popClipRect()
}

func (c *Context) setClip(rect image.Rectangle) {
	cmd := c.appendCommand(commandClip)
	cmd.clip.rect = rect
}

func (c *Context) pushClipRect(rect image.Rectangle) {
	last := c.clipRect()
	c.clipStack = append(c.clipStack, rect.Intersect(last))
}

func (c *Context) popClipRect() {
	c.clipStack = c.clipStack[:len(c.clipStack)-1]
}

func (c *Context) clipRect() image.Rectangle {
	return c.clipStack[len(c.clipStack)-1]
}

func (c *Context) checkClip(bounds image.Rectangle) int {
	cr := c.clipRect()
	if !bounds.Overlaps(cr) {
		return clipAll
	}
	if bounds.In(cr) {
		return 0
	}
	return clipPart
}

// SetScale sets the scale of the UI.
//
// The scale affects the rendering result of the UI.
//
// The default scale is 1.
func (c *Context) SetScale(scale int) {
	if scale < 1 {
		panic("debugui: scale must be >= 1")
	}
	c.scaleMinus1 = scale - 1
}

// Scale returns the scale of the UI.
func (c *Context) Scale() int {
	return c.scaleMinus1 + 1
}

func (c *Context) style() *style {
	return &defaultStyle
}

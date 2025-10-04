# DebugUI for Ebitengine

DebugUI is a UI toolkit for Ebitengine applications, primarily intended for debugging purposes.

**The project is currently under development. The API is subject to frequent changes.**

![Example](./example.png)

DebugUI is based on [Microui](https://github.com/rxi/microui). The original Microui was developed by [@rxi](https://github.com/rxi/microui). The original Go port was developed by [@zeozeozeo](https://github.com/zeozeozeo) and [@Zyko0](https://github.com/Zyko0).

## Example Go file

```go
package main

import (
	"fmt"
	"image"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	debugui debugui.DebugUI
}

func (g *Game) Update() error {
	if _, err := g.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Debugui Window", image.Rect(0, 0, 320, 240), func(layout debugui.ContainerLayout) {
			// Place all your widgets inside a ctx.Window's callback.
			ctx.Text("Some text")

			// Use Loop if you ever need to make a loop to make widgets.
			const loopCount = 10
			ctx.Loop(loopCount, func(index int) {
				// Specify a presssing-button event handler by On.
				ctx.Button(fmt.Sprintf("Button %d", index)).On(func() {
					fmt.Printf("Button %d is pressed")
				})
			})
		})
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.debugui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	game := &Game{}
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
```

## License

DebugUI for Ebitengine is licensed under Apache license version 2.0. See [LICENSE](LICENSE) file.

### Microui

https://github.com/rxi/microui

```
Copyright (c) 2024 rxi

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

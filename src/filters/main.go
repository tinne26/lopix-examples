package main

import "github.com/tinne26/lopix"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/ebitenutil"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

type Game struct {}

func (self *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		lopix.SetScalingFilter((lopix.GetScalingFilter() + 1) % 9)
		lopix.Redraw().Request()
		lopix.Redraw().ScheduleClear()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		lopix.SetScalingFilter((lopix.GetScalingFilter() + 8) % 9)
		lopix.Redraw().Request()
		lopix.Redraw().ScheduleClear()
	}

	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	if !lopix.Redraw().Pending() { return }
	canvas.WritePixels([]byte{
		255, 128,  0, 255, /**/  67, 178,  34, 255, /**/ 203, 12, 255, 255, /**/ 40, 30, 180, 255,
		136,  70, 94, 255, /**/ 200, 221,  73, 255, /**/  6, 240, 219, 255, /**/ 12, 24, 220, 255,
		 40,  21, 67, 255, /**/ 240, 116, 203, 255, /**/ 143, 180, 28, 255, /**/ 99, 199, 75, 255,
		180, 255, 25, 255, /**/  10,  82, 240, 255, /**/  88, 84, 135, 255, /**/ 204, 8, 141, 255,
	})
	lopix.QueueHiResDraw(self.infoDraw)
}

func (self *Game) infoDraw(canvas *ebiten.Image) {
	canvas = canvas.SubImage(lopix.HiResActiveArea()).(*ebiten.Image)
	origin := canvas.Bounds().Min
	ebitenutil.DebugPrintAt(canvas, "Filter: " + lopix.GetScalingFilter().String(), origin.X, origin.Y)
}

func main() {
	ebiten.SetWindowTitle("lopix-examples/filters")
	lopix.SetResolution(4, 4)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	lopix.AutoResizeWindow()
	ebiten.SetScreenClearedEveryFrame(false)
	lopix.Redraw().SetManaged(true)
	
	err := lopix.Run(&Game{})
	if err != nil { panic(err) }
}

package main

import "time"
import "image"
import "image/color"
import "math/rand/v2"

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/inpututil"
import "github.com/tinne26/lopix"

var Colors []color.RGBA = []color.RGBA{
	{215,  48, 48, 255}, {101,  78, 206, 255}, {65,  175,  79, 255},
	{255,  0,  63, 255}, {255, 199,  17, 255}, {183,  47, 214, 255},
	{255, 22, 216, 255}, {255, 185, 173, 255}, { 37, 186, 109, 255},
	{104, 216, 82, 255}, {232, 158, 109, 255}, { 79,  72, 127, 255},
	{89, 127, 124, 255}, {127,  89,  89, 255}, {121, 127,  89, 255},
}

type Game struct {
	start time.Time
	duration time.Duration
	level int
	colors [4]color.RGBA
	answer int
	touches []ebiten.TouchID
}

func (self *Game) Update() error {
	if self.level == -1 { // initialization tick
		for i := range 4 { self.RerollColor(i) }
		self.level = 0
	} else if self.level < 21 { // playing screen
		if self.GetInputDir() == -1 { return nil }
		if self.level == 0 || self.GetInputDir() == self.answer {
			self.level += 1
			switch self.level {
			case  1: self.start = time.Now()
			case 21: self.duration = time.Now().Sub(self.start)
			}
			self.answer = rand.IntN(4)
			self.RerollColor(self.answer)
			for range self.level/3 { self.SwapColors() }
		} else {
			self.duration = time.Now().Sub(self.start)
			self.level = 22
		}
	} else if self.GetInputDir() != -1 { // result screen
		self.level = -1
	}

	return nil
}

func (self *Game) GetInputDir() int {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp)    { return 0 }
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) { return 1 }
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown)  { return 2 }
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft)  { return 3 }
	
	// cursor and touchpad handling
	var x, y int = ebiten.CursorPosition()
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		self.touches = inpututil.AppendJustPressedTouchIDs(self.touches)
		if len(self.touches) == 0 { return -1 }
		x, y = ebiten.TouchPosition(self.touches[0])
		self.touches = self.touches[ : 0]
	}
	fx, fy := lopix.ToRelativeCoords(x, y)
	if fx < fy { // bottom left half
		if 1.0 - fx < fy { return 2 } // down
		return 3 // left
	} else { // upper right half
		if 1.0 - fx < fy { return 1 } // right
		return 0 // up
	}
}

func (self *Game) RerollColor(index int) {
retry:
	clr := Colors[rand.IntN(len(Colors))]	
	for i := range 4 { if clr == self.colors[i] { goto retry } }
	self.colors[index] = clr
}

func (self *Game) SwapColors() {
	a, b := rand.IntN(4), rand.IntN(4)
	if a == self.answer {
		self.answer = b
	} else if b == self.answer {
		self.answer = a
	}
	self.colors[a], self.colors[b] = self.colors[b], self.colors[a]
}

func (self *Game) Draw(canvas *ebiten.Image) {
	canvas.Fill(color.RGBA{216, 243, 255, 255})

	dark := color.RGBA{48, 48, 48, 255}
	if self.level < 21 { // playing screen
		FillArea(canvas,  9,  3, 3, 3, self.colors[0]) // up
		FillArea(canvas, 15,  9, 3, 3, self.colors[1]) // right
		FillArea(canvas,  9, 15, 3, 3, self.colors[2]) // down
		FillArea(canvas,  3,  9, 3, 3, self.colors[3]) // left

		// draw lots of decorations
		for _, bar := range []image.Point{
			{9, 9}, {9, 11}, {9, 1}, {9, 19},
			{3, 7}, {3, 13}, {15, 7}, {15, 13},
		}{ FillArea(canvas, bar.X, bar.Y, 3, 1, dark) }
		for _, bar := range []image.Point{
			{1, 9}, {7, 3}, {13, 3}, {19, 9}, {7, 15}, {13, 15},
		}{ FillArea(canvas, bar.X, bar.Y, 1, 3, dark) }
		for _, pix := range []image.Point{
			{9, 10}, {11, 10}, {7, 7}, {13, 7}, {7, 13}, {13, 13},
		} { Fill(canvas, pix.X, pix.Y, dark) }
	} else { // result screen
		if self.level == 22 { // lose screen case
			canvas.Fill(color.RGBA{0, 0, 0, 255})
		}
		
		var i int64
		millis := self.duration.Milliseconds()
		minutes := millis/60000
		millis -= minutes*60000
		if minutes > 99 { minutes = 99 }
		Fill(canvas, 1, 4, dark); Fill(canvas, 1, 6, dark)
		for i = 0; i < minutes/10; i++ { Fill(canvas, 3 + int(i*2), 4, Colors[0]) }
		for i = 0; i < minutes%10; i++ { Fill(canvas, 3 + int(i*2), 6, Colors[0]) }

		seconds := millis/1000
		millis -= seconds*1000
		Fill(canvas, 1, 9, dark); Fill(canvas, 1, 11, dark)
		for i = 0; i < seconds/10; i++ { Fill(canvas, 3 + int(i*2),  9, Colors[1]) }
		for i = 0; i < seconds%10; i++ { Fill(canvas, 3 + int(i*2), 11, Colors[1]) }

		if millis > 99 { millis /= 10 }
		Fill(canvas, 1, 14, dark); Fill(canvas, 1, 16, dark)
		for i = 0; i < millis/10; i++ { Fill(canvas, 3 + int(i*2), 14, Colors[2]) }
		for i = 0; i < millis%10; i++ { Fill(canvas, 3 + int(i*2), 16, Colors[2]) }
	}
}

func Fill(target *ebiten.Image, x, y int, rgba color.RGBA) {
	FillArea(target, x, y, 1, 1, rgba)
}

func FillArea(target *ebiten.Image, x, y int, w, h int, rgba color.RGBA) {
	rect := image.Rect(x, y, x + w, y + h)
	target.SubImage(rect).(*ebiten.Image).Fill(rgba)
}

func main() {
	ebiten.SetWindowTitle("lopix-examples/color-walk")
	lopix.SetResolution(21, 21)
	lopix.AutoResizeWindow()
	err := lopix.Run(&Game{ level: -1 })
	if err != nil { panic(err) }
}

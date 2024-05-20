package main

import "embed"
import "image"
import "image/png"
import "image/color"

import "github.com/tinne26/lopix"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/ebitenutil"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

// What to look for while testing:
// - Since we use integer positions for both cameras and elements,
//   you will notice that screen scrolling is very jumpy. This is
//   unavoidable on these context, and it's the main reason why
//   lopix should only be used on games without camera scrolling.
//   The example is provided so it can be compared with hipix.
// - At integer scaling factors (the starting windows size should be
//   adjusted to that by default) you will see that all the basic filters
//   provide the same results. The Src* filters will usually make the
//   graphics blurry.
// - If you resize the screen to be bigger, using the Bilinear filter
//   and trying to generate the maximum blurriness possible, you will
//   be able to compare Nearest with Hermite, Bicubic and Bilinear.
//   Depending on the display DPI, it might look like Nearest is the
//   best way to go, but it can create some artifacts like adding a
//   solid isolated pixel on what should have been a sharp corner.
//   I personally favor Hermite for performance/results tradeoff.
// - If you resize the screen much smaller, it will be way easier to
//   see the distortions generated by Nearest filtering. You will
//   also be able to see that Bilinear is inferior to Bicubic and
//   Hermite, producing blurrier results. This is also visible while
//   upscaling, but depending on the game, your sight, the screen DPI
//   and so on, it can be much harder to appreciate.

//go:embed assets/*
var assets embed.FS // see assets/README.md for licensing

const GameWidth, GameHeight = 256, 144

// --- camera ---

type Camera struct {
	rect image.Rectangle
	opts ebiten.DrawImageOptions
}

func (self *Camera) Translate(x, y int) (int, int) {
	return x - self.rect.Min.X, y - self.rect.Min.Y
}

func (self *Camera) CenterAt(x int) {
	delta := x - (self.rect.Min.X + (self.rect.Dx()/2))
	self.rect = self.rect.Add(image.Pt(delta, 0))
}

func (self *Camera) DrawAt(target, source *ebiten.Image, x, y int) {
	rect := source.Bounds().Add(image.Pt(x, y))
	if !rect.Overlaps(self.rect) { return }

	x, y = self.Translate(x, y)
	fx, fy := float64(x), float64(y)
	self.opts.GeoM.Translate(fx, fy)
	target.DrawImage(source, &self.opts)
	self.opts.GeoM.Translate(-fx, -fy)
}

// --- graphic ---

type Graphic struct {
	X, Y int
	Source *ebiten.Image
}

var graphicSources map[string]*ebiten.Image = make(map[string]*ebiten.Image, 32)
func loadGraphic(name string, x, y int) Graphic {
	source, found := graphicSources[name]
	if !found {
		file, err := assets.Open("assets/" + name + ".png")
		if err != nil { panic(err) }
		img, err := png.Decode(file)
		if err != nil { panic(err) }
		source = ebiten.NewImageFromImage(img)
		err = file.Close()
		if err != nil { panic(err) }
		graphicSources[name] = source
	}
	return Graphic{ X: x, Y: y, Source: source }
}

// --- animation ---

type Animation struct {
	frames []*ebiten.Image
	frameDurations []uint8
	frameDurationLeft uint8
	frameIndex, loopIndex uint8
}
func (self *Animation) AddFrame(frame *ebiten.Image, durationTicks uint8) {
	if durationTicks == 0 { panic("durationTicks == 0") }
	self.frames = append(self.frames, frame)
	self.frameDurations = append(self.frameDurations, durationTicks)
	if len(self.frames) == 1 { self.frameDurationLeft = durationTicks }
}
func (self *Animation) Update() {
	self.frameDurationLeft -= 1
	if self.frameDurationLeft == 0 {
		if self.frameIndex == uint8(len(self.frames) - 1) {
			self.frameIndex = self.loopIndex
		} else {
			self.frameIndex += 1
		}
		self.frameDurationLeft = self.frameDurations[self.frameIndex]
	}
}
func (self *Animation) GetFrame() *ebiten.Image {
	return self.frames[self.frameIndex]
}
func (self *Animation) InPreLoopPhase() bool {
	return self.frameIndex < self.loopIndex
}
func (self *Animation) Restart() {
	self.frameIndex = 0
	self.frameDurationLeft = self.frameDurations[0]
}

var IdleAnimation Animation
var MoveAnimation Animation

// --- player ---

type Player struct {
	x, y float64
	animation *Animation
	direction int // -1 = left, 1 = right
	moving bool
}

func (self *Player) Update() {
	self.animation.Update()
	self.updateDirection()
	if self.moving {
		self.ensureAnimation(&MoveAnimation)
		preX := self.x
		switch self.animation.InPreLoopPhase() {
		case true  : self.x += float64(self.direction)*0.48
		case false : self.x += float64(self.direction)*1.24
		}
		self.x = min(max(self.x, 30.5), 285.5 - PlayerFrameWidth)
		if self.x == preX { self.moving = false }
	}
	if !self.moving {
		self.ensureAnimation(&IdleAnimation)
		self.x = float64(int(self.x)) + 0.5
	}
}

func (self *Player) Draw(canvas *ebiten.Image, camera *Camera) {
	x, y := camera.Translate(int(self.x), int(self.y))
	frame := self.animation.GetFrame()
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(-PlayerFrameWidth/2, 0)
	if self.direction == -1 {
		opts.GeoM.Scale(-1, 1)
	}
	opts.GeoM.Translate(float64(x + PlayerFrameWidth/2), float64(y))
	canvas.DrawImage(frame, &opts)
}

func (self *Player) GetCenterX() int {
	return int(self.x) + PlayerFrameWidth/2
}

func (self *Player) updateDirection() {
	self.moving = true
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft)  : self.direction = -1
	case ebiten.IsKeyPressed(ebiten.KeyA)          : self.direction = -1
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight) : self.direction =  1
	case ebiten.IsKeyPressed(ebiten.KeyD)          : self.direction =  1
	default: self.moving = false
	}
}

func (self *Player) ensureAnimation(anim *Animation) {
	if self.animation == anim { return }
	self.animation = anim
	self.animation.Restart()
}

const PlayerFrameWidth  = 17
const PlayerFrameHeight = 51
func getPlayerFrameAt(spritesheet *ebiten.Image, row, col int) *ebiten.Image {
	rect := image.Rect(
		PlayerFrameWidth*col, PlayerFrameHeight*row,
		PlayerFrameWidth*(col + 1), PlayerFrameHeight*(row + 1),
	)
	return spritesheet.SubImage(rect).(*ebiten.Image)
}

// --- game ---

type Game struct {
	backGraphics  []Graphic
	frontGraphics []Graphic
	player Player
	camera Camera
}

func (self *Game) Update() error {
	// scaling filter changes
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		lopix.SetScalingFilter((lopix.GetScalingFilter() + 1) % 9)
		lopix.Redraw().Request()
		lopix.Redraw().ScheduleClear()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		lopix.SetScalingFilter((lopix.GetScalingFilter() + 8) % 9)
		lopix.Redraw().Request()
		lopix.Redraw().ScheduleClear()
	}

	// update player and camera
	self.player.Update()
	self.camera.CenterAt(self.player.GetCenterX())
	return nil
}

func (self *Game) Draw(canvas *ebiten.Image) {
	canvas.Fill(color.RGBA{244, 232, 232, 255})
	for _, graphic := range self.backGraphics {
		self.camera.DrawAt(canvas, graphic.Source, graphic.X, graphic.Y)
	}
	self.player.Draw(canvas, &self.camera)
	for _, graphic := range self.frontGraphics {
		self.camera.DrawAt(canvas, graphic.Source, graphic.X, graphic.Y)
	}
	ebitenutil.DebugPrintAt(canvas, lopix.GetScalingFilter().String() + " filter [page up/down]", 0, GameHeight - 16)
}

func main() {
	ebiten.SetWindowTitle("lopix-examples/gametest")
	lopix.SetResolution(GameWidth, GameHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	lopix.AutoResizeWindow()
	ebiten.SetScreenClearedEveryFrame(false)

	// create animations
	pss := loadGraphic("player", 0, 0).Source // player spritesheet
	idle := getPlayerFrameAt(pss, 0, 0)
	tightFront1 := getPlayerFrameAt(pss, 0, 1)
	IdleAnimation.AddFrame(idle, 255)
	IdleAnimation.AddFrame(tightFront1, 30)
	IdleAnimation.AddFrame(idle, 30)
	IdleAnimation.AddFrame(getPlayerFrameAt(pss, 0, 3), 30)
	IdleAnimation.AddFrame(idle, 30)
	IdleAnimation.AddFrame(tightFront1, 30)
	IdleAnimation.AddFrame(idle, 80)
	
	t := uint8(8)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 1, 0), 11)
	MoveAnimation.loopIndex = 1
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 0), t)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 1), t)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 2), t)
	MoveAnimation.AddFrame(getPlayerFrameAt(pss, 2, 3), t)

	// set up everything for the game
	floorY := GameHeight - 34
	camera := Camera{ rect: image.Rect(0, 0, GameWidth, GameHeight) }
	playerX, playerY := float64(106), float64(floorY - PlayerFrameHeight + 3)
	player := Player{ x: playerX, y: playerY, animation: &IdleAnimation, direction: 1 }
	camera.CenterAt(player.GetCenterX())
	game := &Game{
		backGraphics: []Graphic{
			loadGraphic("dark_floor_left_corner", 30, floorY),
			loadGraphic("dark_floor_side", 30, floorY + 21),
			loadGraphic("dark_floor_center", 67, floorY),
			loadGraphic("dark_floor_center", 127, floorY),
			loadGraphic("dark_floor_center", 187, floorY),
			loadGraphic("dark_floor_right_corner", 247, floorY),
			loadGraphic("dark_floor_side", 247, floorY + 21),
			loadGraphic("platform_flat_horz_small_A", -113, 60),
			loadGraphic("step_small_A", -45, 53),
			loadGraphic("step_small_B", -25, 46),
			loadGraphic("step_small_C", -5, 39),
			loadGraphic("step_small_A", 15, 32),
			loadGraphic("step_long_A", 35, 25),
			loadGraphic("step_small_D", 72, 18),
			loadGraphic("step_small_C", 92, 11),
			loadGraphic("step_long_A", 112, 4),
			loadGraphic("platform_flat_horz_small_A", 183, 26),
			loadGraphic("step_small_D", 251, 20),
			loadGraphic("platform_ground_square_small_B", -37, 93),
			loadGraphic("platform_ground_square_small_A", 310, 88),
			loadGraphic("large_sword_absorbed", 311, 15),
			loadGraphic("right_sign", -78, 33),
			loadGraphic("skeleton_A", 189, 19),
			loadGraphic("back_axe_A", 224, 83),
			loadGraphic("sword_D", -17, 65),
			loadGraphic("back_spear_A", 47, 55),
			loadGraphic("back_skull_A", -27, 87),
			loadGraphic("axe_A", 204, 3),
			loadGraphic("sword_A", 174, 82),
			loadGraphic("skull_B", 165, 102),
			loadGraphic("back_skeleton_A", 44, 103),
			loadGraphic("back_skull_A", 183, 104),
			loadGraphic("back_sword_B", 200, 69),
			loadGraphic("back_spear_B", 255, 55),
			loadGraphic("back_skull_B", 88, 104),
		},
		frontGraphics: []Graphic{
			loadGraphic("sword_B", 74, 69),
			loadGraphic("spear_A", 214, 55),
		},
		camera: camera,
		player: player,
	}

	err := lopix.Run(game)
	if err != nil { panic(err) }
}

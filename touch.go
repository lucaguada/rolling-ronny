package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// On-screen controls for phones/tablets, in logical (320x240) coordinates.
// They appear as soon as a touch is seen and are drawn over the play area, above the HUD.
type button struct {
	r     image.Rectangle
	glyph byte // CP437 arrow in the 8x8 font
}

var (
	btnLeft  = button{image.Rect(4, 150, 48, 202), 0x1B}
	btnRight = button{image.Rect(54, 150, 98, 202), 0x1A}
	btnJump  = button{image.Rect(226, 150, 276, 202), 'A'}
	btnUp    = button{image.Rect(282, 94, 316, 144), 0x18}
	btnDown  = button{image.Rect(282, 150, 316, 202), 0x19}
	btnMusic = button{image.Rect(292, 4, 316, 24), 'M'}
	allBtns  = []button{btnLeft, btnRight, btnJump, btnUp, btnDown, btnMusic}
)

func touchPoints() []image.Point {
	var pts []image.Point
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		pts = append(pts, image.Pt(x, y))
	}
	return pts
}

func anyIn(pts []image.Point, b button) bool {
	for _, p := range pts {
		if p.In(b.r) {
			return true
		}
	}
	return false
}

func justTouched() []image.Point {
	var pts []image.Point
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		pts = append(pts, image.Pt(x, y))
	}
	return pts
}

// touchInput merges the on-screen buttons into the keyboard input.
func (g *Game) touchInput(in Input) Input {
	held, fresh := touchPoints(), justTouched()
	if len(held) > 0 {
		g.touchUI = true
	}
	in.L = in.L || anyIn(held, btnLeft)
	in.R = in.R || anyIn(held, btnRight)
	in.U = in.U || anyIn(held, btnUp)
	in.D = in.D || anyIn(held, btnDown)
	in.JumpHeld = in.JumpHeld || anyIn(held, btnJump)
	in.Jump = in.Jump || anyIn(fresh, btnJump)
	return in
}

func (g *Game) drawTouch(dst *ebiten.Image) {
	if !g.touchUI || g.st != stPlay {
		return
	}
	held := touchPoints()
	for _, b := range allBtns {
		a := float32(0.32)
		if anyIn(held, b) {
			a = 0.5
		}
		r := b.r
		vector.DrawFilledRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), color.NRGBA{255, 255, 255, uint8(a * 255)}, false)
		vector.StrokeRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 1, color.NRGBA{0, 0, 0, 120}, false)
		g.hud.text(dst, string([]byte{b.glyph}), r.Min.X+(r.Dx()-16)/2, r.Min.Y+(r.Dy()-16)/2, 2, [4]float32{1, 1, 1, 1})
	}
}

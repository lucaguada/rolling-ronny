package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// On-screen controls for phones/tablets. They sit in the black side margins next to the 320x240 game
// (logical coordinates; the game is centred, margin() px on each side) so they never cover the action.
const (
	minPad = 90  // minimum margin width when the touch UI is shown
	maxW   = 600 // widest logical screen (ultra-wide phones)
)

type button struct {
	r     image.Rectangle
	glyph byte // CP437 glyph in the 8x8 font
}

type padLayout struct{ left, right, up, down, jump, music button }

func (p padLayout) all() []button { return []button{p.left, p.right, p.up, p.down, p.jump, p.music} }

// pad lays the buttons out for the current logical width: d-pad in the left margin, jump on the right.
func (g *Game) pad() padLayout {
	w := max(g.w, screenW+2*minPad)
	m := (w - screenW) / 2
	bw := (m - 18) / 2
	return padLayout{
		left:  button{image.Rect(6, 150, 6+bw, 202), 0x1B},
		right: button{image.Rect(12+bw, 150, 12+2*bw, 202), 0x1A},
		up:    button{image.Rect(6, 94, 6+bw, 144), 0x18},
		down:  button{image.Rect(12+bw, 94, 12+2*bw, 144), 0x19},
		jump:  button{image.Rect(w-m+6, 94, w-6, 202), 'A'},
		music: button{image.Rect(w-m+6, 6, w-m+6+26, 26), 'M'},
	}
}

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
	p := g.pad()
	in.L = in.L || anyIn(held, p.left)
	in.R = in.R || anyIn(held, p.right)
	in.U = in.U || anyIn(held, p.up)
	in.D = in.D || anyIn(held, p.down)
	in.JumpHeld = in.JumpHeld || anyIn(held, p.jump)
	in.Jump = in.Jump || anyIn(fresh, p.jump)
	return in
}

func (g *Game) drawTouch(dst *ebiten.Image) {
	if !g.touchUI || g.st != stPlay {
		return
	}
	held := touchPoints()
	for _, b := range g.pad().all() {
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

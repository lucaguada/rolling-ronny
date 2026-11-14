package main

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type spark struct {
	x, y, vx, vy float64
	life, max    int
	h            float64 // hue 0..1
}

// fireworks: rockets rise, explode into sparks that fall and fade (title screen).
type fireworks struct{ rockets, sparks []spark }

func (f *fireworks) update() {
	if rand.IntN(30) == 0 {
		f.rockets = append(f.rockets, spark{x: 20 + rand.Float64()*280, y: screenH - 20, vy: -(3 + rand.Float64()*1.5), h: rand.Float64()})
	}
	live := f.rockets[:0]
	for _, r := range f.rockets {
		r.y += r.vy
		r.vy += 0.06
		if r.vy < -0.3 {
			live = append(live, r)
			continue
		}
		for k := 0; k < 40; k++ { // explode
			a, s := 2*math.Pi*float64(k)/40+rand.Float64()*0.2, 0.5+rand.Float64()*1.6
			life := 50 + rand.IntN(25)
			f.sparks = append(f.sparks, spark{x: r.x, y: r.y, vx: math.Cos(a) * s, vy: math.Sin(a) * s, life: life, max: life, h: r.h + rand.Float64()*0.1})
		}
	}
	f.rockets = live
	alive := f.sparks[:0]
	for _, s := range f.sparks {
		s.x, s.y = s.x+s.vx, s.y+s.vy
		s.vx *= 0.985
		s.vy += 0.04
		if s.life--; s.life > 0 {
			alive = append(alive, s)
		}
	}
	f.sparks = alive
}

func (f *fireworks) draw(dst *ebiten.Image) {
	for _, r := range f.rockets {
		vector.FillRect(dst, float32(r.x), float32(r.y), 2, 4, color.White, false)
	}
	for _, s := range f.sparks {
		f := float64(s.life) / float64(s.max)
		vector.FillRect(dst, float32(s.x)-1, float32(s.y)-1, 3, 3, hsv(s.h, min(1, 2*f)), false) // stays bright, fades in the last half
		if f > 0.6 {
			vector.FillRect(dst, float32(s.x), float32(s.y), 1, 1, color.White, false) // hot white core
		}
	}
}

// hsv returns a fully saturated colour of hue h (wraps) scaled by v (premultiplied, so v also fades it).
func hsv(h, v float64) color.RGBA {
	h = (h - math.Floor(h)) * 6
	x := 1 - math.Abs(math.Mod(h, 2)-1)
	c := [6][3]float64{{1, x, 0}, {x, 1, 0}, {0, 1, x}, {0, x, 1}, {x, 0, 1}, {1, 0, x}}[int(h)%6]
	r, g, b := c[0], c[1], c[2]
	return color.RGBA{uint8(r * v * 255), uint8(g * v * 255), uint8(b * v * 255), uint8(v * 255)}
}

// wave draws s centred, each letter bobbing up and down out of phase and cycling through the rainbow.
func (g *Game) wave(dst *ebiten.Image, s string, y int, scale float64) {
	t, x0 := float64(g.frames), (screenW-len(s)*8*int(scale))/2
	for i, c := range []byte(s) {
		dy := 6 * math.Sin(t*0.08+float64(i)*0.55)
		x, yy := x0+i*8*int(scale), y+int(dy)
		k := hsv(t*0.006+float64(i)*0.06, 1)
		g.hud.text(dst, string(c), x+int(scale), yy+int(scale), scale, [4]float32{0, 0, 0, 1})
		g.hud.text(dst, string(c), x, yy, scale, [4]float32{float32(k.R) / 255, float32(k.G) / 255, float32(k.B) / 255, 1})
	}
}

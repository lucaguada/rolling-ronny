package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const hudTop = gameH // panel is 320x32 at the bottom; coordinates below are panel-relative (GAMEPLAY.md section 5)

type HUD struct {
	panel, digits, font          *ebiten.Image
	fullE, emptyE, fullH, emptyH *ebiten.Image
}

func newHUD() *HUD {
	return &HUD{
		panel: loadImage("gfx/panel_bbm.png"), digits: loadImage("sprites/hud/atlas.png"), font: loadFont(),
		fullE: loadImage("gfx/p1_bbm.png"), fullH: loadImage("gfx/p2_bbm.png"),
		emptyE: loadImage("gfx/p3_bbm.png"), emptyH: loadImage("gfx/p4_bbm.png"),
	}
}

// number draws n right-aligned in `width` 8x8 digits starting at x,y (leading digits blank).
func (h *HUD) number(dst *ebiten.Image, n, width, x, y int) {
	for i := width - 1; i >= 0; i-- {
		d, sx := n%10, 0
		if n == 0 && i < width-1 {
			sx = 160 // blank
		} else {
			sx = d * 16
		}
		n /= 10
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x+i*8), float64(y))
		dst.DrawImage(h.digits.SubImage(image.Rect(sx, 0, sx+8, 8)).(*ebiten.Image), op)
	}
}

func (h *HUD) bar(dst, full, empty *ebiten.Image, v, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(empty, op)
	w := full.Bounds().Dx() * v / 84
	if w > 0 {
		dst.DrawImage(full.SubImage(image.Rect(0, 0, w, full.Bounds().Dy())).(*ebiten.Image), op)
	}
}

func (h *HUD) draw(dst *ebiten.Image, g *Game) {
	// the panel has transparent wells (score, time, slots...) that show black behind the digits
	vector.DrawFilledRect(dst, 0, hudTop, screenW, screenH-hudTop, color.Black, false)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, hudTop)
	dst.DrawImage(h.panel, op)
	h.bar(dst, h.fullE, h.emptyE, g.energy, 24, hudTop+5)
	h.bar(dst, h.fullH, h.emptyH, g.health, 23, hudTop+17)
	// digit positions measured on the original screenshot (template match of the digit glyphs, panel top = 209)
	h.number(dst, g.score, 6, 132, hudTop+6)
	h.number(dst, g.pence, 3, 202, hudTop+6)
	h.number(dst, g.cubes, 2, 210, hudTop+17)
	secs := max(g.ticksLeft, 0) / 60
	h.number(dst, secs/60, 1, 148, hudTop+17) // the ':' is part of the panel
	h.number(dst, secs%60/10, 1, 164, hudTop+17)
	h.number(dst, secs%60%10, 1, 172, hudTop+17)
}

// text draws CP437 text from the 8x8 font sheet, scaled.
func (h *HUD) text(dst *ebiten.Image, s string, x, y int, scale float64, clr [4]float32) {
	for i, c := range []byte(s) {
		gx, gy := int(c)%16*8, int(c)/16*8
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(x)+float64(i)*8*scale, float64(y))
		op.ColorScale.SetR(clr[0])
		op.ColorScale.SetG(clr[1])
		op.ColorScale.SetB(clr[2])
		op.ColorScale.SetA(clr[3])
		dst.DrawImage(h.font.SubImage(image.Rect(gx, gy, gx+8, gy+8)).(*ebiten.Image), op)
	}
}

func (h *HUD) centered(dst *ebiten.Image, s string, y int, scale float64) {
	w := int(float64(len(s)) * 8 * scale)
	h.text(dst, s, (screenW-w)/2+int(scale), y+int(scale), scale, [4]float32{0, 0, 0, 1}) // shadow
	h.text(dst, s, (screenW-w)/2, y, scale, [4]float32{1, 1, 1, 1})
}

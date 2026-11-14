package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenW, screenH = 320, 240
	gameH            = 209 // play area; the HUD panel starts at y=209 (verified against assets/ref), its 32nd row is cut off
	mapTop           = 17  // display y of the first tile row
	tileSize         = 16
	mapRows          = 12
)

// tile attribute bits (docs/GAMEPLAY.md section 2)
const (
	attrSolid  = 1
	attrLadder = 4
	attrHazard = 8
)

type Level struct {
	N, Cols int
	tiles   []byte
	attr    [256]byte
	tileImg [256]*ebiten.Image
	sky     *ebiten.Image
	atlas   *ebiten.Image
	defs    map[int]*TypeDef
	ronny   int
	ents    []*Ent
	exitX   int
}

type levelJSON struct {
	Cols int    `json:"cols"`
	Back string `json:"back"`
	Attr []byte `json:"attr"`
}

func loadLevel(n int) *Level {
	dir := fmt.Sprintf("levels/level%d/", n)
	var lj levelJSON
	readJSON(dir+"level.json", &lj)
	l := &Level{N: n, Cols: lj.Cols, tiles: readFile(dir + "map.bin"), defs: map[int]*TypeDef{}}
	copy(l.attr[:], lj.Attr)
	if len(l.tiles) != l.Cols*mapRows {
		panic(fmt.Sprintf("level %d: map size %d != %d*%d", n, len(l.tiles), l.Cols, mapRows))
	}

	sheet := loadImage(dir + "tiles.png")
	for i := range l.tileImg {
		x, y := i%16*tileSize, i/16*tileSize
		l.tileImg[i] = sheet.SubImage(image.Rect(x, y, x+tileSize, y+tileSize)).(*ebiten.Image)
	}

	// sky = one colour per row, taken from column 0 of BACKn.LBM (docs/ENGINE.md section 5)
	back := must(png.Decode(bytes.NewReader(readFile("gfx/" + lowerPNG(lj.Back)))))
	l.sky = ebiten.NewImage(screenW, gameH)
	for y := 0; y < gameH; y++ {
		c := back.At(0, min(y, 199))
		l.sky.SubImage(image.Rect(0, y, screenW, y+1)).(*ebiten.Image).Fill(c)
	}

	sd := fmt.Sprintf("sprites/level%d/", n)
	l.atlas = loadImage(sd + "atlas.png")
	var aj atlasJSON
	readJSON(sd+"atlas.json", &aj)
	l.ronny = aj.RonnyTypedef
	for i := range aj.Typedefs {
		l.defs[aj.Typedefs[i].Idx] = &aj.Typedefs[i]
	}
	l.spawnEntities(sd)
	return l
}

// "BACK1.LBM" -> "back1_lbm.png"
func lowerPNG(s string) string {
	b := []byte(s)
	for i, c := range b {
		switch {
		case c >= 'A' && c <= 'Z':
			b[i] = c + 32
		case c == '.':
			b[i] = '_'
		}
	}
	return string(b) + ".png"
}

// attrAt returns the attribute of the tile under world pixel (px, py) (py = display y).
func (l *Level) attrAt(px, py int) byte {
	r := (py - mapTop) >> 4
	if px < 0 || px >= l.Cols*tileSize || py < mapTop || r >= mapRows {
		return 0
	}
	return l.attr[l.tiles[r*l.Cols+px>>4]]
}

func (l *Level) draw(dst *ebiten.Image, cam int) {
	dst.DrawImage(l.sky, nil)
	var op ebiten.DrawImageOptions
	for c := cam >> 4; c <= (cam+screenW)>>4 && c < l.Cols; c++ {
		for r := 0; r < mapRows; r++ {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(c*tileSize-cam), float64(mapTop+r*tileSize))
			dst.DrawImage(l.tileImg[l.tiles[r*l.Cols+c]], &op)
		}
	}
}

// drawFrame draws frame f of type-def td with its object origin at world (x, y).
func (l *Level) drawFrame(dst *ebiten.Image, td *TypeDef, f int, x, y float64, cam int, flicker bool) {
	if flicker {
		return
	}
	fr := td.Frames[f%len(td.Frames)]
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x+float64(fr.OX)-float64(cam), y+float64(fr.OY))
	dst.DrawImage(l.atlas.SubImage(image.Rect(fr.X, fr.Y, fr.X+fr.W, fr.Y+fr.H)).(*ebiten.Image), op)
}

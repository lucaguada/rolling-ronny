package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const root = "assets/assets/"

//go:embed assets/assets/levels/level*/level.json
//go:embed assets/assets/levels/level*/map.bin
//go:embed assets/assets/levels/level*/tiles.png
//go:embed assets/assets/levels/level*/objects.json
//go:embed assets/assets/sprites/level*/atlas.png
//go:embed assets/assets/sprites/level*/atlas.json
//go:embed assets/assets/sprites/level*/hitboxes.json
//go:embed assets/assets/sprites/hud/atlas.png
//go:embed assets/assets/sprites/hud/atlas.json
//go:embed assets/assets/gfx/*.png
//go:embed assets/assets/fonts/ibm_vga8x8.png
//go:embed assets/assets/sfx/*.wav
//go:embed assets/assets/music/song.ogg
var fsys embed.FS

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func readFile(p string) []byte { return must(fsys.ReadFile(root + p)) }

func readJSON(p string, v any) {
	if err := json.Unmarshal(readFile(p), v); err != nil {
		panic(fmt.Sprintf("%s: %v", p, err))
	}
}

func loadImage(p string) *ebiten.Image {
	return ebiten.NewImageFromImage(must(png.Decode(bytes.NewReader(readFile(p)))))
}

// loadFont turns the 128x128 CP437 sheet (16x16 glyphs of 8x8, index 1 = ink) into white-on-transparent.
func loadFont() *ebiten.Image {
	src := must(png.Decode(bytes.NewReader(readFile("fonts/ibm_vga8x8.png")))).(*image.Paletted)
	dst := image.NewRGBA(src.Bounds())
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			if src.ColorIndexAt(x, y) != 0 {
				dst.Set(x, y, color.White)
			}
		}
	}
	return ebiten.NewImageFromImage(dst)
}

// ---- sprite atlas / object descriptors (JSON written by the extraction tools) ----

type Frame struct {
	X, Y, W, H, OX, OY int
	Hitbox             [4]int // relative to the object's (x,y), inclusive
}

type TypeDef struct {
	Idx    int
	Frames []Frame
}

type atlasJSON struct {
	RonnyTypedef int `json:"ronny_typedef"`
	Typedefs     []TypeDef
}

type spawnJSON struct {
	Type       int `json:"type"`
	Kind       int `json:"kind"`
	TypedefIdx int `json:"typedef_idx"`
	MapX       int `json:"map_x"`
	MapY       int `json:"map_y"`
}

type objectsJSON struct {
	Types []struct {
		Type   int `json:"type"`
		Accept int `json:"accept"`
	} `json:"types"`
	Spawns []spawnJSON `json:"spawns"`
}

// ---- audio ----

type Audio struct {
	ctx   *audio.Context
	sfx   map[string][]byte
	music *audio.Player
	on    bool
}

func newAudio() *Audio {
	a := &Audio{ctx: audio.NewContext(44100), sfx: map[string][]byte{}, on: true}
	for _, n := range []string{"blipblip", "blopp", "bluirp", "explosio", "ronnyko", "hupe", "spacer", "federspo", "fallen2"} {
		s := must(wav.DecodeWithSampleRate(44100, bytes.NewReader(readFile("sfx/"+n+".wav"))))
		a.sfx[n] = must(io.ReadAll(s))
	}
	return a
}

func (a *Audio) play(name string) {
	if a == nil { // headless tests
		return
	}
	if b, ok := a.sfx[name]; ok {
		p := a.ctx.NewPlayerFromBytes(b)
		p.SetVolume(0.6)
		p.Play()
	}
}

// playMusic starts the single looping song (the original has one tune for all levels, docs/SOUND.md);
// it keeps playing across levels. 83.4 s = the song's loop point (song.wav has a short silent tail).
func (a *Audio) playMusic() {
	if a.music != nil {
		return
	}
	s := must(vorbis.DecodeWithSampleRate(44100, bytes.NewReader(readFile("music/song.ogg"))))
	a.music = must(a.ctx.NewPlayer(audio.NewInfiniteLoop(s, int64(83.4*44100)*4)))
	a.music.SetVolume(0.5)
	if a.on {
		a.music.Play()
	}
}

func (a *Audio) stopMusic() {
	if a.music != nil {
		a.music.Close()
		a.music = nil
	}
}

func (a *Audio) toggle() {
	a.on = !a.on
	if a.music != nil {
		if a.on {
			a.music.Play()
		} else {
			a.music.Pause()
		}
	}
}

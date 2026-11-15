package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	lastLevel = 9
	maxMeter  = 84 // health / energy bars (original: 0..0x54)
	levelTime = 5 * 60 * 60
)

type state int

const (
	stTitle state = iota
	stPlay
	stClear
	stOver
	stWin
)

type Game struct {
	st                            state
	n                             int
	level                         *Level
	p                             *Player
	audio                         *Audio
	hud                           *HUD
	title                         *ebiten.Image
	score, pence, cubes, health   int
	energy, ticksLeft, startScore int
	startPence                    int
	cam, frames                   int
	shotPath                      string
	shotAt                        int
	bot                           bool
	touchUI                       bool
	w                             int // logical screen width (Layout)
	frame                         *ebiten.Image
	fx                            fireworks
}

func newGame() *Game {
	g := &Game{audio: newAudio(), hud: newHUD(), title: loadImage("gfx/back1_lbm.png")}
	g.shotPath = os.Getenv("RR_SHOT")
	g.shotAt = 30
	if v, err := strconv.Atoi(os.Getenv("RR_FRAMES")); err == nil {
		g.shotAt = v
	}
	g.bot = os.Getenv("RR_BOT") != ""
	g.touchUI = os.Getenv("RR_TOUCH") != "" // debug: show the on-screen buttons
	if v, err := strconv.Atoi(os.Getenv("RR_LEVEL")); err == nil {
		g.startLevel(v)
		if x, err := strconv.Atoi(os.Getenv("RR_X")); err == nil { // debug: start at world x
			g.p.x, g.p.y = float64(x), 0
		}
	}
	return g
}

func (g *Game) startLevel(n int) {
	g.n, g.st = n, stPlay
	g.level = loadLevel(n)
	g.p = newPlayer(g.level)
	g.health, g.energy, g.ticksLeft, g.cubes = maxMeter, maxMeter, levelTime, 0
	g.startScore, g.startPence = g.score, g.pence
	g.cam = 0
	g.audio.playMusic()
}

func (g *Game) input() Input {
	k := ebiten.IsKeyPressed
	jump := k(ebiten.KeySpace) || k(ebiten.KeyZ)
	in := Input{
		L: k(ebiten.KeyArrowLeft) || k(ebiten.KeyA), R: k(ebiten.KeyArrowRight) || k(ebiten.KeyD),
		U: k(ebiten.KeyArrowUp) || k(ebiten.KeyW), D: k(ebiten.KeyArrowDown) || k(ebiten.KeyS),
		Jump:     inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyZ),
		JumpHeld: jump,
	}
	in = g.touchInput(in)
	if g.bot { // smoke-test driver: run right, hop regularly
		in = Input{R: true, Jump: g.frames%45 == 0, JumpHeld: g.frames%45 < 30}
	}
	return in
}

func (g *Game) confirm() bool {
	if len(justTouched()) > 0 {
		g.touchUI = true
		return true
	}
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

func (g *Game) Update() error {
	g.frames++
	if inpututil.IsKeyJustPressed(ebiten.KeyM) || (g.touchUI && anyIn(justTouched(), g.pad().music)) {
		g.audio.toggle()
	}
	switch g.st {
	case stTitle:
		if birthday {
			g.fx.update()
		}
		if g.confirm() {
			g.score, g.pence = 0, 0
			g.startLevel(1)
		}
	case stPlay:
		g.updatePlay()
	case stClear:
		if g.confirm() {
			if g.n == lastLevel {
				g.st = stWin
			} else {
				g.startLevel(g.n + 1)
			}
		}
	case stOver, stWin:
		if g.confirm() {
			g.audio.stopMusic()
			g.st = stTitle
		}
	}
	return nil
}

func (g *Game) updatePlay() {
	l, p := g.level, g.p
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.audio.stopMusic()
		g.st = stTitle
		return
	}
	p.update(g, l, g.input())
	g.ticksLeft--

	hb := l.defs[l.ronny].Frames[p.frame()%len(l.defs[l.ronny].Frames)].Hitbox
	bx0, by0 := p.x+float64(hb[0]), p.y+float64(hb[1])
	bx1, by1 := p.x+float64(hb[2]), p.y+float64(hb[3])
	live := l.ents[:0]
	for _, e := range l.ents {
		e.update()
		if e.dead && e.deadT > 20 {
			continue
		}
		live = append(live, e)
		if e.dead {
			continue
		}
		ex0, ey0, ex1, ey1 := e.box()
		if e.role == roleDecor || !overlap(bx0, by0, bx1, by1, ex0, ey0, ex1, ey1) {
			continue
		}
		switch e.role {
		case roleEnemy:
			if p.vy > 0 && by1-p.vy <= ey0+8 { // stomp from above
				e.dead = true
				p.vy = -3.2
				g.score += 50
				g.pence++
				g.audio.play("explosio")
			} else {
				p.hurt(g, e.damage)
			}
		case roleHazard:
			p.hurt(g, e.damage)
		case roleExit:
			g.audio.play("federspo")
			g.score += g.ticksLeft / 60 * 10
			g.st = stClear
		default:
			g.collect(e)
		}
	}
	l.ents = live

	if p.cx() >= l.exitX+16 && g.st == stPlay {
		g.st = stClear
	}
	if g.health == 0 || g.ticksLeft <= 0 {
		g.audio.stopMusic()
		g.score, g.pence = g.startScore, g.startPence
		g.st = stOver
	}
	g.cam = max(0, min(int(p.x)-151, l.Cols*tileSize-screenW))
}

func (g *Game) collect(e *Ent) {
	e.dead = true
	g.score += 50
	switch e.role {
	case roleCoin:
		g.pence++
		g.audio.play("blipblip")
	case roleCube:
		g.cubes++
		g.audio.play("blipblip")
	case roleHealth:
		g.health = min(maxMeter, g.health+32)
		g.score += 200
		g.audio.play("bluirp")
	case roleEnergy:
		g.energy = min(maxMeter, g.energy+32)
		g.score += 200
		g.audio.play("bluirp")
	case roleTime:
		g.ticksLeft += 3 * 60 * 60
		g.score += 200
		g.audio.play("bluirp")
	default:
		g.audio.play("spacer")
	}
}

// Draw renders the 320x240 game into g.frame and centres it; the touch buttons live in the black side margins.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.frame == nil {
		g.frame = ebiten.NewImage(screenW, screenH)
	}
	g.drawGame(g.frame)
	screen.Fill(color.Black)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.margin()), 0)
	screen.DrawImage(g.frame, op)
	g.drawTouch(screen)
	g.screenshot(screen)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	switch g.st {
	case stTitle:
		screen.DrawImage(g.title, nil)
		if birthday {
			g.fx.draw(screen)
			g.hud.centered(screen, "ROLLING RONNY", 24, 3)
			g.wave(screen, "Buon Compleanno", 80, 2)
			g.wave(screen, "Erika!", 108, 3)
			g.hud.centered(screen, "PRESS ENTER OR TAP", 170, 1)
		} else {
			g.hud.centered(screen, "ROLLING RONNY", 60, 3)
			g.hud.centered(screen, "PRESS ENTER OR TAP", 130, 1)
		}
		g.hud.centered(screen, "ARROWS/AD MOVE  SPACE/Z JUMP  M MUSIC", 222, 1)
	default:
		g.drawPlay(screen)
		switch g.st {
		case stClear:
			g.hud.centered(screen, "LEVEL "+strconv.Itoa(g.n)+" COMPLETE", 80, 2)
			g.hud.centered(screen, "PRESS ENTER OR TAP", 110, 1)
		case stOver:
			g.hud.centered(screen, "GAME OVER", 80, 3)
			g.hud.centered(screen, "PRESS ENTER OR TAP", 115, 1)
		case stWin:
			g.hud.centered(screen, "YOU FINISHED THE GAME!", 80, 2)
			g.hud.centered(screen, fmt.Sprintf("SCORE %d", g.score), 110, 2)
		}
	}
}

func (g *Game) screenshot(screen *ebiten.Image) {
	if g.shotPath != "" && g.frames == g.shotAt {
		f := must(os.Create(g.shotPath))
		defer f.Close()
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		px := make([]byte, 4*w*h)
		screen.ReadPixels(px)
		rgba := &image.RGBA{Pix: px, Stride: 4 * w, Rect: image.Rect(0, 0, w, h)}
		if err := png.Encode(f, rgba); err != nil {
			log.Fatal(err)
		}
		if g.p != nil {
			fmt.Printf("level=%d state=%d x=%.0f y=%.0f health=%d score=%d pence=%d cubes=%d\n",
				g.n, g.st, g.p.x, g.p.y, g.health, g.score, g.pence, g.cubes)
		}
		os.Exit(0)
	}
}

func (g *Game) drawPlay(screen *ebiten.Image) {
	l := g.level
	l.draw(screen, g.cam)
	for _, e := range l.ents {
		if e.role == roleDecor {
			e.draw(screen, l, g.cam)
		}
	}
	for _, e := range l.ents {
		if e.role != roleDecor {
			e.draw(screen, l, g.cam)
		}
	}
	l.drawFrame(screen, l.defs[l.ronny], g.p.frame(), g.p.x, g.p.y, g.cam, g.p.invuln > 0 && g.p.hurtT == 0 && g.frames/3%2 == 0)
	g.hud.draw(screen, g)
}

// Layout keeps the height at 240 and widens the logical screen to the window's aspect ratio, so the extra
// width becomes black margins. With touch controls active the margins are at least minPad wide.
func (g *Game) Layout(outW, outH int) (int, int) {
	w := max(screenW, min(maxW, screenH*outW/max(outH, 1)))
	if g.touchUI {
		w = max(w, screenW+2*minPad)
	}
	g.w = w
	return w, screenH
}

func (g *Game) margin() int { return (max(g.w, screenW) - screenW) / 2 }

func main() {
	ebiten.SetWindowSize(screenW*3, screenH*3)
	ebiten.SetWindowTitle("Rolling Ronny")
	ebiten.SetTPS(60)
	if v, err := strconv.Atoi(os.Getenv("RR_TPS")); err == nil { // test runs: simulate faster than real time
		ebiten.SetTPS(v)
		ebiten.SetVsyncEnabled(false)
	}
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}

package main

import "math"

type Input struct{ L, R, U, D, Jump, JumpHeld bool }

const (
	moveSpeed = 2.0  // px/tick, as in the original (GAMEPLAY.md section 2)
	jumpV     = 4.4  // peak ~44 px with the jump key held (original: 40 px)
	gHold     = 0.22 // gravity while rising with the key held
	gCut      = 0.55 // gravity once the key is released (short hop)
	gFall     = 0.40
	maxFall   = 4.5
	climbV    = 2.0
	feetOff   = 33 // feet = y+feetOff: first pixel below the sprite
	bodyW     = 29 // sprite-origin -> body centre x
)

// Ronny animation frames (docs/SPRITES.md section 3): index 0 = facing right, 1 = left.
var (
	walkFrames  = [2][]int{{30, 31, 32, 33, 32, 30, 34, 35, 36, 35, 34}, {1, 2, 3, 2, 0, 4, 5, 6, 5, 4}}
	standFrame  = [2]int{30, 0}
	jumpFrames  = [2][]int{{40, 41, 42, 43, 42, 41, 40}, {10, 11, 12, 13, 12, 11, 10}}
	duckFrame   = [2]int{37, 7}
	climbFrames = []int{25, 26, 27, 26, 25, 25, 28, 29, 28, 25}
	hurtFrames  = [2][]int{{44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54}, {14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}}
)

type Player struct {
	x, y, vy      float64
	face          int // 0 right, 1 left
	onGround      bool
	climbing      bool
	ducking       bool
	hurtT, invuln int
	t             int
	safeX, safeY  float64
	walking       bool
}

// newPlayer puts Ronny at the original start column (x=151) on the lowest solid tile there
// (the street: y=160 in level 1, 144 in level 2...).
func newPlayer(l *Level) *Player {
	p := &Player{x: 151}
	for r := mapRows - 1; r >= 0; r-- {
		if l.attrAt(p.cx(), mapTop+r*tileSize)&attrSolid != 0 {
			p.y = float64(mapTop + r*tileSize - feetOff)
			break
		}
	}
	p.safeX, p.safeY = p.x, p.y
	return p
}

func (p *Player) cx() int   { return int(math.Floor(p.x)) + bodyW }
func (p *Player) feet() int { return int(math.Floor(p.y)) + feetOff }

func (p *Player) solidUnder(l *Level, f int) bool {
	return l.attrAt(p.cx()-7, f)&attrSolid != 0 || l.attrAt(p.cx()+5, f)&attrSolid != 0
}

func (p *Player) hurt(g *Game, dmg int) {
	if p.invuln > 0 {
		return
	}
	g.health = max(0, g.health-dmg)
	p.invuln = 120
	if dmg >= 16 {
		p.hurtT = 45
		g.audio.play("ronnyko")
	} else {
		g.audio.play("hupe")
	}
}

func (p *Player) update(g *Game, l *Level, in Input) {
	p.t++
	if p.invuln > 0 {
		p.invuln--
	}
	ax := 0.0
	if p.hurtT > 0 {
		p.hurtT--
		in = Input{}
	}
	if in.L {
		ax, p.face = -moveSpeed, 1
	}
	if in.R {
		ax, p.face = moveSpeed, 0
	}
	p.walking = ax != 0
	px := p.cx()
	feet := p.feet()

	// ladders: attr bit 4 at the feet / just below them
	ladderHere := l.attrAt(px, feet-3)&attrLadder != 0
	ladderBelow := l.attrAt(px, feet+2)&attrLadder != 0
	if !p.climbing && ((in.U && ladderHere) || (in.D && p.onGround && ladderBelow)) {
		p.climbing, p.vy = true, 0
	}
	if p.climbing {
		if in.Jump || (!ladderHere && !ladderBelow) {
			p.climbing = false
		} else {
			if in.U {
				p.y -= climbV
			} else if in.D {
				p.y += climbV
			}
			if !in.U && !in.D && !ladderHere {
				p.climbing = false
			}
			// stepped off the top: snap the feet onto the tile row above so the floor catches us
			if in.U && l.attrAt(px, p.feet()-3)&attrLadder == 0 {
				p.y = float64(((p.feet()-mapTop)>>4)*tileSize + mapTop - feetOff)
				p.climbing = false
			}
			if p.climbing {
				return
			}
		}
	}

	p.ducking = in.D && p.onGround && !ladderBelow && p.hurtT == 0
	if p.ducking {
		ax = 0
	}
	// the original never blocks Ronny horizontally; only the level ends do
	p.x = math.Max(-bodyW+10, math.Min(p.x+ax, float64(l.Cols*tileSize-bodyW-10)))

	if in.Jump && p.onGround && p.hurtT == 0 {
		p.vy, p.onGround = -jumpV, false
		g.audio.play("blopp")
	}
	switch {
	case p.vy < 0 && in.JumpHeld:
		p.vy += gHold
	case p.vy < 0:
		p.vy += gCut
	default:
		p.vy = math.Min(p.vy+gFall, maxFall)
	}

	f0 := p.feet()
	p.y += p.vy
	f1 := p.feet()
	p.onGround = false
	if p.vy >= 0 {
		// one-way platforms: land when the feet cross (or rest on) the top edge of a solid tile
		for f := f0; f <= f1; f++ {
			if (f-mapTop)&15 == 0 && p.solidUnder(l, f) {
				p.y, p.vy, p.onGround = float64(f-feetOff), 0, true
				break
			}
		}
	}
	if p.onGround {
		p.safeX, p.safeY = p.x, p.y
	}

	// hazard tiles at the feet / body
	px, feet = p.cx(), p.feet()
	if l.attrAt(px, feet)&attrHazard != 0 || l.attrAt(px, feet-8)&attrHazard != 0 {
		p.hurt(g, 16)
	}
	// bottom of the screen = pit
	if feet > mapTop+mapRows*tileSize+8 {
		g.audio.play("fallen2")
		g.health = max(0, g.health-16)
		p.x, p.y, p.vy = p.safeX, p.safeY-48, 0
		p.invuln = 120
	}
}

func (p *Player) frame() int {
	k := p.t / 4
	switch {
	case p.hurtT > 0:
		return hurtFrames[p.face][min((45-p.hurtT)/4, 10)]
	case p.climbing:
		return climbFrames[k%len(climbFrames)]
	case !p.onGround:
		if p.vy < -1 {
			return jumpFrames[p.face][min((p.t/4)%7, 3)]
		}
		return jumpFrames[p.face][2]
	case p.ducking:
		return duckFrame[p.face]
	case p.walking:
		return walkFrames[p.face][k%len(walkFrames[p.face])]
	}
	return standFrame[p.face]
}

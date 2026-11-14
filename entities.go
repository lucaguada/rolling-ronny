package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Role int

const (
	roleDecor Role = iota
	roleHazard
	roleEnemy
	roleCoin
	roleCube
	roleHealth
	roleEnergy
	roleTime
	roleItem
	roleExit
)

// Ent is a spawned object: pickup, enemy, hazard, exit or pure decoration.
type Ent struct {
	td          *TypeDef
	role        Role
	damage      int
	x, y        float64
	x0, y0      float64
	dir         float64
	t           int
	dead        bool
	deadT       int
	phase, span int
}

// kind numbers come from docs/OBJECTS.md section 3; accept 0xC0 = deadly, 0x80 = hurts
func roleOf(kind, accept int) Role {
	switch kind {
	case 0:
		if accept != 0 {
			return roleHazard
		}
	case 18:
		return roleEnemy
	case 16, 15:
		return roleCoin
	case 8:
		return roleCube
	case 13:
		return roleHealth
	case 14:
		return roleEnergy
	case 20:
		return roleTime
	case 2, 3, 4, 5, 6, 7:
		return roleItem
	case 19, 21, 39:
		return roleExit
	}
	return roleDecor
}

func (l *Level) spawnEntities(sd string) {
	var oj objectsJSON
	readJSON(fmt.Sprintf("levels/level%d/objects.json", l.N), &oj)
	accept := map[int]int{}
	for _, t := range oj.Types {
		accept[t.Type] = t.Accept
	}
	for i, s := range oj.Spawns {
		td := l.defs[s.TypedefIdx]
		if td == nil || len(td.Frames) == 0 {
			continue
		}
		acc := accept[s.Type]
		e := &Ent{td: td, role: roleOf(s.Kind, acc), x: float64(s.MapX), y: float64(s.MapY + mapTop), dir: 1, phase: i * 3}
		e.x0, e.y0 = e.x, e.y
		switch e.role {
		case roleEnemy, roleHazard:
			e.damage = 8
			if acc&0x40 != 0 {
				e.damage = 16
			}
			e.span = 40
		case roleExit:
			l.exitX = int(e.x)
		}
		l.ents = append(l.ents, e)
	}
	if l.exitX == 0 {
		l.exitX = l.Cols*tileSize - 40
	}
}

func (e *Ent) frame() int {
	if e.dead {
		return 0
	}
	return (e.t + e.phase) / 8
}

// hitbox in world coordinates (inclusive) for the entity's current frame.
func (e *Ent) box() (x0, y0, x1, y1 float64) {
	h := e.td.Frames[e.frame()%len(e.td.Frames)].Hitbox
	return e.x + float64(h[0]), e.y + float64(h[1]), e.x + float64(h[2]), e.y + float64(h[3])
}

func (e *Ent) update() {
	e.t++
	if e.dead {
		e.deadT++
		return
	}
	switch e.role {
	case roleEnemy: // ponytail: plain back-and-forth patrol, the original per-enemy scripts are not interpreted
		e.x += 0.5 * e.dir
		if e.x > e.x0+float64(e.span) || e.x < e.x0-float64(e.span) {
			e.dir = -e.dir
		}
	case roleHazard: // bobs vertically like the original falling/bouncing hazards
		e.y += 0.7 * e.dir
		if e.y > e.y0+float64(e.span) || e.y < e.y0 {
			e.dir = -e.dir
		}
	}
}

func overlap(a0, b0, a1, b1, c0, d0, c1, d1 float64) bool {
	return a0 <= c1 && c0 <= a1 && b0 <= d1 && d0 <= b1
}

func (e *Ent) draw(dst *ebiten.Image, l *Level, cam int) {
	if e.x+64 < float64(cam) || e.x-200 > float64(cam+screenW) {
		return
	}
	l.drawFrame(dst, e.td, e.frame(), e.x, e.y, cam, e.dead && e.deadT/3%2 == 0)
}

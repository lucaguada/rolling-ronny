package main

import "testing"

// Every level must load, and Ronny placed at the original start (151,160) must stand
// safely on solid ground: no hazard hit, no fall into the pit.
func TestLevelsLoadAndStartIsSafe(t *testing.T) {
	for n := 1; n <= lastLevel; n++ {
		l := loadLevel(n)
		if len(l.ents) == 0 {
			t.Errorf("level %d: no entities spawned", n)
		}
		g := &Game{health: maxMeter}
		p := newPlayer(l)
		for i := 0; i < 300; i++ {
			p.update(g, l, Input{})
		}
		if !p.onGround || g.health != maxMeter {
			t.Errorf("level %d: start not safe: onGround=%v y=%.0f health=%d", n, p.onGround, p.y, g.health)
		}
	}
}

// Jumping with the key held must rise ~44 px and come back down onto the same floor.
func TestJumpHeight(t *testing.T) {
	l := loadLevel(1)
	g := &Game{health: maxMeter}
	p := newPlayer(l)
	p.x = 400 // flat street in level 1
	for i := 0; i < 5; i++ {
		p.update(g, l, Input{})
	}
	y0, top := p.y, p.y
	p.update(g, l, Input{Jump: true, JumpHeld: true})
	for i := 0; i < 80; i++ {
		p.update(g, l, Input{JumpHeld: i < 60})
		top = min(top, p.y)
	}
	if rise := y0 - top; rise < 36 || rise > 56 {
		t.Errorf("jump rise = %.0f px, want ~44", y0-top)
	}
	if p.y != y0 || !p.onGround {
		t.Errorf("did not land back on the floor: y=%.0f want %.0f", p.y, y0)
	}
}

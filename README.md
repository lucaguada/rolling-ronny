# Rolling Ronny (Go + Ebitengine)

Side-scrolling platformer built on the graphics/levels extracted from the DOS game *Rolling Ronny* (`assets/`).
Nine levels; tiles, attributes, sprites, HUD, SFX and music all come from the extracted assets (embedded in the binary).

    go1.27.4 run .          # or: go1.27.4 build -o rolling-ronny . && ./rolling-ronny

Keys: ←/→ or A/D move, Space/Z jump (hold = higher), ↑/↓ or W/S ladders, ↓ crouch, M music, Esc title, Enter confirm.
Collect coins/cubes/pickups, stomp enemies, avoid hazards (red/spiky tiles, bobbing hazards), reach the exit before the 5:00 clock runs out.

Debug env vars: `RR_LEVEL=n` start in level n, `RR_X=px` start x, `RR_SHOT=file.png` (+`RR_FRAMES`) screenshot and exit, `RR_BOT=1` auto-run right, `RR_TPS=n` simulate faster.
`go1.27.4 test .` checks that every level loads, the start is safe, and jump height.

## Web (WebAssembly)

    ./build-web.sh && python3 -m http.server -d web 8080     # then open http://localhost:8080

`web/` is a static site (index.html, wasm_exec.js, rolling-ronny.wasm ~18 MB): host it anywhere,
gzip shrinks it to ~5 MB. Click the page once if the browser blocks audio.

## Assets
The graphics/music are extracted from the original DOS game and are **not** part of this repository (`/assets` is git-ignored).
To build, place the extracted tree at `assets/assets/{levels,sprites,sfx,gfx,fonts,music}` (see `go:embed` in `assets.go`).
The playable web build lives on the `gh-pages` branch.

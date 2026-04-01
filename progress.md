# Game Engine Simulation — Progress & Technical Plan

## Architecture

```
mock-engine/          (separate binary, CLI arg for N users)
  main.go

api-server/           (Fiber HTTP server)
  main.go
  game-engine/        (package, imported by api-server)
    engine.go
```

### Data Flow
```
Mock Engine (N goroutines)
  |
  | HTTP POST /submit  {"user_id": "user_42", "answer": true}
  v
API Server (Fiber, :3000)
  |
  | channel send
  v
Game Engine (single consumer goroutine)
  -> declares winner (first correct answer)
  -> ignores subsequent correct answers
```

---

## Technical Plan

### Step 1: Mock User Engine (`mock-engine/main.go`)
- [ ] Accept N as CLI arg (e.g. `go run main.go 1000`)
- [ ] Spawn N goroutines, each:
  - Random user ID: `user_1`, `user_2`, ... `user_N`
  - Random correct flag: ~50% chance true
  - Random delay: 10–1000ms
  - HTTP POST to `http://localhost:3000/submit` with JSON payload
- [ ] Use `sync.WaitGroup` to wait for all goroutines to finish
- [ ] Print summary: total sent, done, correct/incorrect counts

### Step 2: Game Engine (`api-server/game-engine/engine.go`)
- [ ] Define `Submission` struct: `UserID string`, `Answer bool`
- [ ] Define `GameEngine` struct with a channel (`chan Submission`) and winner state
- [ ] `New()` — constructor, starts a consumer goroutine
- [ ] Consumer goroutine: reads from channel, if answer correct AND no winner yet -> declare winner, print to console
- [ ] `Submit(s Submission)` — sends submission into the channel (non-blocking if possible)
- [ ] Thread-safe winner declaration (sync.Once or mutex+bool)

### Step 3: API Server (`api-server/main.go`)
- [ ] Initialize `GameEngine` via `New()`
- [ ] POST `/submit` — parse JSON body into `Submission`, call `engine.Submit()`
- [ ] Return 200 OK with simple JSON response
- [ ] Start Fiber on `:3000`

### Step 4: Test & Verify
- [ ] Run api-server, then mock-engine with N=1000
- [ ] Verify: exactly one winner printed
- [ ] Verify: no race conditions (`go run -race`)

---

## Progress Log

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1 | Mock User Engine | DONE | compiles clean |
| 2 | API Server with /submit | IN PROGRESS | |
| 3 | Game Engine package | TODO | |
| 4 | Integration test (1000 users) | TODO | |
| 5 | Race condition check | TODO | |

---

## Bonus Points (later if time allows)
- [ ] Metrics: correct/incorrect answer counts
- [ ] Print time taken to find winner
- [ ] Channels for event-driven handling (already planned in base design)

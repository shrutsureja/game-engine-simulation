# Technical Implementation Plan
> This document is meant to be consumed by a coding agent. Follow it exactly.

---

## Project Structure

```
game-engine-simulation/
├── mock-engine/
│   ├── go.mod          (already exists, module: mock-engine, go 1.26.1)
│   └── main.go         (CREATE THIS)
└── api-server/
    ├── go.mod          (already exists, has fiber/v3 installed)
    ├── go.sum          (already exists)
    ├── main.go         (CREATE THIS)
    └── game-engine/
        └── engine.go   (CREATE THIS)
```

---

## Step 1: Mock User Engine — `mock-engine/main.go`

### Purpose
Standalone binary that simulates N users concurrently submitting answers to the API server.

### CLI Usage
```
go run main.go <N>
# Example: go run main.go 1000
```

### JSON Payload sent to server
```json
{"user_id": "user_42", "answer": true}
```

### Full Implementation Spec

```
package main

imports:
  - bytes
  - encoding/json
  - fmt
  - math/rand
  - net/http
  - os
  - strconv
  - sync
  - time
```

**`main()` function:**
1. Read `os.Args[1]`, parse as int N. If missing or invalid, print usage and exit.
2. Declare `var wg sync.WaitGroup`
3. Loop i from 1 to N (inclusive):
   - `wg.Add(1)`
   - Launch goroutine: `go sendResponse(i, &wg)`
4. `wg.Wait()`
5. Print: `"All %d users have submitted their responses.\n", N`

**`sendResponse(userID int, wg *sync.WaitGroup)` function:**
1. `defer wg.Done()`
2. Generate random delay: `delay := time.Duration(rand.Intn(991)+10) * time.Millisecond` (range: 10–1000ms)
3. `time.Sleep(delay)`
4. Generate random answer: `answer := rand.Intn(2) == 1` (50% true, 50% false)
5. Build payload struct (anonymous or named):
   ```go
   payload := struct {
       UserID string `json:"user_id"`
       Answer bool   `json:"answer"`
   }{
       UserID: fmt.Sprintf("user_%d", userID),
       Answer: answer,
   }
   ```
6. Marshal to JSON with `json.Marshal(payload)`
7. POST to `http://localhost:3000/submit` using `http.Post(url, "application/json", bytes.NewBuffer(body))`
8. If error, print error and return. If response received, close `resp.Body`.

### No external dependencies needed (stdlib only)

---

## Step 2: Game Engine Package — `api-server/game-engine/engine.go`

### Purpose
Package `gameengine` (note: Go package name cannot have hyphen, use `gameengine`).
Receives submissions via a channel, declares the first correct answer as winner.

### Types

```go
package gameengine

type Submission struct {
    UserID string
    Answer bool
}

type GameEngine struct {
    submissions chan Submission
    once        sync.Once
}
```

### Functions

**`New() *GameEngine`**
- Create engine: `e := &GameEngine{submissions: make(chan Submission, 1000)}`
  - Buffered channel of size 1000 to avoid blocking HTTP handlers under load
- Start consumer goroutine: `go e.run()`
- Return `e`

**`(e *GameEngine) Submit(s Submission)`**
- Non-blocking send: use select with default
  ```go
  select {
  case e.submissions <- s:
  default:
      // channel full, drop submission (game already over)
  }
  ```

**`(e *GameEngine) run()` (private goroutine)**
- Loop: `for s := range e.submissions`
  - If `s.Answer == true`:
    - Call `e.once.Do(func() { fmt.Printf("Winner: %s\n", s.UserID) })`
- This loop runs forever (goroutine lives as long as the server)

### Key design decisions
- `sync.Once` guarantees exactly one winner is printed, even if multiple correct answers arrive simultaneously
- Buffered channel (1000) absorbs burst traffic from mock engine
- `run()` is a single goroutine — no need for additional locking on winner state

---

## Step 3: API Server — `api-server/main.go`

### Purpose
Fiber v3 HTTP server. Receives POST `/submit`, forwards to game engine.

### Import path for game engine
```go
import "api-server/game-engine"
```
Package alias needed because of hyphen in directory name:
```go
import gameengine "api-server/game-engine"
```

### Implementation Spec

```
package main

imports:
  - github.com/gofiber/fiber/v3
  - api-server/game-engine  (aliased as gameengine)
  - log
```

**`main()` function:**
1. Create engine: `engine := gameengine.New()`
2. Create Fiber app: `app := fiber.New()`
3. Register route: `app.Post("/submit", submitHandler(engine))`
4. Start server: `log.Fatal(app.Listen(":3000"))`

**`submitHandler(engine *gameengine.GameEngine) fiber.Handler`**
Returns a closure:
```go
func submitHandler(engine *gameengine.GameEngine) fiber.Handler {
    return func(c fiber.Ctx) error {
        var s gameengine.Submission
        if err := c.Bind().Body(&s); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
        }
        engine.Submit(s)
        return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "received"})
    }
}
```

### JSON field mapping
`Submission` struct in `game-engine/engine.go` needs JSON tags for BodyParser:
```go
type Submission struct {
    UserID string `json:"user_id"`
    Answer bool   `json:"answer"`
}
```

---

## Build & Run Instructions

### Terminal 1 — Start API Server
```bash
cd api-server
go run main.go
```
Expected output: Fiber banner + listening on :3000

### Terminal 2 — Run Mock Engine
```bash
cd mock-engine
go run main.go 1000
```
Expected output in **api-server terminal**: `Winner: user_<N>`

### Race Condition Check
```bash
cd api-server
go run -race main.go
# In another terminal:
cd mock-engine
go run main.go 1000
```
Expected: no DATA RACE warnings.

---

## Constraints & Gotchas for the Coding Agent

1. **Package name**: directory is `game-engine` but Go package must be `gameengine` (no hyphens allowed in package names)
2. **Import path**: `"api-server/game-engine"` — hyphen in path is fine, hyphen in package name is not
3. **Fiber v3 API**: use `c.Bind().Body(&s)` NOT `c.BodyParser(&s)` — BodyParser was removed in v3. Context type is `fiber.Ctx` (not `*fiber.Ctx`)
4. **Do NOT** close the `submissions` channel — the server runs indefinitely
5. **Do NOT** use `sync.Mutex` for winner state — `sync.Once` is cleaner and correct
6. **rand** — no seed needed in Go 1.20+, global rand is auto-seeded

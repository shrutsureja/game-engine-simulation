Name : shrut sureja

Email: shrutsureja.work@gmail.com

## Project Overview

A Go-based backend system that simulates multiple users answering a game question, evaluates responses in real-time, and announces a winner. The project is split into **three separate components** as required:

## Project Structure

```
game-engine-simulation/
├── api-server/              # Component 2: API Server
│   ├── main.go
│   ├── game-engine/         # Component 3: Game Engine (separate package)
│   │   └── engine.go
│   ├── go.mod
│   └── go.sum
├── mock-engine/             # Component 1: Mock User Engine
│   ├── main.go
│   └── go.mod
├── task.md
└── README.md
```

## Component 1: Mock User Engine (`mock-engine/`)

A standalone Go module that simulates N concurrent users.

- Accepts an input N (number of users to simulate).
- For each user, randomly assigns a correct/incorrect answer and adds a random delay (10–1000ms) to simulate network lag.
- Sends all responses **concurrently** using goroutines to the API server.
- Uses `sync.WaitGroup` for goroutine coordination and `atomic.Int64` for thread-safe counting.

## Component 2: API Server (`api-server/`)

A Fiber-based HTTP server that receives user submissions and forwards them to the Game Engine.

| Endpoint       | Method | Description                          |
|----------------|--------|--------------------------------------|
| `/submit`      | POST   | Receive a user response (`user_id`, `answer`) and forward to Game Engine |
| `/start`       | GET    | Start the game engine timer          |
| `/reset`       | GET    | Reset the game engine state          |

## Component 3: Game Engine (`api-server/game-engine/`)

A **separate Go package** that contains the core game logic. It is imported by the API server as `api-server/game-engine`.

- Runs a **single goroutine event loop** using channels (not mutexes) for event-driven response handling.
- Determines the **first user** who sent a correct answer and declares them the winner.
- Ignores all subsequent correct responses once a winner is found.
- Tracks correct/incorrect answer counts in real-time.
- Records the time taken to find the winner.
- After 3 seconds of inactivity, prints the final game results (winner, response time, score breakdown).

Key types:
- `GameEngine` — The main struct; created via `gameengine.New()`.
- `Submission` — Payload struct with `UserID` (string) and `Answer` (bool).

## Prerequisites

- Go 1.26.1 or later

## How to Run

### Step 1: Start the API Server

```bash
cd api-server
go run main.go
```

The server starts on `http://localhost:3000`.

### Step 2: Run the Mock User Engine (in a separate terminal)

```bash
cd mock-engine
go run main.go <N>
```

Replace `<N>` with the number of simulated users. For example:

```bash
go run main.go 1000
```

This will:
1. Reset the game engine
2. Start the game timer
3. Spawn 1000 goroutines, each submitting a random answer with a random delay (10–1000ms)
4. Print a summary of correct vs incorrect submissions

### Step 3: Observe Results

The **API server terminal** will display:
- A running tally of submissions as they arrive
- After 3 seconds of inactivity, the final game results including the winner and their response time

## Concurrency Approach

- **Mock Engine**: Uses goroutines + `sync.WaitGroup` for concurrent HTTP requests, `atomic.Int64` for thread-safe counters.
- **Game Engine**: Uses **channels** for event-driven handling (submissions, start, reset) — no mutexes needed. A single goroutine processes all events sequentially via `select`, ensuring only one winner is declared without race conditions.

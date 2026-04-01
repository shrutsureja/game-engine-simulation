package gameengine

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Submission struct {
	UserID string `json:"user_id"`
	Answer bool   `json:"answer"`
}

type GameEngine struct {
	submissions    chan Submission
	resets         chan struct{}
	starts         chan struct{}
	startTime      time.Time
	started        atomic.Bool
	winnerFound    bool
	winnerTime     atomic.Int64 // nanoseconds from start to winner
	correctCount   int
	incorrectCount int
}

func New() *GameEngine {
	e := &GameEngine{
		submissions: make(chan Submission, 100),
		resets:      make(chan struct{}),
		starts:      make(chan struct{}),
	}
	go e.run()
	return e
}

func (e *GameEngine) Submit(s Submission) {
	e.submissions <- s
}

func (e *GameEngine) Start() {
	e.starts <- struct{}{}
}

func (e *GameEngine) Reset() {
	e.resets <- struct{}{}
}

func (e *GameEngine) run() {
	idleTimer := time.NewTimer(0)
	idleTimer.Stop()
	winner := ""

	for {
		select {
		case <-e.starts:
			e.startTime = time.Now()
			e.started.Store(true)
			fmt.Println("Game engine started.")

		case s := <-e.submissions:

			if s.Answer {
				e.correctCount++
				if !e.winnerFound {
					e.winnerFound = true
					winner = s.UserID
					e.winnerTime.Store(int64(time.Since(e.startTime)))
				}
			} else {
				e.incorrectCount++
			}
			fmt.Printf("Correct: %d | Incorrect: %d | Total: %d | user: %s\n", e.correctCount, e.incorrectCount, e.correctCount+e.incorrectCount, s.UserID)

			// Reset idle timer on every submission
			idleTimer.Stop()
			idleTimer.Reset(3 * time.Second)

		case <-idleTimer.C:
			// 3 seconds since last submission — game is done
			fmt.Println("\n=================game over========================")
			if winner != "" {
				fmt.Printf("Winner: %s\n", winner)
				fmt.Printf("Winner time: %s\n", time.Duration(e.winnerTime.Load()))
			} else {
				fmt.Println("No winner — no correct answers received.")
			}
			fmt.Printf("Correct: %d | Incorrect: %d | Total: %d\n", e.correctCount, e.incorrectCount, e.correctCount+e.incorrectCount)
			fmt.Println("==================================================")
		case <-e.resets:
			idleTimer.Stop()
			e.correctCount = 0
			e.incorrectCount = 0
			e.winnerFound = false
			e.winnerTime.Store(0)
			e.started.Store(false)
			winner = ""
			fmt.Println("Game engine reset.")
		}
	}
}

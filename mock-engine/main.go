package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var correctCount atomic.Int64
var incorrectCount atomic.Int64

type payload struct {
	UserID string `json:"user_id"`
	Answer bool   `json:"answer"`
}

func sendResponse(userID int, wg *sync.WaitGroup) {
	defer wg.Done()

	answer := rand.Intn(2) == 1
	delay := time.Duration(rand.Intn(991)+10) * time.Millisecond
	time.Sleep(delay)

	p := payload{
		UserID: fmt.Sprintf("user_%d", userID),
		Answer: answer,
	}

	body, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("user_%d: failed to marshal payload: %v\n", userID, err)
		return
	}

	resp, err := http.Post("http://localhost:3000/submit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("user_%d: request failed: %v\n", userID, err)
		return
	}
	resp.Body.Close()

	if answer {
		correctCount.Add(1)
	} else {
		incorrectCount.Add(1)
	}
	fmt.Printf("user_id:%s | answer:%t | delay:%s\n", p.UserID, p.Answer, delay)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <N>")
		fmt.Println("Example: go run main.go 1000")
		os.Exit(1)
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n <= 0 {
		fmt.Println("Error: N must be a positive integer")
		os.Exit(1)
	}

	// Reset game engine
	resp, err := http.Get("http://localhost:3000/reset")
	if err != nil {
		fmt.Println("Warning: could not reset game engine:", err)
	} else {
		resp.Body.Close()
	}

	// Start game engine timer
	resp, err = http.Get("http://localhost:3000/start")
	if err != nil {
		fmt.Println("Error: could not start game engine:", err)
		os.Exit(1)
	}
	resp.Body.Close()

	fmt.Printf("Starting %d users...\n", n)

	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go sendResponse(i, &wg)
	}

	wg.Wait()
	fmt.Printf("All %d users have submitted their responses.\n", n)
	fmt.Printf("Correct: %d | Incorrect: %d\n", correctCount.Load(), incorrectCount.Load())
}

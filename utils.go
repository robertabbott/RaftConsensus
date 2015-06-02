package raft

import (
	"math/rand"
	"strings"
	"time"
)

const TimeoutMax = 800
const TimeoutMin = 300

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func startTimeout(toChan chan int) {
	to := time.Duration(random(TimeoutMin, TimeoutMax))
	time.Sleep(to * time.Millisecond)
	toChan <- int(TIMEOUT_ELAPSED)
}

func getPort(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[len(parts)-1]
}

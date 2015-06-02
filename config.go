package raft

import (
	"math/rand"
	"time"
)

const TimeoutMax = 800
const TimeoutMin = 300

type RaftConfig struct {
	Addr    string   // this node's address
	members []string // addresses of all cluster members
	timeout time.Duration
}

func DefaultConfig() *RaftConfig {
	c := &RaftConfig{
		Addr:    "localhost:6868",
		timeout: TimeoutMax * time.Millisecond,
	}
	return c
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

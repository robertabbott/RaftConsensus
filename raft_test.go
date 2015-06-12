package raft

import (
	"fmt"
	"testing"
	"time"
)

func ThreeNodeRaft() (*RaftNode, *RaftNode, *RaftNode) {
	members := []string{"localhost:6868", "localhost:6969", "localhost:7070"}
	conf1 := CreateConfig("localhost:6868", members[1:])
	conf3 := CreateConfig("localhost:6969", append([]string{members[0]}, []string{members[2]}...))
	conf2 := CreateConfig("localhost:7070", members[:2])
	r1 := initRaft(conf1)
	r2 := initRaft(conf2)
	r3 := initRaft(conf3)
	return r1, r2, r3
}

func TestRun1Raft(t *testing.T) {
	conf := DefaultConfig()
	r := initRaft(conf)
	go r.RunRaft()
	time.Sleep(4 * time.Second)
	if r.State != LEADER {
		t.Fatalf("expected node to be leader. Was: %d", r.State)
	}
}

// given 3 nodes leader should be elected
// leader should send heartbeats indefinitely
func TestRun3Raft(t *testing.T) {
	r1, r2, r3 := ThreeNodeRaft()
	go r1.RunRaft()
	go r2.RunRaft()
	go r3.RunRaft()

	// leader not being elected
	// some issues with networking stuff

	time.Sleep(10 * time.Second)
	fmt.Println(r1.State)
	fmt.Println(r2.State)
	fmt.Println(r3.State)
	r1.Shutdown()
	r2.Shutdown()
	r3.Shutdown()
}

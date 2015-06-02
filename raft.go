package raft

import (
	"sync"
)

type State int

const (
	FOLLOWER State = 1 + iota
	CANDIDATE
	LEADER
	SHUTDOWN
)

type RaftNode struct {
	State  State
	config *RaftConfig

	currentTerm int
	votedFor    string
	commitIndex int
	lastApplied int

	leaderLock sync.RWMutex // guards leader
	leader     string

	leaderState *LeaderState // only used when this node is leader
}

type LeaderState struct {
	commitCh                chan struct{}
	replicationState        map[string]int // member addr -> index known to be replicated
	doWhatNoDictatorEverHas chan struct{}  // relinquish power
}

// initialize RaftNode struct
func initRaft(config *RaftConfig) *RaftNode {
	r := &RaftNode{
		State:  FOLLOWER,
		config: config,
	}
	return r
}

func (r *RaftNode) RunRaft() error {
	for {
		if r.State == FOLLOWER {
			r.runFollower()
		} else if r.State == CANDIDATE {
			r.runCandidate()
		} else if r.State == LEADER {
			r.runLeader()
		} else {
			r.shutdown()
			return nil
		}
	}
}

// runs as follower until node changes state
func (r *RaftNode) runFollower() {

}

// runs as candidate until node changes state
func (r *RaftNode) runCandidate() {

}

// runs as leader until node changes state
func (r *RaftNode) runLeader() {

}

func (r *RaftNode) shutdown() {

}

func main() {
	// until I figure out a better way to get the config:
	config := DefaultConfig()
	// Initialize node
	node := initRaft(config)
	// run raft forever
	node.RunRaft()
}

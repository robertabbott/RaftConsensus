package raft

import (
	"net"
	"sync"
)

type State int

const (
	FOLLOWER State = 1 + iota
	CANDIDATE
	LEADER
	SHUTDOWN
	TIMEOUT_ELAPSED
	RECEIVED_RPC
)

type RaftNode struct {
	State  State
	config *RaftConfig

	currentTerm int
	votedFor    string
	commitIndex int
	lastApplied int

	rpcCh chan interface{}
	conn  net.Conn

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
	// setup networking to receive RPCs

	// we need to be able to operate on the RaftNode based on the RPC we receive.
	// To check information in the struct to determine our response and to change
	// information in the struct based on the message we are receiving.
	//
	// listenForRPCs will connect to a port and receive RPC messages from the leader
	// or from candidates
	//
	// The server will run RaftRPC methods. The question is how do we give the RaftRPC
	// methods access to the RaftNode struct

	///////////////
	// This needs to be rethought since ServeConn blocks and fucks everything up...
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
	ch := make(chan int, 1)
	go startTimeout(ch)
	for {
	}
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

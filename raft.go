package raft

import (
	"fmt"
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

	logDataLock sync.RWMutex // guards log metadata
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
	tochan := make(chan int, 1)
	go startTimeout(tochan)
	for {
		go r.RunTCPServer()
		for {
			select {
			case to := <-tochan:
				if to == int(TIMEOUT_ELAPSED) {
					r.State = CANDIDATE
				}
				return
			case rpc := <-r.rpcCh:
				switch rpc.(type) {
				case *AppendEntries:
					r.appendEntries(rpc.(*AppendEntries))
				case *AppendEntriesResp:
					r.appendEntriesResp(rpc.(*AppendEntriesResp))
				case *RequestVote:
					r.requestVote(rpc.(*RequestVote))
				case *RequestVoteResp:
					r.requestVoteResp(rpc.(*RequestVoteResp))
				default:
					fmt.Println("fuck")
					return
				}
				go startTimeout(tochan)
			}
		}
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

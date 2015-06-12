package raft

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
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
	commitTerm  int // term in highest committed log entry
	lastApplied int

	tochan chan int
	rpcCh  chan interface{}
	conn   net.Conn

	leaderLock sync.RWMutex // guards leader
	leader     string

	leaderState *LeaderState // only used when this node is leader

	logger *log.Logger
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
		logger: log.New(os.Stdout, "", log.LstdFlags),
		tochan: make(chan int, 1),
		rpcCh:  make(chan interface{}, 1),
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
	receivedRPC := false
	go r.RunTCPServer()
	go startTimeout(r.tochan)
	r.logger.Printf("started TCP server")
	for {
		select {
		case to := <-r.tochan:
			if receivedRPC == true { // if rpc has been received this to is ignored
				receivedRPC = false
				continue
			} else if to == int(TIMEOUT_ELAPSED) {
				r.State = CANDIDATE
				r.currentTerm += 1
			}
			r.logger.Printf("[INFO] timed out. Switching to candidate")
			return
		case rpc := <-r.rpcCh:
			r.logger.Printf("[INFO] received rpc")
			receivedRPC = true
			switch rpc.(type) {
			case *AppendEntries:
				r.appendEntries(rpc.(*AppendEntries))
			case *AppendEntriesResp:
				r.appendEntriesResp(rpc.(*AppendEntriesResp))
			case *RequestVote:
				r.HandleVoteRequest(rpc.(*RequestVote))
			default:
				fmt.Println("fuck")
				return
			}
			go startTimeout(r.tochan)
		}
	}
}

// runs as candidate until node changes state
func (r *RaftNode) runCandidate() {
	voteCount := 1 // node votes for itself
	electionTimeout := 0
	r.votedFor = r.config.addr
	majority := len(r.config.members) / 2

	for _, member := range r.config.members {
		go r.SendRequestVote(member)
	}
	// read from rpcCh until majority is achieved
	// if gets AppendEntries or VoteReq with > term revert to follower
	// else keep counting votes
	for r.State == CANDIDATE {
		go startTimeout(r.tochan)
		select {
		case rpc := <-r.rpcCh:
			switch rpc.(type) {
			case *AppendEntries:
				r.appendEntries(rpc.(*AppendEntries))
			case *AppendEntriesResp:
				r.appendEntriesResp(rpc.(*AppendEntriesResp))
			case *RequestVote:
				r.HandleVoteRequest(rpc.(*RequestVote))
			case *RequestVoteResp:
				r.HandleVoteReqResp(rpc.(*RequestVoteResp), &voteCount)
			default:
				fmt.Println("fuck")
				return
			}
		case _ = <-r.tochan:
			electionTimeout += 1
			go startTimeout(r.tochan)
		}
		if voteCount > majority {
			r.State = LEADER
			return
		}
		if electionTimeout > 2 {
			r.currentTerm += 1
			electionTimeout = 0
		}
	}
}

// runs as leader until node changes state
func (r *RaftNode) runLeader() {
	r.logger.Printf("taking up the mantle of leadership")
	time.Sleep(5 * time.Second)
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

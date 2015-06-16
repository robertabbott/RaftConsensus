package raft

import (
	"errors"
	"log"
	"net"
	"os"
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
	commitTerm  int // term in highest committed log entry
	lastApplied int

	tochan     chan int
	shutdownCh chan bool
	rpcCh      chan *RaftRPC    // chan for RPCs from raft members
	reqCh      chan interface{} // chan for client requests
	conn       net.Conn

	leaderLock  sync.RWMutex // guards leader
	leader      string
	leaderState *LeaderState // only used when this node is leader

	Log []*LogEntry

	logger *log.Logger
}

type LeaderState struct {
	commitCh        chan interface{}
	replicatedIndex map[string]int // member addr -> index known to be replicated
}

// initialize RaftNode struct
func initRaft(config *RaftConfig) *RaftNode {
	ls := newLeaderState(config.members)
	r := &RaftNode{
		State:       FOLLOWER,
		config:      config,
		logger:      log.New(os.Stdout, "", log.LstdFlags),
		tochan:      make(chan int, 1),
		rpcCh:       make(chan *RaftRPC, 1),
		leaderState: ls,
		Log:         []*LogEntry{},
	}
	return r
}

func newLeaderState(members []string) *LeaderState {
	ls := &LeaderState{
		commitCh:                make(chan interface{}, 1),
		doWhatNoDictatorEverHas: make(chan interface{}, 1),
		replicatedIndex:         make(map[string]int),
	}
	for _, member := range members {
		ls.replicatedIndex[member] = 0
	}
	return ls
}

func (r *RaftNode) RunRaft() error {
	go r.RunTCPServer()
	for {
		if r.State == FOLLOWER {
			r.runFollower()
		} else if r.State == CANDIDATE {
			r.runCandidate()
		} else if r.State == LEADER {
			r.runLeader()
		} else {
			return errors.New("Raft put in invalid state")
		}
	}
}

// runs as follower until node changes state
func (r *RaftNode) runFollower() {
	receivedRPC := false
	go startTimeout(r.tochan)
	r.logger.Printf("started TCP server")
	for {
		select {
		case to := <-r.tochan:
			if receivedRPC == true { // if rpc has been received this to is ignored
				receivedRPC = false
				continue
			} else if to == int(TIMEOUT_ELAPSED) {
				log.Printf("[INFO] switching to candidate")
				r.State = CANDIDATE
				r.currentTerm += 1
			}
			return
		case rpc := <-r.rpcCh:
			r.logger.Printf("[INFO] received rpc")
			receivedRPC = true
			switch rpc.St.(type) {
			case AppendEntries:
				r.HandleAppendEntries(rpc.St.(AppendEntries))
			case AppendEntriesResp:
				r.HandleAppendEntriesResp(rpc.St.(AppendEntriesResp))
			case RequestVote:
				r.HandleVoteRequest(rpc.St.(RequestVote))
			default:
				return
			}
		case sd := <-r.shutdownCh:
			if sd == true {
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
			switch rpc.St.(type) {
			case AppendEntries:
				r.HandleAppendEntries(rpc.St.(AppendEntries))
			case AppendEntriesResp:
				r.HandleAppendEntriesResp(rpc.St.(AppendEntriesResp))
			case RequestVote:
				r.HandleVoteRequest(rpc.St.(RequestVote))
			case RequestVoteResp:
				r.HandleVoteReqResp(rpc.St.(RequestVoteResp), &voteCount)
			default:
				r.logger.Printf("%+v\n", rpc)
			}
		case _ = <-r.tochan:
			r.logger.Printf("[INFO] %s received %d votes", r.config.addr, voteCount)
			electionTimeout += 1
			go startTimeout(r.tochan)
		case sd := <-r.shutdownCh:
			if sd == true {
				return
			}
		}
		if voteCount > majority {
			r.logger.Printf("[LEADER] taking power %s", r.config.addr)
			r.State = LEADER
			return
		}
		if electionTimeout > 2 {
			r.currentTerm += 1
			electionTimeout = 0
			voteCount = 1
			r.logger.Printf("[INFO] term is %d", r.currentTerm)
			for _, member := range r.config.members {
				go r.SendRequestVote(member)
			}
		}
	}
}

// runs as leader until node changes state
func (r *RaftNode) runLeader() {
	r.logger.Printf("[INFO] %s is now leader", r.config.addr)
	//go r.ServeClientTCP()
	r.NoOp() // commit noop entry
	for r.State == LEADER {
		for _, member := range r.config.members {
			go r.SendHeartbeat(member)
		}
		select {
		case rpc := <-r.rpcCh:
			switch rpc.St.(type) {
			case AppendEntries:
				r.HandleAppendEntries(rpc.St.(AppendEntries))
			case AppendEntriesResp:
				r.HandleAppendEntriesResp(rpc.St.(AppendEntriesResp))
			case RequestVote:
				r.HandleVoteRequest(rpc.St.(RequestVote))
			}
		case sd := <-r.shutdownCh:
			if sd == true {
				return
			}
		default:
			continue
		}
	}
}

func (r *RaftNode) NoOp() {
	noop := &LogEntry{
		Index: r.commitIndex,
		Term:  r.commitTerm,
		Type:  NOOP,
	}
	r.Log = append(r.Log, noop)
	r.commitIndex += 1
	r.commitTerm += 1
}

func main() {
	// until I figure out a better way to get the config:
	config := DefaultConfig()
	// Initialize node
	node := initRaft(config)
	// run raft forever
	node.RunRaft()
}

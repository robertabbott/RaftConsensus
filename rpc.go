package raft

import (
	"sync"
)

// defines rpc messages that can be sent

type RaftRPC struct {
	msg    interface{}
	respCh chan RaftRPCResp

	mu sync.RWMutex // guards raftNode
	rn *RaftNode
}

type RaftRPCResp struct {
	msg interface{}
	err error
}

type RequestVote struct {
	candidateTerm int
	candidateId   string // addr of candidate

	lastLogIndex int // index of candidates highest commit
	lastLogTerm  int // term of candidates highest log entry
}

type RequestVoteResp struct {
	voterId string // addr of voter

	// follower votes for a candidate if candidate is more up to date. voteGranted is
	// set to false if the followers term > candidateTerm
	voteGranted bool
}

type AppendEntries struct {
	leaderTerm int
	addr       string

	leaderCommit int // highest index committed by the leader
	prevLogIndex int // log index preceding the index of the first new entry
	prevLogTerm  int // term that corresponds to prevLogIndex

	newEntries []*LogEntry
}

type AppendEntriesResp struct {
	success        bool // set to false if follower term > leaderTerm
	term           int  // follower term so leader can track
	followerCommit int  // highest index committed by follower
}

const (
	AppendEntriesRPC uint8 = iota
	RequestVoteRPC
)

// methods for sending/receiving rpc messages

// Server exports RaftNode object and exposes HandleVoteReq to client
func (r *RaftNode) requestVote(req *RequestVote) error {

	return nil
}

func (r *RaftNode) requestVoteResp(req *RequestVoteResp) error {

	return nil
}

func (r *RaftNode) appendEntries(req *AppendEntries) error {

	return nil
}

func (r *RaftNode) appendEntriesResp(req *AppendEntriesResp) error {

	return nil
}

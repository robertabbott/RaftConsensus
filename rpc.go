package raft

import (
	"sync"
)

const (
	AppendEntriesRPC uint8 = iota
	RequestVoteRPC
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

type RequestVote struct {
	CandidateTerm int
	CandidateId   string // addr of candidate

	LastLogIndex int // index of candidates highest commit
	LastLogTerm  int // term of candidates highest log entry
}

type RequestVoteResp struct {
	VoterId string // addr of voter

	// follower votes for a candidate if candidate is more up to date. VoteGranted is
	// set to false if the followers term > CandidateTerm
	VoteGranted bool
}

// methods for sending/receiving rpc messages

// Server exports RaftNode object and exposes HandleVoteReq to client
func (r *RaftNode) HandleVoteRequest(req *RequestVote) error {
	resp := RequestVoteResp{}
	if r.State == FOLLOWER {
		if r.currentTerm < req.CandidateTerm {
			resp.VoteGranted = true
		}
	} else if r.State == CANDIDATE || r.State == LEADER {
		if r.currentTerm < req.CandidateTerm {
			r.State = FOLLOWER
			resp.VoteGranted = true
		} else {
			resp.VoteGranted = false
		}
	}
	go SendStructTCP(req.CandidateId, resp)
	return nil
}

func (r *RaftNode) HandleVoteReqResp(req *RequestVoteResp, count *int) error {
	if r.State == CANDIDATE && req.VoteGranted == true {
		*count += 1
	}
	return nil
}

func (r *RaftNode) SendRequestVote(addr string) error {
	rv := RequestVote{
		CandidateTerm: r.currentTerm,
		CandidateId:   r.config.addr,

		LastLogIndex: r.commitIndex,
		LastLogTerm:  r.commitTerm,
	}
	go SendStructTCP(addr, rv)
	return nil
}

func (r *RaftNode) appendEntries(req *AppendEntries) error {

	return nil
}

func (r *RaftNode) appendEntriesResp(req *AppendEntriesResp) error {

	return nil
}

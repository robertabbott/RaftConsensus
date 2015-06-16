package raft

import ()

const (
	AppendEntriesRPC uint8 = iota
	RequestVoteRPC
)

// defines rpc messages that can be sent

type RaftRPC struct {
	St interface{}
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

type AppendEntries struct {
	LeaderTerm int
	Addr       string

	LeaderCommit int // highest index committed by the leader
	PrevLogIndex int // log index preceding the index of the first new entry
	PrevLogTerm  int // term that corresponds to prevLogIndex

	NewEntries []*LogEntry
}

type AppendEntriesResp struct {
	Success        bool // set to false if follower term > leaderTerm
	Term           int  // follower term so leader can track
	FollowerCommit int  // highest index committed by follower
	Addr           string
}

// methods for handling and sending RPC messages

func (r *RaftNode) HandleAppendEntriesResp(req AppendEntriesResp) error {
	if req.Success == false {
		r.State = FOLLOWER
	}
	r.leaderState.replicatedIndex[req.Addr] = req.FollowerCommit
	return nil
}

func (r *RaftNode) HandleAppendEntries(req AppendEntries) error {
	if r.State == CANDIDATE || r.State == LEADER {
		if r.currentTerm <= req.LeaderTerm {
			r.State = FOLLOWER
		}
	}
	resp := AppendEntriesResp{
		Addr: r.config.addr,
	}
	if r.currentTerm > req.LeaderTerm {
		resp.Success = false
		go SendStructTCP(req.Addr, resp)
		return nil
	} else {
		resp.Success = true
	}

	// TODO these entries might not be properly committed...
	// update log and send response
	r.Log = append(r.Log[:req.PrevLogIndex], req.NewEntries...)
	r.currentTerm = req.LeaderTerm
	r.commitIndex = len(r.Log) - 1

	resp.Term = r.currentTerm
	resp.FollowerCommit = r.commitIndex
	go SendStructTCP(req.Addr, resp)

	// update leader if necessary
	r.leaderLock.RLock()
	defer r.leaderLock.RUnlock()
	if r.leader != req.Addr {
		r.leaderLock.Lock()
		defer r.leaderLock.Unlock()
		r.leader = req.Addr
	}
	return nil
}

func (r *RaftNode) SendHeartbeat(addr string) error {
	prevLogIndex := r.leaderState.replicatedIndex[addr]
	prevLogTerm := r.Log[prevLogIndex].Term
	ae := AppendEntries{
		// leader metadata
		LeaderTerm:   r.currentTerm,
		Addr:         r.config.addr,
		LeaderCommit: r.commitIndex,

		// log info
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		NewEntries:   r.Log[prevLogIndex:],
	}
	go SendStructTCP(addr, ae)
	return nil
}

// Server exports RaftNode object and exposes HandleVoteReq to client
func (r *RaftNode) HandleVoteRequest(req RequestVote) error {
	resp := RequestVoteResp{}
	if r.State == FOLLOWER {
		if r.currentTerm < req.CandidateTerm {
			resp.VoteGranted = true
			r.votedFor = req.CandidateId
		}
	} else if r.State == CANDIDATE || r.State == LEADER {
		if r.currentTerm < req.CandidateTerm {
			r.State = FOLLOWER
			resp.VoteGranted = true
			r.votedFor = req.CandidateId
		}
	}
	if resp.VoteGranted == true {
		r.logger.Printf("[INFO] voting true")
	}
	go SendStructTCP(req.CandidateId, resp)
	return nil
}

func (r *RaftNode) HandleVoteReqResp(req RequestVoteResp, count *int) error {
	r.logger.Printf("[CANDIDATE] got resp: %b", req.VoteGranted)
	if r.State == CANDIDATE && req.VoteGranted == true {
		r.logger.Printf("[INFO] RECEIVED A VOTE!!!")
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

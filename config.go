package raft

import ()

type RaftConfig struct {
	addr    string   // this node's address
	members []string // addresses of all cluster members
}

func DefaultConfig() *RaftConfig {
	c := &RaftConfig{
		addr: "localhost:6767",
	}
	return c
}

func CreateConfig(addr string, members []string) *RaftConfig {
	c := &RaftConfig{
		addr:    addr,
		members: members,
	}
	return c
}

package raft

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

type TestRPC struct {
	S string
}

func HandleConnection(conn net.Conn) {
	dec := gob.NewDecoder(conn)
	p := &TestRPC{}
	dec.Decode(p)
	fmt.Println(p.S)
	// return struct type
}

func (r *RaftNode) RunTCPServer(port string) {
	fmt.Println("start")
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept() // this blocks until connection or error
		if err != nil {
			log.Fatal(err)
		}
		go HandleConnection(conn) // a goroutine handles conn so that the loop can accept other connections
		rpc := <-r.rpcCh
		switch rpc.(type) {
		case *AppendEntries:
			r.appendEntries(rpc.(*AppendEntries))
		case *AppendEntriesResp:
			r.appendEntriesResp(rpc.(*AppendEntriesResp))
		case *RequestVote:
			r.requestVote(rpc.(*RequestVote))
		case *RequestVoteResp:
			r.requestVoteResp(rpc.(*RequestVoteResp))
		}
	}
}

func ConnectTCP(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func sendStruct(st interface{}, conn net.Conn) {
	enc := gob.NewEncoder(conn)
	err := enc.Encode(st)
	if err != nil {
		log.Fatal(err)
	}
}

func closeConn(conn net.Conn) {
	conn.Close()
}

func SendStructTCP(addr, msg string) {
	conn := ConnectTCP(addr)
	T := TestRPC{S: msg}
	sendStruct(T, conn)
	closeConn(conn)

}

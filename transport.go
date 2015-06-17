package raft

import (
	"encoding/gob"
	"log"
	"net"
)

func HandleConnection(conn net.Conn, ch chan *RaftRPC) {
	dec := gob.NewDecoder(conn)
	p := &RaftRPC{}
	dec.Decode(p)
	ch <- p // put struct in ch
}

func (r *RaftNode) RunTCPServer() {
	ln, err := net.Listen("tcp", ":"+getPort(r.config.addr))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept() // this blocks until connection or error
		if err != nil {
			log.Fatal(err)
		}
		go HandleConnection(conn, r.rpcCh) // a goroutine handles conn so that the loop can accept other connections
	}
}

func SendStructTCP(addr string, st interface{}) {
	conn := ConnectTCP(addr)
	rrpc := &RaftRPC{
		St: st,
	}
	sendStruct(rrpc, conn)
	conn.Close()
}

func ConnectTCP(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func sendStruct(st *RaftRPC, conn net.Conn) {
	gob.Register(st.St)
	enc := gob.NewEncoder(conn)
	err := enc.Encode(st)
	if err != nil {
		log.Fatal(err)
	}
}

func ShutdownServer(conn *net.TCPListener) {
	conn.Close()
}

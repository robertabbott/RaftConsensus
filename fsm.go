package raft

import ()

// LogType describes various types of log entries.
type LogType uint8

const (
	LOGCOMMAND LogType = iota
	NOOP
	ADDSERVER
	REMOVESERVER
)

type LogEntry struct {
	Index int
	Term  int

	Type LogType
	Data []byte
}

package database

import (
	"goRedis/interface/resp"
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

// Database is the interface for redis style storage engine
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection)
	Close()
}

// DataEntity stores data bound to a key, including a string, list, hash, set and so on
type DataEntity struct {
	Data interface{}
}

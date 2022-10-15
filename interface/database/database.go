package database

import (
	"goRedis/interface/resp"
)

// CmdLine 是[][]字节的别名，表示命令行
type CmdLine = [][]byte

// Database redis风格存储接口
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection)
	Close()
}

// DataEntity 存储绑定到键的数据，包括字符串、列表、哈希、集等
type DataEntity struct {
	Data interface{}
}

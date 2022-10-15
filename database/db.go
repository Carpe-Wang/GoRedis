// Package database is a memory database with redis compatible interface
package database

import (
	"goRedis/datastruct/dict"
	"goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/resp/reply"
	"strings"
)

// DB stores data and execute user's commands
type DB struct {
	index  int // 数据编号
	data   dict.Dict
	addAof func(CmdLine)
}

// ExecFunc command执行器的接口
// 参数不包括cmdline
type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

// makeDB 创建DB实例
func makeDB() *DB {
	db := &DB{
		data:   dict.MakeSyncDict(),
		addAof: func(line CmdLine) {},
	}
	return db
}

// Exec 在一个DB中执行命令
func (db *DB) Exec(c resp.Connection, cmdLine [][]byte) resp.Reply {

	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "'")
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.executor
	return fun(db, cmdLine[1:])
}

func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

/* ---- 连接数据库 ----- */

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {

	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove 指定的key清除
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// Removes 根据key，清除数据库
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Flush() {
	db.data.Clear()
}

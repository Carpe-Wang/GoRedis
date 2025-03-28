package database

import (
	"goRedis/datastruct/sortedset"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/lib/wildcard"
	"goRedis/resp/reply"
	"strconv"
	"time"
)

// execDel removes a key from db
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

// execExists checks if a is existed in db
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// execFlushDB removes all data in current db
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLine2("flushdb", args...))
	return &reply.OkReply{}
}

// execType returns the type of entity, including: string, list, hash, set and zset
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	case *sortedset.SortedSet:
		return reply.MakeStatusReply("zset")
	}
	return &reply.UnknownErrReply{}
}

// execRename a key
func execRename(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[0])
	dest := string(args[1]) //需要修改为的名字

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("rename", args...))
	return &reply.OkReply{}
}

// execRenameNx a key, only if the new key does not exist
func execRenameNx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.Removes(src, dest) // clean src and dest with their ttl
	db.PutEntity(dest, entity)
	db.addAof(utils.ToCmdLine2("renamenx", args...))
	return reply.MakeIntReply(1)
}

// execKeys returns all keys matching the given pattern
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

// execExpire sets expiration time for the given key
func execExpire(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Parse seconds
	seconds, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	if seconds <= 0 {
		// Remove key if given a non-positive expiration time
		db.Remove(key)
		return reply.MakeIntReply(1)
	}

	// Get entity
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	// Calculate expiration time
	expireTime := time.Now().UnixNano()/1e6 + seconds*1000 // Convert to milliseconds
	entity.ExpireTime = expireTime

	// Store the key back
	db.PutEntity(key, entity)

	db.addAof(utils.ToCmdLine2("expire", args...))
	return reply.MakeIntReply(1)
}

// execTTL returns the remaining time to live of a key
func execTTL(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get entity
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeIntReply(-2) // Key does not exist
	}

	// If no expiration
	if entity.ExpireTime == 0 {
		return reply.MakeIntReply(-1) // No expiration
	}

	// Calculate remaining time
	now := time.Now().UnixNano() / 1e6 // current time in milliseconds
	remaining := entity.ExpireTime - now

	if remaining <= 0 {
		// Key has expired
		db.Remove(key)
		return reply.MakeIntReply(-2)
	}

	// Return remaining seconds
	return reply.MakeIntReply(remaining / 1000) // Convert from milliseconds to seconds
}

func init() {
	RegisterCommand("Del", execDel, -2)
	RegisterCommand("Exists", execExists, -2)
	RegisterCommand("Keys", execKeys, 2)
	RegisterCommand("FlushDB", execFlushDB, -1)
	RegisterCommand("Type", execType, 2)
	RegisterCommand("Rename", execRename, 3)
	RegisterCommand("RenameNx", execRenameNx, 3)
	RegisterCommand("Expire", execExpire, 3)
	RegisterCommand("TTL", execTTL, 2)
}

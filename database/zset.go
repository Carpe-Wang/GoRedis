package database

import (
	"goRedis/datastruct/sortedset"
	"goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/resp/reply"
	"strconv"
	"strings"
)

// getAsZSet gets entity as zset
func (db *DB) getAsZSet(key string) (*sortedset.SortedSet, reply.ErrorReply) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, nil
	}
	zset, ok := entity.Data.(*sortedset.SortedSet)
	if !ok {
		return nil, &reply.WrongTypeErrReply{}
	}
	return zset, nil
}

func (db *DB) getOrInitZSet(key string) (zset *sortedset.SortedSet, inited bool, errReply reply.ErrorReply) {
	zset, errReply = db.getAsZSet(key)
	if errReply != nil {
		return nil, false, errReply
	}
	inited = false
	if zset == nil {
		// create new set
		zset = sortedset.Make()
		db.PutEntity(key, &database.DataEntity{
			Data: zset,
		})
		inited = true
	}
	return zset, inited, nil
}

// execZAdd adds member to sorted set
func execZAdd(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 || len(args)%2 == 0 {
		return reply.MakeSyntaxErrReply()
	}
	key := string(args[0])
	zset, _, errReply := db.getOrInitZSet(key)
	if errReply != nil {
		return errReply
	}

	// parse options
	var options int
	for i := 1; i < len(args); {
		arg := strings.ToUpper(string(args[i]))
		if arg != "NX" && arg != "XX" {
			break
		}
		if arg == "NX" {
			options |= 1
		} else if arg == "XX" {
			options |= 2
		}
		i++
	}

	// parse score-member pairs
	pairs := make([]sortedset.Element, 0)
	for i := 1; i < len(args); i += 2 {
		scoreBytes := args[i]
		member := string(args[i+1])
		score, err := strconv.ParseFloat(string(scoreBytes), 64)
		if err != nil {
			return reply.MakeErrReply("ERR value is not a valid float")
		}
		pairs = append(pairs, sortedset.Element{
			Member: member,
			Score:  score,
		})
	}

	// execute
	var addedCount int64 = 0
	for _, pair := range pairs {
		isNew := zset.Add(pair.Member, pair.Score)
		if isNew {
			addedCount++
		}
	}
	db.addAof(utils.ToCmdLine2("zadd", args...))
	return reply.MakeIntReply(addedCount)
}

// execZScore gets score of member in sorted set
func execZScore(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	member := string(args[1])
	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return &reply.NullBulkReply{}
	}

	score, exists := zset.GetScore(member)
	if !exists {
		return &reply.NullBulkReply{}
	}
	return reply.MakeBulkReply([]byte(strconv.FormatFloat(score, 'f', -1, 64)))
}

// execZRank gets rank of member in sorted set
func execZRank(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	member := string(args[1])
	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return &reply.NullBulkReply{}
	}

	rank, exists := zset.GetRank(member, false)
	if !exists {
		return &reply.NullBulkReply{}
	}
	return reply.MakeIntReply(rank)
}

// execZRevRank gets rank of member in sorted set in reverse order
func execZRevRank(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	member := string(args[1])
	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return &reply.NullBulkReply{}
	}

	rank, exists := zset.GetRank(member, true)
	if !exists {
		return &reply.NullBulkReply{}
	}
	return reply.MakeIntReply(rank)
}

// execZCard gets number of members in sorted set
func execZCard(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(zset.Len())
}

// execZRange gets members in range
func execZRange(db *DB, args [][]byte) resp.Reply {
	withScores := false
	if len(args) >= 4 {
		if strings.ToUpper(string(args[3])) == "WITHSCORES" {
			withScores = true
		}
	}
	return rangeByRank(db, args[0], args[1], args[2], withScores, false)
}

// execZRevRange gets members in range in reverse order
func execZRevRange(db *DB, args [][]byte) resp.Reply {
	withScores := false
	if len(args) >= 4 {
		if strings.ToUpper(string(args[3])) == "WITHSCORES" {
			withScores = true
		}
	}
	return rangeByRank(db, args[0], args[1], args[2], withScores, true)
}

// rangeByRank is the underlying implementation of range and rev_range
func rangeByRank(db *DB, key []byte, startBytes []byte, stopBytes []byte, withScores bool, reverse bool) resp.Reply {
	zset, errReply := db.getAsZSet(string(key))
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return &reply.EmptyMultiBulkReply{}
	}

	start, err := strconv.ParseInt(string(startBytes), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop, err := strconv.ParseInt(string(stopBytes), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}

	// handle out of range values
	if start < 0 {
		start = zset.Len() + start
	}
	if stop < 0 {
		stop = zset.Len() + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= zset.Len() {
		stop = zset.Len() - 1
	}
	if start > stop {
		return &reply.EmptyMultiBulkReply{}
	}

	// get elements
	elements := make([]*sortedset.Element, 0)
	zset.Range(start, stop, reverse, func(element *sortedset.Element) bool {
		elements = append(elements, element)
		return true
	})

	// format reply
	size := len(elements)
	if withScores {
		result := make([][]byte, 2*size)
		for i, element := range elements {
			result[2*i] = []byte(element.Member)
			result[2*i+1] = []byte(strconv.FormatFloat(element.Score, 'f', -1, 64))
		}
		return reply.MakeMultiBulkReply(result)
	} else {
		result := make([][]byte, size)
		for i, element := range elements {
			result[i] = []byte(element.Member)
		}
		return reply.MakeMultiBulkReply(result)
	}
}

// execZRem removes members
func execZRem(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return reply.MakeIntReply(0)
	}

	count := 0
	for i := 1; i < len(args); i++ {
		member := string(args[i])
		if zset.Remove(member) {
			count++
		}
	}
	if count > 0 {
		db.addAof(utils.ToCmdLine2("zrem", args...))
	}
	return reply.MakeIntReply(int64(count))
}

// execZIncrBy increments score of member
func execZIncrBy(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	increment, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not a valid float")
	}
	member := string(args[2])

	zset, _, errReply := db.getOrInitZSet(key)
	if errReply != nil {
		return errReply
	}

	// get current score
	score, exists := zset.GetScore(member)
	if !exists {
		score = 0
	}
	score += increment
	zset.Add(member, score)

	db.addAof(utils.ToCmdLine2("zincrby", args...))
	return reply.MakeBulkReply([]byte(strconv.FormatFloat(score, 'f', -1, 64)))
}

// execZCount counts members with score in range
func execZCount(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	min, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}
	max, err := strconv.ParseFloat(string(args[2]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}

	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return reply.MakeIntReply(0)
	}

	return reply.MakeIntReply(zset.Count(min, max))
}

// execZRangeByScore gets members with score in range
func execZRangeByScore(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	min, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}
	max, err := strconv.ParseFloat(string(args[2]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}

	withScores := false
	var offset, limit int64 = 0, -1
	if len(args) > 3 {
		for i := 3; i < len(args); i++ {
			arg := strings.ToUpper(string(args[i]))
			if arg == "WITHSCORES" {
				withScores = true
			} else if arg == "LIMIT" && i+2 < len(args) {
				offsetStr := string(args[i+1])
				limitStr := string(args[i+2])
				offset, err = strconv.ParseInt(offsetStr, 10, 64)
				if err != nil {
					return reply.MakeErrReply("ERR value is not an integer or out of range")
				}
				limit, err = strconv.ParseInt(limitStr, 10, 64)
				if err != nil {
					return reply.MakeErrReply("ERR value is not an integer or out of range")
				}
				i += 2
			}
		}
	}

	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return &reply.EmptyMultiBulkReply{}
	}

	elements := zset.GetByScoreRange(min, max, offset, limit, false)
	if withScores {
		result := make([][]byte, 2*len(elements))
		for i, element := range elements {
			result[2*i] = []byte(element.Member)
			result[2*i+1] = []byte(strconv.FormatFloat(element.Score, 'f', -1, 64))
		}
		return reply.MakeMultiBulkReply(result)
	} else {
		result := make([][]byte, len(elements))
		for i, element := range elements {
			result[i] = []byte(element.Member)
		}
		return reply.MakeMultiBulkReply(result)
	}
}

// execZRevRangeByScore gets members with score in range in reverse order
func execZRevRangeByScore(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	max, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}
	min, err := strconv.ParseFloat(string(args[2]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}

	withScores := false
	var offset, limit int64 = 0, -1
	if len(args) > 3 {
		for i := 3; i < len(args); i++ {
			arg := strings.ToUpper(string(args[i]))
			if arg == "WITHSCORES" {
				withScores = true
			} else if arg == "LIMIT" && i+2 < len(args) {
				offsetStr := string(args[i+1])
				limitStr := string(args[i+2])
				offset, err = strconv.ParseInt(offsetStr, 10, 64)
				if err != nil {
					return reply.MakeErrReply("ERR value is not an integer or out of range")
				}
				limit, err = strconv.ParseInt(limitStr, 10, 64)
				if err != nil {
					return reply.MakeErrReply("ERR value is not an integer or out of range")
				}
				i += 2
			}
		}
	}

	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return &reply.EmptyMultiBulkReply{}
	}

	elements := zset.GetByScoreRange(min, max, offset, limit, true)
	if withScores {
		result := make([][]byte, 2*len(elements))
		for i, element := range elements {
			result[2*i] = []byte(element.Member)
			result[2*i+1] = []byte(strconv.FormatFloat(element.Score, 'f', -1, 64))
		}
		return reply.MakeMultiBulkReply(result)
	} else {
		result := make([][]byte, len(elements))
		for i, element := range elements {
			result[i] = []byte(element.Member)
		}
		return reply.MakeMultiBulkReply(result)
	}
}

// execZRemRangeByRank removes members with rank in range
func execZRemRangeByRank(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	start, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}

	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return reply.MakeIntReply(0)
	}

	// handle out of range values
	if start < 0 {
		start = zset.Len() + start
	}
	if stop < 0 {
		stop = zset.Len() + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= zset.Len() {
		stop = zset.Len() - 1
	}
	if start > stop {
		return reply.MakeIntReply(0)
	}

	// get elements to remove
	elements := make([]*sortedset.Element, 0)
	zset.Range(start, stop, false, func(element *sortedset.Element) bool {
		elements = append(elements, element)
		return true
	})

	// remove elements
	count := 0
	for _, element := range elements {
		if zset.Remove(element.Member) {
			count++
		}
	}

	if count > 0 {
		db.addAof(utils.ToCmdLine2("zremrangebyrank", args...))
	}
	return reply.MakeIntReply(int64(count))
}

// execZRemRangeByScore removes members with score in range
func execZRemRangeByScore(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	min, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}
	max, err := strconv.ParseFloat(string(args[2]), 64)
	if err != nil {
		return reply.MakeErrReply("ERR min or max is not a float")
	}

	zset, errReply := db.getAsZSet(key)
	if errReply != nil {
		return errReply
	}
	if zset == nil {
		return reply.MakeIntReply(0)
	}

	// get elements to remove
	elements := zset.GetByScoreRange(min, max, 0, -1, false)

	// remove elements
	count := 0
	for _, element := range elements {
		if zset.Remove(element.Member) {
			count++
		}
	}

	if count > 0 {
		db.addAof(utils.ToCmdLine2("zremrangebyscore", args...))
	}
	return reply.MakeIntReply(int64(count))
}

func init() {
	RegisterCommand("ZAdd", execZAdd, -4)
	RegisterCommand("ZScore", execZScore, 3)
	RegisterCommand("ZRank", execZRank, 3)
	RegisterCommand("ZRevRank", execZRevRank, 3)
	RegisterCommand("ZCard", execZCard, 2)
	RegisterCommand("ZRange", execZRange, -4)
	RegisterCommand("ZRevRange", execZRevRange, -4)
	RegisterCommand("ZRem", execZRem, -3)
	RegisterCommand("ZIncrBy", execZIncrBy, 4)
	RegisterCommand("ZCount", execZCount, 4)
	RegisterCommand("ZRangeByScore", execZRangeByScore, -4)
	RegisterCommand("ZRevRangeByScore", execZRevRangeByScore, -4)
	RegisterCommand("ZRemRangeByRank", execZRemRangeByRank, 4)
	RegisterCommand("ZRemRangeByScore", execZRemRangeByScore, 4)
}

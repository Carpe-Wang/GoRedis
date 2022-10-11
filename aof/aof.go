package aof

import (
	"goRedis/config"
	databaseface "goRedis/interface/database"
	"goRedis/lib/logger"
	"goRedis/lib/utils"
	"goRedis/resp/connection"
	"goRedis/resp/parser"
	"goRedis/resp/reply"
	"io"
	"os"
	"strconv"
	"sync"
)

type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16 //避免魔法值 65535
)

type payload struct {
	cmdLine CmdLine //指令本身
	dbIndex int     //写入那个DB
}

type AofHandler struct {
	db          databaseface.Database //Redis核心
	aofChan     chan *payload         //写文件的一个缓冲区，文件要落入到硬盘中，速度较慢，需要加Chan
	aofFile     *os.File              //后期读取appendonly.aof文件
	aofFilename string
	aofFinished chan struct{}
	pausingAof  sync.RWMutex
	currentDB   int //记录指令保存到那个DB
}

// NewAOFHandler 新建handler
func NewAOFHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename //找到配置文件的文件名
	handler.db = db
	handler.LoadAof(0)
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600) //入参依次是，文件名，flag（只读，只写，追加），文件模式
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize) //设置Chan长度，防止落硬盘速度慢。
	handler.aofFinished = make(chan struct{})
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof 用户的指令包装成payload放入缓冲区
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil { //判断是否开AOF功能
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// handleAof 从缓冲区取出，保存到硬盘中
func (handler *AofHandler) handleAof() {
	// serialized execution
	handler.currentDB = 0
	for p := range handler.aofChan {
		handler.pausingAof.RLock() // prevent other goroutines from pausing aof
		if p.dbIndex != handler.currentDB {
			// select db
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue // skip this command
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
		handler.pausingAof.RUnlock()
	}
	handler.aofFinished <- struct{}{}
}

// LoadAof 重启系统后从文件中加载到内存中，防止数据丢失
func (handler *AofHandler) LoadAof(maxBytes int) {
	aofChan := handler.aofChan
	handler.aofChan = nil
	defer func(aofChan chan *payload) {
		handler.aofChan = aofChan
	}(aofChan)

	file, err := os.Open(handler.aofFilename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return
		}
		logger.Warn(err)
		return
	}
	defer file.Close()

	var reader io.Reader
	if maxBytes > 0 {
		reader = io.LimitReader(file, int64(maxBytes))
	} else {
		reader = file
	}
	ch := parser.ParseStream(reader)
	fakeConn := &connection.FakeConn{}
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		ret := handler.db.Exec(fakeConn, r.Args)
		if reply.IsErrorReply(ret) {
			logger.Error("exec err", err)
		}
	}
}

// Close gracefully stops aof persistence procedure
func (handler *AofHandler) Close() {
	if handler.aofFile != nil {
		close(handler.aofChan)
		<-handler.aofFinished // wait for aof finished
		err := handler.aofFile.Close()
		if err != nil {
			logger.Warn(err)
		}
	}
}

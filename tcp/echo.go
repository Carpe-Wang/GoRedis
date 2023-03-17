package tcp

import (
	"goRedis/lib/sync/atomic"
	"goRedis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

/**
 * 方便后期测试TCP服务的性能
 */

import (
	"bufio"
	"context"
	"goRedis/lib/logger"
	"io"
)

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait //多实现一个超时功能
}

func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second) //等待一段时间，防止任务一直做不完
	c.Conn.Close()                              //因为这里直接关闭了，就不对err做处理了
	return nil
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		// 可能会发生clientEOF，client超时，handler提前关闭
		msg, err := reader.ReadString('\n') //
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// Close 关闭停止echoHandler
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}

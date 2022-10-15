package tcp

import (
	"context"
	"net"
)

// HandleFunc 表示程序处理程序函数
type HandleFunc func(ctx context.Context, conn net.Conn)

// Handler 表示tcp上的程序处理程序
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

package tcp

import (
	"context"
	"net"
)

// Handler represents application server over tcp
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

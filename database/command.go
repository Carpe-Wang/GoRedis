package database

import (
	"strings"
)

var cmdTable = make(map[string]*command)

type command struct {
	executor ExecFunc
	arity    int // 参数数量
}

// RegisterCommand
// arity允许命令参数数量,如果arity < 0 就意味着len()args >= -arity
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		arity:    arity,
	}
}

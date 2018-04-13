package command

import (
	"sync"

	"github.com/NANNERPISS/NANNERPISS/context"
)

var (
	cmdsMu sync.RWMutex
	cmds   = make(map[string]context.BotFunc)
)

func Register(name string, function context.BotFunc) {
	cmdsMu.Lock()
	defer cmdsMu.Unlock()
	if _, dup := cmds[name]; dup {
		panic("command: Register called twice for command " + name)
	}
	cmds[name] = function
}

func Get(name string) (context.BotFunc, bool) {
	cmdsMu.RLock()
	defer cmdsMu.RUnlock()
	cmd, ok := cmds[name]
	return cmd, ok
}

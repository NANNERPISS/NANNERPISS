package command

import (
	"sync"

	"github.com/NANNERPISS/NANNERPISS/context"

	"gopkg.in/telegram-bot-api.v4"
)

type cmdFunc func(*context.Context, *tgbotapi.Message) error

var (
	cmdsMu sync.RWMutex
	cmds   = make(map[string]cmdFunc)
)

func Register(name string, function cmdFunc) {
	cmdsMu.Lock()
	defer cmdsMu.Unlock()
	if _, dup := cmds[name]; dup {
		panic("command: Register called twice for command " + name)
	}
	cmds[name] = function
}

func Get(name string) (cmdFunc, bool) {
	cmdsMu.RLock()
	defer cmdsMu.RUnlock()
	cmd, ok := cmds[name]
	return cmd, ok
}

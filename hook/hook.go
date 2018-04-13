package hook

import (
	"github.com/NANNERPISS/NANNERPISS/context"
)

type Hook struct {
	Name string
	Func context.BotFunc
}

var Hooks []Hook

func Register(name string, function context.BotFunc) {
	Hooks = append(Hooks, Hook{Name: name, Func: function})
}

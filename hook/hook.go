package hook

import (
	"github.com/NANNERPISS/NANNERPISS/context"

	"gopkg.in/telegram-bot-api.v4"
)

type hookFunc func(*context.Context, *tgbotapi.Message) error

type Hook struct {
	Name string
	Func hookFunc
}

var Hooks []Hook

func Register(name string, function hookFunc) {
	Hooks = append(Hooks, Hook{Name: name, Func: function})
}

package hook

import (
	"sync"

	"github.com/NANNERPISS/NANNERPISS/config"
	"github.com/NANNERPISS/NANNERPISS/db"

	"gopkg.in/telegram-bot-api.v4"
)

type Context struct {
	Config *config.Config
	DB     db.DB
	TG     *tgbotapi.BotAPI
	cache  struct {
		mu   sync.RWMutex
		data map[string]interface{}
	}
}

type hookFunc func(*Context, *tgbotapi.Message) error

type Hook struct {
	Name string
	Func hookFunc
}

var Hooks []Hook

func Register(name string, function hookFunc) {
	Hooks = append(Hooks, Hook{Name: name, Func: function})
}

package context

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
	Cache  struct {
		Mu   sync.RWMutex
		Data map[string]interface{}
	}
}

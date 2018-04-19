package context

import (
	"sync"

	"github.com/NANNERPISS/NANNERPISS/config"
	"github.com/NANNERPISS/NANNERPISS/db"

	"github.com/ChimeraCoder/anaconda"
	"gopkg.in/telegram-bot-api.v4"
)

type BotFunc func(*Context, *tgbotapi.Message) error

type Context struct {
	Config *config.Config
	DB     db.DB
	TG     *tgbotapi.BotAPI
	TW     *anaconda.TwitterApi
	Cache  struct {
		Mu   sync.RWMutex
		Data map[string]interface{}
	}
}

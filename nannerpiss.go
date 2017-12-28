package bot

import (
	"fmt"

	"github.com/NANNERPISS/NANNERPISS/command"
	"github.com/NANNERPISS/NANNERPISS/config"
	"github.com/NANNERPISS/NANNERPISS/db"
	"github.com/NANNERPISS/NANNERPISS/hook"

	"gopkg.in/telegram-bot-api.v4"
)

type Bot struct {
	Config *config.Config
	DB     db.DB
	TG     *tgbotapi.BotAPI
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.TG.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	cmdCtx := &command.Context{Config: b.Config, DB: b.DB, TG: b.TG}
	hookCtx := &hook.Context{Config: b.Config, DB: b.DB, TG: b.TG}

	for update := range updates {
		go (func(update tgbotapi.Update) {
			message := update.Message
			if message == nil {
				return
			}
			if !message.Chat.IsGroup() && !message.Chat.IsSuperGroup() {
				return
			}
			if message.IsCommand() {
				cmdName := message.Command()
				cmd, ok := command.Get(cmdName)
				if ok {
					go (func(cmdCtx *command.Context, cmdName string, message *tgbotapi.Message) {
						err := cmd(cmdCtx, message)
						if err != nil {
							fmt.Printf("[%s] %s\n", cmdName, err)
						}
					})(cmdCtx, cmdName, message)
				}
			}

			for _, h := range hook.Hooks {
				go (func(hookCtx *hook.Context, h hook.Hook, message *tgbotapi.Message) {
					err := h.Func(hookCtx, message)
					if err != nil {
						fmt.Printf("[%s] %s\n", h.Name, err)
					}
				})(hookCtx, h, message)
			}

			fmt.Printf("[%s] %s\n", message.From.UserName, message.Text)
		})(update)
	}
}

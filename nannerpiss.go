package bot

import (
	"fmt"

	"github.com/NANNERPISS/NANNERPISS/command"
	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/hook"

	"gopkg.in/telegram-bot-api.v4"
)

type bot struct {
	*context.Context
}

func New() *bot {
	return &bot{Context: &context.Context{}}
}

func (b *bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.TG.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		go (func(update tgbotapi.Update) {
			var message *tgbotapi.Message
			if message = update.Message; message == nil {
				if message = update.EditedMessage; message == nil {
					// No usable message
					return
				}
			}
			if !message.Chat.IsGroup() && !message.Chat.IsSuperGroup() {
				return
			}
			if message.IsCommand() {
				cmdName := message.Command()
				cmd, ok := command.Get(cmdName)
				if ok {
					go (func(ctx *context.Context, cmdName string, message *tgbotapi.Message) {
						err := cmd(ctx, message)
						if err != nil {
							fmt.Printf("[%s] %s\n", cmdName, err)
						}
					})(b.Context, cmdName, message)
				}
			}

			for _, h := range hook.Hooks {
				go (func(ctx *context.Context, h hook.Hook, message *tgbotapi.Message) {
					err := h.Func(ctx, message)
					if err != nil {
						fmt.Printf("[%s] %s\n", h.Name, err)
					}
				})(b.Context, h, message)
			}

			fmt.Printf("[%s] %s\n", message.From.UserName, message.Text)
		})(update)
	}
}

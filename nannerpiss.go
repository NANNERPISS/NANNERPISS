package bot

import (
	"fmt"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/command"
	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/hook"

	//"github.com/davecgh/go-spew/spew"
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
			if update.Message != nil {
				message = update.Message
			} else if update.EditedMessage != nil {
				message = update.EditedMessage
			} else {
				// No usable message
				return
			}

			if !message.Chat.IsGroup() && !message.Chat.IsSuperGroup() {
				return
			}

			//messageInfo := spew.Sdump(message)
			//fmt.Printf(messageInfo)

			if message.Text == "" && message.Caption != "" && message.Entities == nil {
				message.Text = message.Caption
				if strings.HasPrefix(message.Caption, "/") {
					c := strings.SplitN(message.Caption, " ", 2)[0]
					e := tgbotapi.MessageEntity{Type: "bot_command", Offset: 0, Length: len(c)}
					message.Entities = &[]tgbotapi.MessageEntity{e}
				}
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

			fmt.Printf("[%d][%s] %s\n", message.Chat.ID, message.From.UserName, message.Text)
		})(update)
	}
}

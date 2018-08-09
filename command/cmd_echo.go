package command

import (
	"github.com/NANNERPISS/NANNERPISS/context"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("echo", Echo)
}

func Echo(ctx *context.Context, message *tgbotapi.Message) error {
	var args string
	if args = message.CommandArguments(); args == "" {
		return nil
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, args)
	_, err := ctx.TG.Send(reply)
	return err
}

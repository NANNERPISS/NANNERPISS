package command

import (
	"github.com/NANNERPISS/NANNERPISS/context"

	"github.com/go-telegram-bot-api/telegram-bot-api"
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

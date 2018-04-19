package command

import (
	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

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

	reply := util.ReplyTo(message, args, "")
	_, err := ctx.TG.Send(reply)
	return err
}

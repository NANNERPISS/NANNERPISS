package middleware

import (
	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func Admin(cmd context.BotFunc) context.BotFunc {
	return func(ctx *context.Context, message *tgbotapi.Message) error {
		if message.ForwardFrom != nil {
			return nil
		}

		sender, err := util.GetSender(ctx.TG, message)
		if err != nil {
			return err
		}

		if sender.IsAdministrator() || sender.IsCreator() {
			return cmd(ctx, message)
		}

		return nil
	}
}

package middleware

import (
	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"github.com/ChimeraCoder/anaconda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func TweetError(cmd context.BotFunc) context.BotFunc {
	return func(ctx *context.Context, message *tgbotapi.Message) error {
		err := cmd(ctx, message)
		if terr, ok := err.(*anaconda.ApiError); ok {
			reply := util.ReplyTo(message, terr.Decoded.Error(), "")
			_, err = ctx.TG.Send(reply)
		}

		return err
	}
}

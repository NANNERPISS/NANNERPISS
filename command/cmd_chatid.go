package command

import (
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/middleware"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("chatid", middleware.Admin(ChatID))
}

func ChatID(ctx *context.Context, message *tgbotapi.Message) error {
	reply := util.ReplyTo(message, strconv.FormatInt(message.Chat.ID, 10), "")
	_, err := ctx.TG.Send(reply)
	return err
}

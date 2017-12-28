package command

import (
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("chatid", ChatID)
}

func ChatID(ctx *Context, message *tgbotapi.Message) error {
	sender, err := util.GetSender(ctx.TG, message)
	if err != nil {
		return err
	}

	if sender.IsAdministrator() || sender.IsCreator() {
		reply := util.ReplyTo(message, strconv.FormatInt(message.Chat.ID, 10))
		_, err := ctx.TG.Send(reply)
		return err
	}

	return nil
}

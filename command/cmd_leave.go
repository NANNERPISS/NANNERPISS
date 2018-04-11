package command

import (
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("leave", Leave)
}

func Leave(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID == ctx.Config.TG.ControlGroup {
		sender, err := util.GetSender(ctx.TG, message)
		if err != nil {
			return err
		}

		if sender.IsAdministrator() || sender.IsCreator() {
			if args := message.CommandArguments(); args != "" {
				chatID, err := strconv.ParseInt(args, 10, 64)
				if err != nil {
					return err
				}
				resp, err := ctx.TG.LeaveChat(tgbotapi.ChatConfig{ChatID: chatID})
				if err != nil {
					return err
				}
				if resp.Ok {
					reply := util.ReplyTo(message, "Left chat", "")
					_, err := ctx.TG.Send(reply)
					return err
				}
			}
		}
	}

	return nil
}

package command

import (
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/middleware"
	"github.com/NANNERPISS/NANNERPISS/util"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func init() {
	Register("leave", middleware.Admin(Leave))
}

func Leave(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID != ctx.Config.TG.ControlGroup {
		return nil
	}

	var args string
	if args = message.CommandArguments(); args == "" {
		reply := util.ReplyTo(message, "Please include the chat ID to leave", "")
		_, err := ctx.TG.Send(reply)
		return err
	}

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

	return nil
}

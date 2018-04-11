package command

import (
	"fmt"
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("chatinfo", Admin(ChatInfo))
}

func ChatInfo(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID != ctx.Config.TG.ControlGroup {
		return nil
	}
	
	var args string
	if args = message.CommandArguments(); args == "" {
		return nil
	}
	
	chatID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		return err
	}
	chat, err := ctx.TG.GetChat(tgbotapi.ChatConfig{ChatID: chatID})
	if err != nil {
		return err
	}
	chatInfoString := spew.Sdump(chat)
	fmt.Printf(chatInfoString)
	reply := util.ReplyTo(message, chatInfoString, "")
	_, err = ctx.TG.Send(reply)
	return err
}

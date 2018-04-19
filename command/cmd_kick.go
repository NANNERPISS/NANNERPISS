package command

import (
	"fmt"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/middleware"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("ban", middleware.Admin(Kick))
	Register("kick", middleware.Admin(Kick))
}

func Kick(ctx *context.Context, message *tgbotapi.Message) error {
	if message.ReplyToMessage == nil {
		reply := util.ReplyTo(message, "Please respond to the user you want to kick", "")
		_, err := ctx.TG.Send(reply)
		return err
	}

	id := message.ReplyToMessage.From.ID

	chatMemberConfig := tgbotapi.ChatMemberConfig{ChatID: message.Chat.ID, UserID: id}
	chatMember, err := ctx.TG.GetChatMember(tgbotapi.ChatConfigWithUser{ChatID: message.Chat.ID, UserID: id})
	if err != nil {
		return err
	}
	userStr, err := util.FormatUser(chatMember.User)
	if err != nil {
		return err
	}
	resp, err := ctx.TG.KickChatMember(tgbotapi.KickChatMemberConfig{ChatMemberConfig: chatMemberConfig})
	if !resp.Ok {
		response := fmt.Sprintf("Sorry, I can't kick %s", userStr)
		reply := util.ReplyTo(message, response, "html")
		_, err = ctx.TG.Send(reply)
		return err
	}

	response := fmt.Sprintf("Kicked %s", userStr)
	reply := util.ReplyTo(message, response, "html")
	_, err = ctx.TG.Send(reply)
	return err
}

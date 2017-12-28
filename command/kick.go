package command

import (
	"fmt"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("kick", Kick)
}

func Kick(ctx *context.Context, message *tgbotapi.Message) error {
	sender, err := util.GetSender(ctx.TG, message)
	if err != nil {
		return err
	}

	if sender.IsAdministrator() || sender.IsCreator() {
		var kickUserIDs []int
		if message.ReplyToMessage != nil {
			kickUserIDs = append(kickUserIDs, message.ReplyToMessage.From.ID)
		}

		// TODO: Apparently there's no way to get the user ID of an @'d user without storing them client side,
		// so only works with replies for now
		for _, id := range kickUserIDs {
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
				reply := util.ReplyTo(message, response)
				_, err = ctx.TG.Send(reply)
				return err
			}

			response := fmt.Sprintf("Kicked %s", userStr)
			reply := util.ReplyTo(message, response)
			_, err = ctx.TG.Send(reply)
			return err
		}
	}

	return nil
}

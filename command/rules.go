package command

import (
	"fmt"

	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("rules", Rules)
	Register("rulesset", RulesSet)
}

func Rules(ctx *Context, message *tgbotapi.Message) error {
	rules, err := ctx.DB.RulesGet(message.Chat.ID)
	if err != nil {
		return err
	}

	reply := util.ReplyTo(message, rules)
	_, err = ctx.TG.Send(reply)

	return err
}

func RulesSet(ctx *Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		sender, err := util.GetSender(ctx.TG, message)
		if err != nil {
			return err
		}

		if sender.IsAdministrator() || sender.IsCreator() {
			err = ctx.DB.RulesSet(message.Chat.ID, args)
			if err != nil {
				return err
			}

			response := fmt.Sprintf(`<b>Rules have been updated</b>`)
			reply := util.ReplyTo(message, response)
			_, err = ctx.TG.Send(reply)

			return err
		}
	}

	return nil
}

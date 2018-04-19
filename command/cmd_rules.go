package command

import (
	"fmt"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/middleware"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("rules", Rules)
	Register("rulesset", middleware.Admin(RulesSet))
}

func Rules(ctx *context.Context, message *tgbotapi.Message) error {
	rules, err := ctx.DB.RulesGet(message.Chat.ID)
	if err != nil {
		return err
	}

	reply := util.ReplyTo(message, rules, "")
	_, err = ctx.TG.Send(reply)

	return err
}

func RulesSet(ctx *context.Context, message *tgbotapi.Message) error {
	var args string
	if args = message.CommandArguments(); args == "" {
		reply := util.ReplyTo(message, "Rules cannot be blank", "")
		_, err := ctx.TG.Send(reply)
		return err
	}

	err := ctx.DB.RulesSet(message.Chat.ID, args)
	if err != nil {
		return err
	}

	response := fmt.Sprintf(`<b>Rules have been updated</b>`)
	reply := util.ReplyTo(message, response, "HTML")
	_, err = ctx.TG.Send(reply)
	return err
}

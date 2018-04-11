package command

import (
	"fmt"
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("warn", Warn)
	Register("warnmaxset", WarnMaxSet)
}

func Warn(ctx *context.Context, message *tgbotapi.Message) error {
	sender, err := util.GetSender(ctx.TG, message)
	if err != nil {
		return err
	}

	if sender.IsAdministrator() || sender.IsCreator() {
		if message.ReplyToMessage == nil {
			reply := util.ReplyTo(message, "Please reply to the user you want to warn", "")
			_, err := ctx.TG.Send(reply)
			return err
		}

		err := ctx.DB.WarnAdd(message.Chat.ID, message.ReplyToMessage.From.ID)
		if err != nil {
			return err
		}

		warnCount, err := ctx.DB.WarnCount(message.Chat.ID, message.ReplyToMessage.From.ID)
		if err != nil {
			return err
		}

		warnMax, err := ctx.DB.WarnMax(message.Chat.ID)
		if err != nil {
			return err
		}

		userStr, err := util.FormatUser(message.ReplyToMessage.From)
		if err != nil {
			return err
		}
		warnMsg := fmt.Sprintf(`%s<b> has been warned for this message</b> (<code>%d/%d</code>)`, userStr, warnCount, warnMax)

		reply := util.ReplyTo(message.ReplyToMessage, warnMsg, "HTML")
		_, err = ctx.TG.Send(reply)
		if err != nil {
			return err
		}

		if warnCount >= warnMax {
			err = ctx.DB.WarnSet(message.Chat.ID, message.ReplyToMessage.From.ID, 0)
			if err != nil {
				return err
			}
			
			err = Kick(ctx, message)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func WarnMaxSet(ctx *context.Context, message *tgbotapi.Message) error {
	sender, err := util.GetSender(ctx.TG, message)
	if err != nil {
		return err
	}

	if sender.IsAdministrator() || sender.IsCreator() {
		if args := message.CommandArguments(); args != "" {
			parsedCount, err := strconv.Atoi(args)
			if err != nil {
				return err
			}

			err = ctx.DB.WarnMaxSet(message.Chat.ID, parsedCount)
			if err != nil {
				return err
			}

			reply := util.ReplyTo(message, fmt.Sprintf(`<b>Max Warning Count</b> has been set to <code>%d</code>`, parsedCount), "HTML")
			_, err = ctx.TG.Send(reply)

			return err
		}
	}

	return nil
}

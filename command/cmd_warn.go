package command

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/middleware"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("warn", middleware.Admin(Warn))
	Register("warnmaxset", middleware.Admin(WarnMaxSet))
	Register("warnings", Warnings)
}

func Warn(ctx *context.Context, message *tgbotapi.Message) error {
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

	return nil
}

func WarnMaxSet(ctx *context.Context, message *tgbotapi.Message) error {
	var args string
	if args = message.CommandArguments(); args == "" {
		reply := util.ReplyTo(message, "Please include the max warn amount to set", "")
		_, err := ctx.TG.Send(reply)
		return err
	}

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

func Warnings(ctx *context.Context, message *tgbotapi.Message) error {
	var replyMessage *tgbotapi.Message
	if message.ReplyToMessage != nil {
		replyMessage = message.ReplyToMessage
	} else if message.From != nil {
		replyMessage = message
	} else {
		reply := util.ReplyTo(message, "Please reply to the user whose warning count you want the see", "")
		_, err := ctx.TG.Send(reply)
		return err
	}

	warnCount, err := ctx.DB.WarnCount(message.Chat.ID, replyMessage.From.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		warnCount = 0
	}

	warnMax, err := ctx.DB.WarnMax(message.Chat.ID)
	if err != nil {
		return err
	}

	userStr, err := util.FormatUser(replyMessage.From)
	if err != nil {
		return err
	}

	plural := "s"
	if warnCount == 1 {
		plural = ""
	}
	warnMsg := fmt.Sprintf(`%s<b> currently has</b> <code>%d/%d</code> <b>warning%s</b>`, userStr, warnCount, warnMax, plural)
	reply := util.ReplyTo(replyMessage, warnMsg, "HTML")
	_, err = ctx.TG.Send(reply)

	return err
}

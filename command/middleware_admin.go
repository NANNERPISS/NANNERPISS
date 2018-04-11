package command

import (
	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"
	
	"gopkg.in/telegram-bot-api.v4"
)

func Admin(cmd cmdFunc) cmdFunc {
	return func(ctx *context.Context, message *tgbotapi.Message) error {
		sender, err := util.GetSender(ctx.TG, message)
		if err != nil {
			return err
		}

		if sender.IsAdministrator() || sender.IsCreator() {
			return cmd(ctx, message)
		}
		
		return nil
	}
}
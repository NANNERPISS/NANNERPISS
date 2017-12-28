package hook

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("wordlog", WordLog)
}

func WordLog(ctx *context.Context, message *tgbotapi.Message) error {
	if ctx.Cache.Data == nil {
		ctx.Cache.Mu.Lock()
		if ctx.Cache.Data == nil {
			whitelist, err := ctx.DB.WordLogWlGet()
			if err != nil {
				ctx.Cache.Mu.Unlock()
				return err
			}

			blacklist, err := ctx.DB.WordLogBlGet()
			if err != nil {
				ctx.Cache.Mu.Unlock()
				return err
			}

			ctx.Cache.Data = make(map[string]interface{})
			ctx.Cache.Data["whitelist"] = whitelist
			ctx.Cache.Data["blacklist"] = blacklist
		}
		ctx.Cache.Mu.Unlock()
	}

	if message.Chat.ID == ctx.Config.WL.ControlGroup {
		sender, err := util.GetSender(ctx.TG, message)
		if err != nil {
			return err
		}

		if sender.IsAdministrator() || sender.IsCreator() {
			if message.IsCommand() {
				cmdName := message.Command()
				switch cmdName {
				case "whitelistadd":
					err := WordLogWlAdd(ctx, message)
					return err
				case "whitelistdel":
					err := WordLogWlDel(ctx, message)
					return err
				case "blacklistadd":
					err := WordLogBlAdd(ctx, message)
					return err
				case "blacklistdel":
					err := WordLogBlDel(ctx, message)
					return err
				}
			}
		}
	}

	ctx.Cache.Mu.RLock()
	defer ctx.Cache.Mu.RUnlock()

	found, err := WordLogText(ctx, message)
	if err != nil {
		return err
	}
	
	if found {
		err := LogMessage(ctx, message)
		return err
	}
	
	return nil
}

func WordLogText(ctx *context.Context, message *tgbotapi.Message) (bool, error) {
	whitelist, ok := ctx.Cache.Data["whitelist"].([]string)
	if !ok {
		return false, fmt.Errorf("whitelist not available")
	}

	blacklist, ok := ctx.Cache.Data["blacklist"].([]string)
	if !ok {
		return false, fmt.Errorf("blacklist not available")
	}
	
	var messageText string
	if message.Text != "" {
		messageText = message.Text
	} else if message.Caption != "" {
		messageText = message.Caption
	}

	reader := strings.NewReader(messageText)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
	whitelisted:
		for _, word := range blacklist {
			if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower(word)) {
				for _, excluded := range whitelist {
					if strings.ToLower(scanner.Text()) == strings.ToLower(excluded) {
						break whitelisted
					}
				}
				// Not whitelisted
				return true, nil
			}
		}
	}
	
	return false, nil
}

func LogMessage(ctx *context.Context, message *tgbotapi.Message) error {
	forward := tgbotapi.NewForward(ctx.Config.WL.LogChannel, message.Chat.ID, message.MessageID)
	_, err := ctx.TG.Send(forward)
	return err
}

func WordLogWlAdd(ctx *context.Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		word := strings.Split(args, " ")[0]
		err := ctx.DB.WordLogWlAdd(word)
		if err != nil {
			return err
		}

		ctx.Cache.Mu.Lock()
		defer ctx.Cache.Mu.Unlock()

		whitelist, ok := ctx.Cache.Data["whitelist"].([]string)
		if !ok {
			return fmt.Errorf("whitelist not available")
		}

		ctx.Cache.Data["whitelist"] = append(whitelist, word)

		response := fmt.Sprintf("<code>%s</code><b> has been added to the whitelist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

func WordLogWlDel(ctx *context.Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		ctx.Cache.Mu.Lock()
		defer ctx.Cache.Mu.Unlock()

		whitelist, ok := ctx.Cache.Data["whitelist"].([]string)
		if !ok {
			return fmt.Errorf("whitelist not available")
		}

		word := strings.Split(args, " ")[0]
		wordIndex := -1
		for i := range whitelist {
			if whitelist[i] == word {
				wordIndex = i
				break
			}
		}

		if wordIndex == -1 {
			return nil
		}

		err := ctx.DB.WordLogWlDel(word)
		if err != nil {
			return err
		}

		ctx.Cache.Data["whitelist"] = append(whitelist[:wordIndex], whitelist[wordIndex+1:]...)

		response := fmt.Sprintf("<code>%s</code><b> has been remove from the whitelist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

func WordLogBlAdd(ctx *context.Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		word := strings.Split(args, " ")[0]
		err := ctx.DB.WordLogBlAdd(word)
		if err != nil {
			return err
		}

		ctx.Cache.Mu.Lock()
		defer ctx.Cache.Mu.Unlock()

		blacklist, ok := ctx.Cache.Data["blacklist"].([]string)
		if !ok {
			return fmt.Errorf("blacklist not available")
		}

		ctx.Cache.Data["blacklist"] = append(blacklist, word)

		response := fmt.Sprintf("<code>%s</code><b> has been added to the blacklist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

func WordLogBlDel(ctx *context.Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		ctx.Cache.Mu.Lock()
		defer ctx.Cache.Mu.Unlock()

		blacklist, ok := ctx.Cache.Data["blacklist"].([]string)
		if !ok {
			return fmt.Errorf("blacklist not available")
		}

		word := strings.Split(args, " ")[0]
		wordIndex := -1
		for i := range blacklist {
			if blacklist[i] == word {
				wordIndex = i
				break
			}
		}

		if wordIndex == -1 {
			return nil
		}

		err := ctx.DB.WordLogBlDel(word)
		if err != nil {
			return err
		}

		ctx.Cache.Data["blacklist"] = append(blacklist[:wordIndex], blacklist[wordIndex+1:]...)

		response := fmt.Sprintf("<code>%s</code><b> has been remove from the blacklist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

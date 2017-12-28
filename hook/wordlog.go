package hook

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
)

func init() {
	Register("wordlog", WordLog)
}

func WordLog(ctx *Context, message *tgbotapi.Message) error {
	if ctx.cache.data == nil {
		ctx.cache.mu.Lock()
		if ctx.cache.data == nil {
			whitelist, err := ctx.DB.WordLogWlGet()
			if err != nil {
				ctx.cache.mu.Unlock()
				return err
			}

			blacklist, err := ctx.DB.WordLogBlGet()
			if err != nil {
				ctx.cache.mu.Unlock()
				return err
			}

			ctx.cache.data = make(map[string]interface{})
			ctx.cache.data["whitelist"] = whitelist
			ctx.cache.data["blacklist"] = blacklist
		}
		ctx.cache.mu.Unlock()
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

	ctx.cache.mu.RLock()
	defer ctx.cache.mu.RUnlock()

	whitelist, ok := ctx.cache.data["whitelist"].([]string)
	if !ok {
		return fmt.Errorf("whitelist not available")
	}

	blacklist, ok := ctx.cache.data["blacklist"].([]string)
	if !ok {
		return fmt.Errorf("blacklist not available")
	}

	reader := strings.NewReader(message.Text)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)

	var containsWord bool
done:
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
				containsWord = true
				break done
			}
		}
	}
	if containsWord {
		err := LogMessage(ctx, message)
		return err
	}
	return nil
}

func LogMessage(ctx *Context, message *tgbotapi.Message) error {
	forward := tgbotapi.NewForward(ctx.Config.WL.LogChannel, message.Chat.ID, message.MessageID)
	_, err := ctx.TG.Send(forward)
	return err
}

func WordLogWlAdd(ctx *Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		word := strings.Split(args, " ")[0]
		err := ctx.DB.WordLogWlAdd(word)
		if err != nil {
			return err
		}

		ctx.cache.mu.Lock()
		defer ctx.cache.mu.Unlock()

		whitelist, ok := ctx.cache.data["whitelist"].([]string)
		if !ok {
			return fmt.Errorf("whitelist not available")
		}

		ctx.cache.data["whitelist"] = append(whitelist, word)

		response := fmt.Sprintf("<code>%s</code><b> has been added to the whitelist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

func WordLogWlDel(ctx *Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		ctx.cache.mu.Lock()
		defer ctx.cache.mu.Unlock()

		whitelist, ok := ctx.cache.data["whitelist"].([]string)
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

		ctx.cache.data["whitelist"] = append(whitelist[:wordIndex], whitelist[wordIndex+1:]...)

		response := fmt.Sprintf("<code>%s</code><b> has been remove from the whitelist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

func WordLogBlAdd(ctx *Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		word := strings.Split(args, " ")[0]
		err := ctx.DB.WordLogBlAdd(word)
		if err != nil {
			return err
		}

		ctx.cache.mu.Lock()
		defer ctx.cache.mu.Unlock()

		blacklist, ok := ctx.cache.data["blacklist"].([]string)
		if !ok {
			return fmt.Errorf("blacklist not available")
		}

		ctx.cache.data["blacklist"] = append(blacklist, word)

		response := fmt.Sprintf("<code>%s</code><b> has been added to the blacklist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

func WordLogBlDel(ctx *Context, message *tgbotapi.Message) error {
	if args := message.CommandArguments(); args != "" {
		ctx.cache.mu.Lock()
		defer ctx.cache.mu.Unlock()

		blacklist, ok := ctx.cache.data["blacklist"].([]string)
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

		ctx.cache.data["blacklist"] = append(blacklist[:wordIndex], blacklist[wordIndex+1:]...)

		response := fmt.Sprintf("<code>%s</code><b> has been remove from the blacklist</b>", word)
		reply := util.ReplyTo(message, response)
		_, err = ctx.TG.Send(reply)

		return err
	}

	return nil
}

package hook

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"github.com/otiai10/gosseract"
	"golang.org/x/image/webp"
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

	var found bool
	var err error

	found, err = WordLogText(ctx, message)
	if err != nil {
		return err
	}

	if found {
		err = LogMessage(ctx, message)
		return err
	}

	found, err = WordLogFile(ctx, message)
	if err != nil {
		return err
	}

	if found {
		err = LogMessage(ctx, message)
		return err
	}

	return nil
}

func containsBlacklist(ctx *context.Context, messageText string) (bool, error) {
	whitelist, ok := ctx.Cache.Data["whitelist"].([]string)
	if !ok {
		return false, fmt.Errorf("whitelist not available")
	}

	blacklist, ok := ctx.Cache.Data["blacklist"].([]string)
	if !ok {
		return false, fmt.Errorf("blacklist not available")
	}

	reader := strings.NewReader(messageText)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
	whitelisted:
		for _, word := range blacklist {
			if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower(word)) {
				for _, excluded := range whitelist {
					if strings.Index(strings.ToLower(scanner.Text()), strings.ToLower(excluded)) == 0 {
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

func WordLogText(ctx *context.Context, message *tgbotapi.Message) (bool, error) {
	var messageText string
	if message.Text != "" {
		messageText = message.Text
	} else if message.Caption != "" {
		messageText = message.Caption
	} else {
		return false, nil
	}

	blacklisted, err := containsBlacklist(ctx, messageText)

	return blacklisted, err
}

func WordLogFile(ctx *context.Context, message *tgbotapi.Message) (bool, error) {
	var fileID string
	switch {
	case message.Photo != nil:
		if len(*message.Photo) != 0 {
			fileID = (*message.Photo)[len(*message.Photo)-1].FileID
		}
	case message.Document != nil:
		switch message.Document.MimeType {
		case "image/jpeg", "image/png", "image/bmp":
			fileID = message.Document.FileID
		}
	case message.Sticker != nil:
		fileID = message.Sticker.FileID
	}

	if fileID == "" {
		return false, nil
	}

	downloadURL, err := ctx.TG.GetFileDirectURL(fileID)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(ctx.Config.WL.DataDir); os.IsNotExist(err) {
		err = os.MkdirAll(ctx.Config.WL.DataDir, 0700)
		if err != nil {
			return false, err
		}
	}
	outputPath := filepath.Join(ctx.Config.WL.DataDir, fileID)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return false, err
	}
	defer os.Remove(outputPath)
	defer outputFile.Close()

	client := &http.Client{}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer(nil)

	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return false, err
	}

	resp.Body.Close()

	if mime := http.DetectContentType(buf.Bytes()); mime == "image/webp" {
		image, err := webp.Decode(buf)
		if err != nil {
			return false, err
		}

		err = png.Encode(outputFile, image)
		if err != nil {
			return false, err
		}
	} else {
		io.Copy(outputFile, buf)
	}

	outputFile.Close()

	ocr := gosseract.NewClient()
	defer ocr.Close()

	ocr.SetImage(outputPath)

	messageText, err := ocr.Text()
	if err != nil {
		return false, nil
	}

	fmt.Println(messageText)

	blacklisted, err := containsBlacklist(ctx, messageText)

	return blacklisted, err
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

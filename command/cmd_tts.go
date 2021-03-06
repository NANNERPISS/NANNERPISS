package command

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"
	"github.com/NANNERPISS/NANNERPISS/util/parse"
	"github.com/NANNERPISS/NANNERPISS/util/parse/filter"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func init() {
	Register("tts", TTS)
}

const ttsAddress string = `https://tts.fed.bz/`

func TTS(ctx *context.Context, message *tgbotapi.Message) error {
	var msg []string
	if msg = strings.SplitN(message.Text, " ", 2); len(msg) != 2 {
		reply := util.ReplyTo(message, "Please include a message to read.\nYou can choose a voice with /tts@<Voice> <message>.", "")
		_, err := ctx.TG.Send(reply)
		return err
	}

	args := msg[1]

	var voice string
	if c := msg[0]; strings.Contains(c, "@") {
		voice = strings.SplitN(c, "@", 2)[1]
	} else {
		voice = "Agnes"
	}
	client := &http.Client{}
	client.Timeout = time.Second * 30

	form := url.Values{}
	form.Add("voice", voice)
	form.Add("text", args)

	resp, err := client.PostForm(ttsAddress, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	node, err := parse.Parse(resp.Body)
	if err != nil {
		return err
	}

	audioPath, ok := node.FindNode(filter.ID("downloadogg")).GetAttr("href")
	if !ok {
		alertTextNode := node.FindNode(filter.Attr("role", "alert"))
		if alertTextNode == nil || alertTextNode.LastChild == nil {
			return fmt.Errorf("Couldn't parse download link")
		}
		alertText := alertTextNode.LastChild.Data
		reply := util.ReplyTo(message, alertText, "")
		_, err := ctx.TG.Send(reply)
		return err
	}

	audioURL := ttsAddress + audioPath

	audio, err := client.Get(audioURL)
	if err != nil {
		return err
	}
	defer audio.Body.Close()

	vc := tgbotapi.NewVoiceUpload(message.Chat.ID, tgbotapi.FileReader{Name: "tts.ogg", Reader: audio.Body, Size: -1})
	vc.BaseFile.BaseChat.ReplyToMessageID = message.MessageID
	_, err = ctx.TG.Send(vc)
	return err
}

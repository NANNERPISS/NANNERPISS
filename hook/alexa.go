package hook

import (
	"os/exec"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func init() {
	Register("alexa", Alexa)
}

const cmdPrefix = "alexa, play "

func Alexa(ctx *context.Context, message *tgbotapi.Message) error {
	if !strings.HasPrefix(strings.ToLower(message.Text), cmdPrefix) {
		return nil
	}

	ytdlCmd := exec.Command("youtube-dl",
		"-f", "bestaudio[filesize<=10485760]",
		"-o", "-",
		"ytsearch:"+message.Text[len(cmdPrefix):])

	ytdlOut, err := ytdlCmd.StdoutPipe()
	if err != nil {
		return err
	}

	ffCmd := exec.Command("ffmpeg",
		"-i", "-",
		"-c:a", "libopus",
		"-b:a", "48k",
		"-f", "ogg",
		"-",
	)

	ffCmd.Stdin = ytdlOut
	ffOut, err := ffCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err = ytdlCmd.Start(); err != nil {
		return err
	}
	defer ytdlCmd.Wait()

	if err = ffCmd.Start(); err != nil {
		return err
	}
	defer ffCmd.Wait()

	vc := tgbotapi.NewVoiceUpload(message.Chat.ID, tgbotapi.FileReader{Name: "alexa.ogg", Reader: ffOut, Size: -1})
	vc.BaseFile.BaseChat.ReplyToMessageID = message.MessageID
	_, err = ctx.TG.Send(vc)

	if err != nil {
		reply := util.ReplyTo(message, "no", "")
		ctx.TG.Send(reply)
	}

	return err
}

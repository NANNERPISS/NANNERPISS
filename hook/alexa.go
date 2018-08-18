package hook

import (
	"os/exec"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func init() {
	Register("alexa", Alexa)
}

func Alexa(ctx *context.Context, message *tgbotapi.Message) error {
	prefix := "alexa, play "
	if !strings.HasPrefix(strings.ToLower(message.Text), prefix) {
		return nil
	}

	ytdlCmd := exec.Command("youtube-dl",
		"-f", "bestaudio[filesize<=5242880]",
		"-o", "-",
		"ytsearch:"+message.Text[len(prefix):])

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
	if err = ffCmd.Start(); err != nil {
		return err
	}

	vc := tgbotapi.NewVoiceUpload(message.Chat.ID, tgbotapi.FileReader{Name: "alexa.ogg", Reader: ffOut, Size: -1})
	vc.BaseFile.BaseChat.ReplyToMessageID = message.MessageID
	_, err = ctx.TG.Send(vc)
	return err
}

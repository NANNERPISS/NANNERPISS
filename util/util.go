package util

import (
	"bytes"
	"html/template"

	"gopkg.in/telegram-bot-api.v4"
)

const userFormat = `<a href="tg://user?id={{.ID}}">{{.FirstName}}</a>`

var userFormatTemplate = template.Must(template.New("user").Parse(userFormat))

func FormatUser(user *tgbotapi.User) (string, error) {
	response := &bytes.Buffer{}
	err := userFormatTemplate.Execute(response, user)
	if err != nil {
		return "", err
	}

	return response.String(), nil
}

func ReplyTo(message *tgbotapi.Message, response string) *tgbotapi.MessageConfig {
	reply := tgbotapi.NewMessage(message.Chat.ID, response)
	reply.ParseMode = "HTML"
	reply.ReplyToMessageID = message.MessageID
	return &reply
}

func GetSender(tg *tgbotapi.BotAPI, message *tgbotapi.Message) (tgbotapi.ChatMember, error) {
	senderChatConfig := tgbotapi.ChatConfigWithUser{}
	senderChatConfig.ChatID = message.Chat.ID
	senderChatConfig.UserID = message.From.ID

	return tg.GetChatMember(senderChatConfig)
}

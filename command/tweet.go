package command

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
	"github.com/ChimeraCoder/anaconda"
)

func init() {
	Register("tweet", Tweet)
}

func Tweet(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID == ctx.Config.TW.ControlGroup {
		sender, err := util.GetSender(ctx.TG, message)
		if err != nil {
			return err
		}

		if sender.IsAdministrator() || sender.IsCreator() {
			status := message.CommandArguments()
			var media *anaconda.Media
			
			if message.Photo != nil && len(*message.Photo)!= 0 {
				fileID := (*message.Photo)[len(*message.Photo)-1].FileID
				
				downloadURL, err := ctx.TG.GetFileDirectURL(fileID)
				if err != nil {
					return err
				}
				
				client := &http.Client{}
				client.Timeout = time.Second * 30
				
				req, err := http.NewRequest("GET", downloadURL, nil)
				if err != nil {
					return err
				}
				
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				
				img, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				
				imgb64 := base64.StdEncoding.EncodeToString(img)
				
				m, err := ctx.TW.UploadMedia(imgb64)
				if err != nil {
					return err
				}
				
				media = &m
			}
			
			if media == nil && status == "" {
				reply := util.ReplyTo(message, "Please include a message to tweet", "")
				_, err = ctx.TG.Send(reply)
				return err
			}
			
			values := url.Values{}
			
			if media != nil {
				values.Add("media_ids", media.MediaIDString)
			}
			
			values.Add("lat", "61.1940413")
			values.Add("long", "-149.8775202")
			values.Add("display_coordinates", "true")
			
			t, err := ctx.TW.PostTweet(status, values)
			if err != nil {
				return err
			}
			
			reply := util.ReplyTo(message, "https://twitter.com/NANNERPISS/status/" + t.IdStr, "")
			_, err = ctx.TG.Send(reply)
			return err
		}
	}

	return nil
}

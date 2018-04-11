package command

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
	"github.com/ChimeraCoder/anaconda"
)

func init() {
	Register("tweet", Admin(Tweet))
}

func getTweetID(urlStr string) (string, error) {
	replyURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if replyURL.Host == "twitter.com" && len(replyURL.Path) > 8 {
		splitURL := strings.Split(replyURL.Path, "/")
		if splitURL[2] == "status" {
			return splitURL[3], nil
		}
	}
	
	return "", fmt.Errorf("Couldn't extract ID from link")
}

func Tweet(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID != ctx.Config.TW.ControlGroup {
		return nil
	}
	
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
	
	var replyID string
	if v := message.CommandWithAt(); strings.Contains(v, "@") {
		replyID = strings.SplitN(v, "@", 2)[1]
	} else {
		if message.ReplyToMessage != nil && message.ReplyToMessage.Entities != nil {
			var urlString string
			for _, e := range *message.ReplyToMessage.Entities {
				if e.Type == "url" {
					urlString = message.ReplyToMessage.Text[e.Offset:e.Offset+e.Length]
				}
			}
			
			if urlString != "" {
				var err error
				replyID, err = getTweetID(urlString)
				if err != nil {
					return err
				}
			}
		}
	}
	
	if media == nil && status == "" {
		reply := util.ReplyTo(message, "Please include a message to tweet", "")
		_, err := ctx.TG.Send(reply)
		return err
	}
	
	values := url.Values{}
	
	if media != nil {
		values.Add("media_ids", media.MediaIDString)
	}
	
	if replyID != "" {
		values.Add("in_reply_to_status_id", replyID)
		values.Add("auto_populate_reply_metadata", "true")
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

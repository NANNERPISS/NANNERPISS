package command

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"strconv"
	"strings"

	"github.com/NANNERPISS/NANNERPISS/context"
	"github.com/NANNERPISS/NANNERPISS/util"

	"gopkg.in/telegram-bot-api.v4"
	"github.com/ChimeraCoder/anaconda"
)

func init() {
	Register("tweet", Admin(Tweet))
	Register("retweet", Admin(ReTweet))
}

func getTweetID(urlStr string) string {
	replyURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	if replyURL.Host == "twitter.com" && len(replyURL.Path) > 8 {
		splitURL := strings.Split(replyURL.Path, "/")
		if splitURL[2] == "status" {
			return splitURL[3]
		}
	}
	
	return ""
}

func extractURL(message *tgbotapi.Message) string {
	if message.Entities == nil {
		return ""
	}
	
	for _, e := range *message.Entities {
		if e.Type == "url" {
			return message.Text[e.Offset:e.Offset+e.Length]
		}
	}
	
	return ""
}

func ReTweet(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID != ctx.Config.TW.ControlGroup {
		return nil
	}
	
	var retweetID string
	if v := message.CommandWithAt(); strings.Contains(v, "@") {
		retweetID = strings.SplitN(v, "@", 2)[1]
	} else if message.ReplyToMessage != nil {
		urlString := extractURL(message.ReplyToMessage)
		
		if urlString != "" {
			retweetID = getTweetID(urlString)
		}
	} else if message.CommandArguments() != "" {
		urlString := extractURL(message)
		
		if urlString != "" {
			retweetID = getTweetID(urlString)
		}
	}
	
	if retweetID == "" {
		reply := util.ReplyTo(message, "Please include the tweet you want to reweet by using <code>/retweet@&lt;tweetID&gt;</code> or replying to a mesage with a twitter link with <code>/retweet</code>", "html")
		_, err := ctx.TG.Send(reply)
		return err
	}
	
	retweetIDInt64, err := strconv.ParseInt(retweetID, 10, 64)
	if err != nil {
		return err
	}
	
	t, err := ctx.TW.Retweet(retweetIDInt64, false)
	if err != nil {
		return err
	}
	
	reply := util.ReplyTo(message, "https://twitter.com/" + t.User.ScreenName + "/status/" + t.IdStr, "")
	_, err = ctx.TG.Send(reply)
	return err
}

func Tweet(ctx *context.Context, message *tgbotapi.Message) error {
	if message.Chat.ID != ctx.Config.TW.ControlGroup {
		return nil
	}
	
	status := message.CommandArguments()
	var media *anaconda.Media
	
	var photo *[]tgbotapi.PhotoSize
	if message.Photo != nil {
		photo = message.Photo
	} else if message.ReplyToMessage != nil && message.ReplyToMessage.Photo != nil {
		photo = message.ReplyToMessage.Photo
	}
	
	if photo != nil && len(*photo)!= 0 {
		fileID := (*photo)[len(*photo)-1].FileID
		
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
				replyID = getTweetID(urlString)
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
	
	reply := util.ReplyTo(message, "https://twitter.com/" + t.User.ScreenName + "/status/" + t.IdStr, "")
	_, err = ctx.TG.Send(reply)
	return err
}

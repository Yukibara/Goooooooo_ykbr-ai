package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/ikawaha/kagome/tokenizer"
)

func getAPI() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))
	return anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))
}

func kagomeInit() {
	tokenizer.SysDic()
}

var dict = make(map[string][]string)

func kagomeParse(str string) {
	t := tokenizer.New()
	// 辞書で単語毎に次の文字を管理するとよさそう
	morphs := t.Tokenize(str)
	KeyEOS := "\n"

	// BGNはbeginでENDはend
	// メモしないと忘れるので
	next := ""
	for i := range morphs {
		if i+1 < len(morphs) {
			next = morphs[i+1].Surface
		} else {
			next = "END"
		}
		if strings.Contains(morphs[i].Surface, KeyEOS) {
			if dict["BGN"] == nil {
				dict["BGN"] = make([]string, 0)
			}
			dict["BGN"] = append(dict["BGN"], next)
			continue
		}
		if strings.Contains(next, KeyEOS) {
			next = "END"
		}
		if dict[morphs[i].Surface] == nil {
			dict[morphs[i].Surface] = make([]string, 0)
		}
		dict[morphs[i].Surface] = append(dict[morphs[i].Surface], next)
	}
}

func main() {
	api := getAPI()
	v := url.Values{}
	v.Set("screen_name", "ykbr_")
	v.Add("count", "200")

	searchRes, _ := api.GetUserTimeline(v)
	resText := ""
	for i, tweet := range searchRes {
		if strings.HasPrefix(tweet.Text, "RT") || strings.HasPrefix(tweet.Text, "@") {
			searchRes = append(searchRes[:i], searchRes[i+1:]...)
		} else {
			kagomeParse(tweet.FullText)
		}
	}
	fmt.Println(resText)
}

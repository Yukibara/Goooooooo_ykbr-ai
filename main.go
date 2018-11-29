package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/ikawaha/kagome/tokenizer"
)

func getAPI() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))
	return anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))
}

var dict = make(map[string][]string)

func kagomeParse(str string) {
	udic, err := tokenizer.NewUserDic("./userdic.txt")
	if err != nil {
		panic(err)
	}
	t := tokenizer.New()
	t.SetUserDic(udic)
	// 辞書で単語毎に次の文字を管理するとよさそう
	morphs := t.Tokenize(str)
	KeyEOS := "\n"

	// BEGINはbeginでENDはend
	// メモしないと忘れるので
	next := ""
	for i := range morphs {
		if i+1 < len(morphs) {
			next = morphs[i+1].Surface
		} else {
			next = "END"
		}
		if strings.Contains(morphs[i].Surface, KeyEOS) {
			if dict["BEGIN"] == nil {
				dict["BEGIN"] = make([]string, 0)
			}
			dict["BEGIN"] = append(dict["BEGIN"], next)
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

func genWord() string {
	timee := int64(time.Now().UnixNano())
	rand.Seed(timee)
	res := ""
	temp := dict["BEGIN"][rand.Intn(len(dict["BEGIN"]))]

	for {
		if temp == "END" {
			break
		}
		res += temp
		if len(dict[temp]) > 0 {
			temp = dict[temp][rand.Intn(len(dict[temp]))]
		}
	}
	return strings.Replace(res, "EOS", "", -1)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func makeGo7Go(num int) string {
	res := ""
	temp := dict["BEGIN"][random(0, len(dict["BEGIN"]))]

	for {
		if temp == "END" {
			break
		} else if len(res) == num {
			break
		}
		res += temp
		if len(dict[temp]) > 0 {
			temp = dict[temp][random(0, len(dict[temp]))]
		}
	}

	return strings.Replace(res, "EOS", "", -1)
}

func go7goFunc() string {
	res := ""

	temp := makeGo7Go(5)
	res += temp + "\n"

	temp = makeGo7Go(7)
	res += temp + "\n"

	temp = makeGo7Go(5)
	res += temp

	return strings.Replace(res, "EOS", "", -1)
}

func main() {
	api := getAPI()
	v := url.Values{}
	v.Set("screen_name", "ykbr_")
	v.Add("count", "200")
	vre := url.Values{}
	vre.Set("track", "\"@ykbr__ai 575\"")
	twitterStream := api.PublicStreamFilter(vre)

	searchRes, _ := api.GetUserTimeline(v)
	for i, tweet := range searchRes {
		if strings.HasPrefix(tweet.Text, "RT") || strings.HasPrefix(tweet.Text, "@") {
			searchRes = append(searchRes[:i], searchRes[i+1:]...)
		} else {
			kagomeParse(tweet.FullText)
		}
	}

	//マルコフ射出
	// res := genWord()
	// fmt.Println(res)
	// api.PostTweet(res, nil)

	// postStr := go7goFunc()
	// fmt.Println(postStr)

	for {
		x := <-twitterStream.C
		switch tw := x.(type) {
		case anaconda.Tweet:
			searchRes, _ := api.GetUserTimeline(vre)
			for i, tweet := range searchRes {
				if strings.HasPrefix(tweet.Text, "RT") || strings.HasPrefix(tweet.Text, "@") {
					searchRes = append(searchRes[:i], searchRes[i+1:]...)
				} else {
					kagomeParse(tweet.FullText)
				}
			}
			fmt.Println("Success")
			v2 := url.Values{}
			v2.Add("in_reply_to_status_id", tw.IdStr)
			postStr := "@"
			postStr += tw.User.ScreenName
			postStr += " ここで一句:\n"
			postStr += go7goFunc()
			fmt.Println(postStr)
			api.PostTweet(postStr, v)
		default:
		}
	}

}

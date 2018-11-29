package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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

// 内部で使う用のやつ
func makeGo7Go(num int, word string) string {
	res := ""
	// たまに次につながる単語が無くて落ちるので虚無をする
	if dict[word] == nil {
		dict[word] = make([]string, 0)
		dict[word] = append(dict[word], "虚無")
	}
	temp := dict[word][random(0, len(dict[word]))]
	for {
		if temp == "END" {
			break
		} else if len(res) >= num {
			break
		}
		res += temp
		if len(dict[temp]) > 0 {
			temp = dict[temp][random(0, len(dict[temp]))]
		}
	}

	return strings.Replace(res, "EOS", "", -1)
}

const five = 15
const seven = 21

// ここから呼び出すmakeGo7Goで単語を生成
func partsGo7Go(num int, worded string) (string, string) {
	re := regexp.MustCompile(`(\p{Han}|\p{Katakana}|\p{Hiragana})*`)
	res := ""
	lastWord := ""

	// 読みで575するための辞書
	udic, err := tokenizer.NewUserDic("./userdic.txt")
	if err != nil {
		panic(err)
	}
	t := tokenizer.New()
	t.SetUserDic(udic)
	for {
		temp := makeGo7Go(num, worded)
		if len(temp) == num {
			a := strings.Join(re.FindAllString(temp, -1), "")
			count := 0
			morphs := t.Tokenize(a)
			for _, m := range morphs {
				features := m.Features()
				b := len(features)
				// 中身があれば文字数に加算
				// 存在を確認しましょう 例外で落ちるので(1敗)
				if b >= 8 {
					// 読みがカタカナで格納されているところ
					c := (features[7])
					// 音数のカウント
					count += utf8.RuneCountInString(c)
					lastWord = features[6]
				}
			}
			if count == num/3 {
				res += temp + "\n"
				break
			}
			continue
		}
	}
	return res, lastWord
}

// これ→partsofGo7Go→makeGo7Go
func go7goFunc() string {
	res := ""
	tmp := ""
	lastword := ""

	tmp, lastword = partsGo7Go(five, "BEGIN")
	res += tmp
	tmp, lastword = partsGo7Go(seven, lastword)
	res += tmp
	tmp, _ = partsGo7Go(five, lastword)
	res += tmp

	return strings.Replace(res, "EOS", "", -1)
}

func main() {
	api := getAPI()
	v := url.Values{}
	v.Set("screen_name", "ykbr_")
	v.Add("count", "200")
	vre := url.Values{}
	vre.Set("track", "@ykbr__ai 575")

	searchRes, _ := api.GetUserTimeline(v)
	for _, tweet := range searchRes {
		if !(strings.HasPrefix(tweet.Text, "RT") || strings.HasPrefix(tweet.Text, "@")) {
			if strings.Contains(tweet.FullText, "http://") || strings.Contains(tweet.FullText, "https://") {
				tweet.FullText = strings.Split(tweet.FullText, "http")[0]
			}
			kagomeParse(tweet.FullText)
		}
	}

	//575テスト
	postStr := go7goFunc()
	fmt.Println(postStr)

	// 定期ツイート
	timeTw := time.NewTicker(1 * time.Hour)

	// PublicStreamの監視
	twitterStream := api.PublicStreamFilter(vre)
	for {
		fmt.Println("Listening...")
		select {
		// マルコフが出る
		case <-timeTw.C:
			fmt.Println("【定期】マルコフ連鎖実行")
			res := genWord()
			fmt.Println(res)
			api.PostTweet(res, nil)
		// 575が出る
		case x := <-twitterStream.C:
			switch tw := x.(type) {
			case anaconda.Tweet:
				searchRes, _ := api.GetUserTimeline(vre)
				for _, tweet := range searchRes {
					if !(strings.HasPrefix(tweet.Text, "RT") || strings.HasPrefix(tweet.Text, "@")) {
						if strings.Contains(tweet.FullText, "http://") || strings.Contains(tweet.FullText, "https://") {
							tweet.FullText = strings.Split(tweet.FullText, "http")[0]
						}
						kagomeParse(tweet.FullText)
					}
				}
				fmt.Println("Catch!")
				v2 := url.Values{}
				v2.Add("in_reply_to_status_id", tw.IdStr)
				postStr := "@"
				postStr += tw.User.ScreenName
				postStr += " ここで一句:\n"
				postStr += go7goFunc()
				fmt.Println(postStr)
				api.PostTweet(postStr, v2)
			default:
			}
		}
	}

}

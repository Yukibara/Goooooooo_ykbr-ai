package main

import (
	"net/url"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

func getAPI() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))
	return anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))
}

func deleteSlice(strings []string, search string) []string {
	res := []string{}
	for _, v := range strings {
		if v != search {
			res = append(res, v)
		}
	}
	return res
}

func main() {
	api := getAPI()
	v := url.Values{}
	v.Set("screen_name", "ykbr_")
	v.Add("count", "200")

	searchRes, _ := api.GetUserTimeline(v)
	for i, tweet := range searchRes {
		if strings.HasPrefix(tweet.Text, "RT") || strings.HasPrefix(tweet.Text, "@") {
			searchRes = append(searchRes[:i], searchRes[i+1:]...)
		}
	}

}

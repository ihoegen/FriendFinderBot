// Copyright 2017 Google Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to writing, software distributed
// under the License is distributed on a "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/url"
	"os"
	"strings"

	"../postAnalysis"
	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
)

var (
	consumerKey       = "3NNsJq3o7CTo0SdxnakaaZnUa"
	consumerSecret    = "1cQLmVhSVqToEd38dZqoo7yj1TPjD9c0IDiajED4kBJ2MG8Kxl"
	accessToken       = "922005253666488321-SAV07T486AJjuHoyKEdLZv3uuRs0sSM"
	accessTokenSecret = "3gHxViZyXQzgDGc8XzlFU5lw4upGPqTCcgqCGZg4PXk61"
)

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}

func main() {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)

	log := &logger{logrus.New()}
	api.SetLogger(log)

	stream := api.PublicStreamFilter(url.Values{
		"track": []string{"@FriendFinderBot"},
	})

	defer stream.Stop()

	for v := range stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			log.Warningf("received unexpected value of type %T", v)
			continue
		}
		log.Info(t.User.ScreenName)
		userTweets, err := api.GetUserTimeline(url.Values{
			"screen_name": []string{t.User.ScreenName},
		})
		if err != nil {
			log.Errorf("Could not get tweets for %v: %v", t.User.ScreenName, err)
		}
		tweets := make(map[string]int)
		var retweeter []anaconda.User
		for _, v := range userTweets {
			if !v.Retweeted {
				spliced := strings.Split(strings.ToLower(v.Text), " ")
				for _, str := range spliced {
					tweets[str] = tweets[str] + 1
				}
			} else {
				rtls, err := api.GetRetweets(v.Id, nil)
				if err != nil {
					log.Errorf("Could not get tweets for %v: %v", v.Id, err)
				}
				for _, rt := range rtls {
					retweeter = append(retweeter, rt.User)
					log.Info("User ", rt.User.ScreenName)
					retweeterTweets, _ := api.GetUserTimeline(url.Values{"screen_id": []string{rt.User.ScreenName}})
					retweeterMap := postAnalysis.WordCount(retweeterTweets)
					relationship := postAnalysis.FindMatches(tweets, retweeterMap)
					if relationship > 0.8 {
						print("Yay")
					}
				}
			}
		}
	}
}

type logger struct {
	*logrus.Logger
}

func (log *logger) Critical(args ...interface{})                 { log.Error(args...) }
func (log *logger) Criticalf(format string, args ...interface{}) { log.Errorf(format, args...) }
func (log *logger) Notice(args ...interface{})                   { log.Info(args...) }
func (log *logger) Noticef(format string, args ...interface{})   { log.Infof(format, args...) }

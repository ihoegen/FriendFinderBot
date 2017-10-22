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
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
)

var (
	consumerKey       = getenv("TWITTER_CONSUMER_KEY")
	consumerSecret    = getenv("TWITTER_CONSUMER_SECRET")
	accessToken       = getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret = getenv("TWITTER_ACCESS_TOKEN_SECRET")
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
		message := "Thank you for the mention @" + t.User.ScreenName
		_, err := api.PostTweet(message, url.Values{})
		if err != nil {
			log.Errorf("could not post %v: %v", message, err)
			continue
		}
		log.Infof("posted %v", message)
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
				}
			}
		}
	}
}

type logger struct {
	*logrus.Logger
}

func FindMatches(userKeywords map[string]int, PotentialMatches map[string]int) float64 {
	keys := make([]string, 0, len(userKeywords))
	userTotal := 0
	userAverage := float64(userTotal) / float64(len(userKeywords))
	matchTotal := 0
	matchAverage := float64(matchTotal) / float64(len(PotentialMatches))
	for k := range userKeywords {
		keys = append(keys, k)
		userTotal += userKeywords[k]
	}
	for k := range PotentialMatches {
		matchTotal += PotentialMatches[k]
	}
	topSum := 0.0
	bottomX := 0.0
	bottomY := 0.0
	for _, key := range keys {
		topSum += (float64(userKeywords[key]) - userAverage) * (float64(PotentialMatches[key]) - matchAverage)
		bottomX += math.Pow((float64(userKeywords[key]) - userAverage), 2)
		bottomY += math.Pow((float64(PotentialMatches[key]) - matchAverage), 2)
	}
	return (topSum / (math.Sqrt(bottomX) * math.Sqrt(bottomY)))
}
func (log *logger) Critical(args ...interface{})                 { log.Error(args...) }
func (log *logger) Criticalf(format string, args ...interface{}) { log.Errorf(format, args...) }
func (log *logger) Notice(args ...interface{})                   { log.Info(args...) }
func (log *logger) Noticef(format string, args ...interface{})   { log.Infof(format, args...) }

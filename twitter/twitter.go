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
	"regexp"
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
			"screen_name": []string{t.User.ScreenName}, "count": []string{"200"},
		})
		print(userTweets)
		print(err)
		if err != nil {
			log.Errorf("Could not get tweets for %v: %v", t.User.ScreenName, err)
		}
		friends, _ := api.GetFriendsList(url.Values{
			"screen_name": []string{t.User.ScreenName}})
		print(friends.Users)
		tweets := make(map[string]int)
		for _, v := range userTweets {
			spliced := strings.Split(strings.ToLower(v.Text), " ")
			for _, str := range spliced {
				regex, err := regexp.Compile("[^a-zA-z]+")
				spaceregex, err := regexp.Compile("[\\s]+")
				cleaned := regex.ReplaceAllString(str, "")
				cleaned = spaceregex.ReplaceAllString(cleaned, "")
				if err != nil {
					log.Error("Regex issue: ", err)
				}
				tweets[cleaned]++
			}
		}
		var mostUsed string
		var mostUsedCount int = 0
		for word, count := range tweets {
			if count > mostUsedCount {
				if postAnalysis.InSlice(word, postAnalysis.CommonWords) {
					continue
				}
				mostUsed = word
				mostUsedCount = count
			}
		}
		var otherUser []string
		userStream := api.PublicStreamFilter(url.Values{
			"track": []string{mostUsed},
		})
		var tweetCount int
		defer userStream.Stop()
		for ot := range userStream.C {
			t2, ok := ot.(anaconda.Tweet)
			if !ok {
				log.Warningf("received unexpected value of type %T", v)
				continue
			}
			if tweetCount > 50 {
				userStream.Stop()
			}
			otherUser = append(otherUser, t2.User.ScreenName)
			tweetCount++
		}
		defer stream.Stop()
		log.Info("Retweets: ", otherUser)
		log.Info("User Heatmap: ", tweets)
		var topMatchUser string
		var topMatchPercent float64
		for _, user := range otherUser {
			otherUserTweets, _ := api.GetUserTimeline(url.Values{"screen_name": []string{user}, "count": []string{"200"}})
			otherUserMap := postAnalysis.WordCount(otherUserTweets)
			log.Info("RT heatmap: ", otherUserMap)
			log.Info("Looking at " + user)
			relationship := postAnalysis.FindMatches(tweets, otherUserMap)
			log.Info("User ", user, " has a relationship of ", relationship)
			if relationship > topMatchPercent {
				topMatchPercent = relationship
				topMatchUser = user
			}
			if relationship > 0.8 {
				log.Info("Friend")
				message := "@" + t.User.ScreenName + ", we would like to introduce you to " + user
				log.Info("Message sent ", message)
				api.PostTweet(message, nil)
				break
			}
		}
		log.Infof("Best match %d", topMatchPercent)
		message := "@" + t.User.ScreenName + ", we would like to introduce you to " + topMatchUser
		log.Info("Message sent ", message)
		api.PostTweet(message, nil)
	}
}

func inSlice(fr anaconda.User, c []anaconda.User) bool {
	for _, b := range c {
		if b.ScreenName == fr.ScreenName {
			return true
		}
	}
	return false
}

func inSlice2(fr string, c []string) bool {
	for _, b := range c {
		if fr == b {
			return true
		}
	}
	return false
}

type logger struct {
	*logrus.Logger
}

func (log *logger) Critical(args ...interface{})                 { log.Error(args...) }
func (log *logger) Criticalf(format string, args ...interface{}) { log.Errorf(format, args...) }
func (log *logger) Notice(args ...interface{})                   { log.Info(args...) }
func (log *logger) Noticef(format string, args ...interface{})   { log.Infof(format, args...) }

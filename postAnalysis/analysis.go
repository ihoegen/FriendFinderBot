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

package postAnalysis

import (
	"math"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

func WordCount(userTweets []anaconda.Tweet) (words map[string]int) {
	words = make(map[string]int)
	for _, v := range userTweets {
		spliced := strings.Split(strings.ToLower(v.Text), " ")
		for _, str := range spliced {
			words[str] = words[str] + 1
		}
	}
	return
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
	coef := (topSum / (math.Sqrt(bottomX) * math.Sqrt(bottomY)))
	return coef
}

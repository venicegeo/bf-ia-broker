// Copyright 2017, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package planet

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

// BestSceneInput contains the inputs for the BestScene function
type BestSceneInput struct {
	ItemType string
	Point    *geojson.Point
	Tides    bool
}

// BestScene returns the best scene based on age, cloud cover, and tides
func BestScene(options SearchOptions, context *Context) (string, error) {
	var (
		result    string
		err       error
		scenes    *geojson.FeatureCollection
		bestScore float64
		currScore float64
	)
	if scenes, err = GetScenes(options, context); err != nil {
		return result, err
	}
	log.Print("XXXXXXXX")
	log.Print(scenes)
	for _, scene := range scenes.Features {
		if result == "" {
			result = scene.IDStr()
			bestScore = scoreScene(scene, context)
		} else {
			currScore = scoreScene(scene, context)
			if currScore > bestScore {
				result = scene.IDStr()
				bestScore = currScore
			}
		}
	}
	return result, nil
}

var date2015 = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

func scoreScene(scene *geojson.Feature, context util.LogContext) float64 {
	var (
		result       = 1.0
		acquiredDate time.Time
		err          error
	)
	cloudCover := scene.PropertyFloat("cloudCover")
	currTide := scene.PropertyFloat("CurrentTide")
	minTide := scene.PropertyFloat("24hrMinTide")
	maxTide := scene.PropertyFloat("24hrMaxTide")
	acquiredDateString := scene.PropertyString("acquiredDate")
	if acquiredDate, err = time.Parse(model.PlanetTimeFormat, acquiredDateString); err != nil {
		util.LogInfo(context, fmt.Sprintf("Received invalid date of %v: ", acquiredDateString))
		return 0.0
	}
	// Older scenes are unlikely to be in the archive
	// unless they happen to have very good cloud cover so discourage them
	acquiredDateUnix := acquiredDate.Unix()
	if acquiredDateUnix < date2015 {
		result = 0.5
	}
	now := time.Now().Unix()
	result -= math.Sqrt(cloudCover / 100.0)
	result -= float64(acquiredDateUnix-now) / (60.0 * 60.0 * 24.0 * 365.0 * 10.0)
	if math.IsNaN(currTide) {
		// If no tide is available for some reason, assume low tide
		// log.Printf("No tide available for %v", scene.ID)
		result -= math.Sqrt(0.1)
	} else {
		result -= math.Sqrt(0.1) * (maxTide - currTide) / (maxTide - minTide)
	}
	return result
}

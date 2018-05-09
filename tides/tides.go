// Copyright 2016, RadiantBlue Technologies, Inc.
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

package tides

import (
	"fmt"
	"time"

	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

func toTideIn(bbox geojson.BoundingBox, timeStr string) *InputLocation {
	var (
		center  *geojson.Point
		dtgTime time.Time
		err     error
	)
	if center = bbox.Centroid(); center == nil {
		return nil
	}
	if dtgTime, err = time.Parse("2006-01-02T15:04:05Z", timeStr); err != nil {
		return nil
	}
	return &InputLocation{Lat: center.Coordinates[1], Lon: center.Coordinates[0], Dtg: dtgTime.Format("2006-01-02-15-04")}
}

func toTidesIn(features []*geojson.Feature, context util.LogContext) (result Input, dtgFeatureMap map[string]*geojson.Feature) {
	dtgFeatureMap = make(map[string]*geojson.Feature)
	for _, feature := range features {
		currTideIn := toTideIn(feature.ForceBbox(), feature.PropertyString("acquiredDate"))
		if currTideIn == nil {
			util.LogInfo(context, fmt.Sprintf("Could not get tide information from feature %v because required elements did not exist. BBOX: %#v, Date: %v",
				feature.IDStr(),
				feature.ForceBbox(),
				feature.PropertyString("acquiredDate")))
			continue
		}
		result.Locations = append(result.Locations, *currTideIn)
		dtgFeatureMap[currTideIn.Dtg] = feature
	}
	return
}

// GetTides returns the tide information for the features provided.
// Features must have a geometry and an acquiredDate property.
func GetTides(fc *geojson.FeatureCollection, context *Context) (*geojson.FeatureCollection, error) {
	var (
		err    error
		tout   Output
		result *geojson.FeatureCollection
	)
	tidesURL := context.TidesURL
	tin, dtgFeatureMap := toTidesIn(fc.Features, context)

	util.LogAudit(context, util.LogAuditInput{Actor: "anon user", Action: "POST", Actee: tidesURL, Message: "Requesting tide information", Severity: util.INFO})
	if _, err = util.ReqByObjJSON("POST", tidesURL, "", tin, &tout); err != nil {
		return nil, err
	}
	util.LogAudit(context, util.LogAuditInput{Actor: tidesURL, Action: "POST response", Actee: "anon user", Message: "Retrieving tide information", Severity: util.INFO})

	tideFeatures := []*geojson.Feature{}
	for _, outputLocation := range tout.Locations {
		if _, ok := dtgFeatureMap[outputLocation.Dtg]; !ok {
			util.LogInfo(context, "Failed to find location for output dtg "+outputLocation.Dtg)
			continue
		}
		oldFeaturePtr, _ := dtgFeatureMap[outputLocation.Dtg]
		newFeature := *oldFeaturePtr
		newFeature.Properties["CurrentTide"] = outputLocation.Results.CurrTide
		newFeature.Properties["MinimumTide24Hours"] = outputLocation.Results.MinTide
		newFeature.Properties["MaximumTide24Hours"] = outputLocation.Results.MaxTide
		tideFeatures = append(tideFeatures, &newFeature)
	}

	result = geojson.NewFeatureCollection(tideFeatures)

	return result, err
}

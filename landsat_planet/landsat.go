// Copyright 2018, RadiantBlue Technologies, Inc.
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

package landsat

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/venicegeo/bf-ia-broker/util"
)

// Old LandSat IDs come back in the form LC80060522017107LGN00

var landSatSceneIDPattern = regexp.MustCompile("LC8([0-9]{3})([0-9]{3}).*")

// IsValidLandSatID returns whether an ID is a valid LandSat ID
func IsValidLandSatID(sceneID string) bool {
	return landSatSceneIDPattern.MatchString(sceneID)
}

const preCollectionLandSatAWSURL = "%s/L8/%s/%s/%s/%s"

func formatPreCollectionIDToURL(sceneID string) string {
	m := landSatSceneIDPattern.FindStringSubmatch(sceneID)[1:]
	return fmt.Sprintf(preCollectionLandSatAWSURL, util.GetLandsatHost(), m[0], m[1], sceneID, "")
}

var preCollectionDataTypes = []string{"L1T", "L1GT", "L1G"}

// IsPreCollectionDataType returns whether a data type is a Pre-"Collection 1" type
// Reference: https://landsat.usgs.gov/landsat-processing-details
func IsPreCollectionDataType(dataType string) bool {
	for _, t := range preCollectionDataTypes {
		if dataType == t {
			return true
		}
	}
	return false
}

var collection1DataTypes = []string{"L1TP", "L1GT", "L1GS"}

// IsCollection1DataType returns whether a data type is a "Collection 1" type
// Reference: https://landsat.usgs.gov/landsat-processing-details
func IsCollection1DataType(dataType string) bool {
	dataType = strings.ToUpper(dataType)
	for _, t := range collection1DataTypes {
		if dataType == t {
			return true
		}
	}
	return false
}

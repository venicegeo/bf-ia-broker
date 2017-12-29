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

package planet

import (
	"errors"

	"github.com/venicegeo/bf-ia-broker/landsat"
)

var landSatBandsSuffixes = map[string]string{
	"coastal":      "_B1.TIF",
	"blue":         "_B2.TIF",
	"green":        "_B3.TIF",
	"red":          "_B4.TIF",
	"nir":          "_B5.TIF",
	"swir1":        "_B6.TIF",
	"swir2":        "_B7.TIF",
	"panchromatic": "_B8.TIF",
	"cirrus":       "_B9.TIF",
	"tirs1":        "_B10.TIF",
	"tirs2":        "_B11.TIF",
}

func addLandsatS3BandsToProperties(landSatID string, dataType string, properties *map[string]interface{}) error {
	if !landsat.IsValidLandSatID(landSatID) {
		return errors.New("Not a valid LandSat ID: " + landSatID)
	}

	awsFolder, prefix, err := landsat.GetSceneFolderURL(landSatID, dataType)
	if err != nil {
		return err
	}

	bands := make(map[string]string)
	for band, suffix := range landSatBandsSuffixes {
		bands[band] = awsFolder + prefix + suffix
	}
	(*properties)["bands"] = bands

	return nil
}

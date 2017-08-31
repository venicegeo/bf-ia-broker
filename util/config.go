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

package util

import (
	"os"
	"strconv"
)

const LANDSAT_HOST  = "LANDSAT_HOST"
const defaultLandSatHost = "https://landsat-pds.s3.amazonaws.com"

const SENTINEL_HOST = "SENTINEL_HOST"
const defaultSentinelHost = "https://sentinel-s2-l1c.s3.amazonaws.com"

const PL_API_URL = "PL_API_URL"
const defaultPlApiUrl = "https://api.planet.com"

const BF_TIDE_PREDICTION_URL = "BF_TIDE_PREDICTION_URL"
const defaultTidesUrl = "https://bf-tideprediction.int.geointservices.io/tides"

const PL_DISABLE_PERMISSIONS_CHECK = "PL_DISABLE_PERMISSIONS_CHECK"


func GetLandsatHost() string {
	landSatHost := os.Getenv(LANDSAT_HOST)
	if landSatHost == "" {
		LogAlert(&BasicLogContext{}, "Didn't get Landsat Host URL from the environment. Using default.")
		landSatHost = defaultLandSatHost
	}
	return landSatHost
}

func GetSentinelHost() string {
	sentinelHost := os.Getenv(SENTINEL_HOST)
	if sentinelHost == "" {
		LogAlert(&BasicLogContext{}, "Didn't get Sentinel Host URL from the environment. Using default.")
		sentinelHost = defaultSentinelHost
	}
	return sentinelHost
}

func GetPlanetLabsApiUrl() string {
	planetBaseURL := os.Getenv(PL_API_URL)
	if planetBaseURL == "" {
		LogAlert(&BasicLogContext{}, "Didn't get Planet Labs API URL from the environment. Using default.")
		planetBaseURL = defaultPlApiUrl
	}
	return planetBaseURL
}

func GetTidesUrl() string {
	tidesURL := os.Getenv(BF_TIDE_PREDICTION_URL)
	if tidesURL == "" {
		LogAlert(&BasicLogContext{}, "Didn't get Tide Prediction URL from the environment. Using default.")
		tidesURL = defaultTidesUrl
	}
	return tidesURL
}

func IsPlanetPermissionsDisabled() (bool, error) {
	return strconv.ParseBool(os.Getenv(PL_DISABLE_PERMISSIONS_CHECK))
}

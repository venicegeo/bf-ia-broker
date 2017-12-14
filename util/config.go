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
	"fmt"
	"os"
	"strconv"
)

const DOMAIN = "DOMAIN"

const LANDSAT_HOST = "LANDSAT_HOST"

const SENTINEL_HOST = "SENTINEL_HOST"

const PL_API_URL = "PL_API_URL"

const BF_TIDE_PREDICTION_URL = "BF_TIDE_PREDICTION_URL"
const defaultTidesUrl = "https://bf-tideprediction.int.geointservices.io/tides"

const PL_DISABLE_PERMISSIONS_CHECK = "PL_DISABLE_PERMISSIONS_CHECK"

func GetBeachfrontDomain() string {
	domain := os.Getenv(DOMAIN)
	if len(domain) == 0 {
		LogAlert(&BasicLogContext{}, "Didn't get domain from environment.")
	}
	return domain
}

func GetLandsatHost() string {
	landSatHost := os.Getenv(LANDSAT_HOST)
	if len(landSatHost) == 0 {
		LogAlert(&BasicLogContext{}, "Did not get Landsat Host URL from the environment. Landsat will not be available.")
	}
	return landSatHost
}

func GetSentinelHost() string {
	sentinelHost := os.Getenv(SENTINEL_HOST)
	if len(sentinelHost) == 0 {
		LogAlert(&BasicLogContext{}, "Didn't get Sentinel Host URL from the environment. Sentinel will not be available.")
	}
	return sentinelHost
}

func GetPlanetLabsApiUrl() string {
	planetBaseURL := os.Getenv(PL_API_URL)
	if len(planetBaseURL) == 0 {
		LogAlert(&BasicLogContext{}, "Didn't get Planet API URL from the environment. Planet API will not be available.")
	}
	return planetBaseURL
}

func GetTidesUrl() string {
	tidesURL := os.Getenv(BF_TIDE_PREDICTION_URL)
	if len(tidesURL) == 0 {
		LogInfo(&BasicLogContext{}, "Didn't get explicit Tide Prediction URL from the environment. Using implied URL based on domain.")
		domain := GetBeachfrontDomain()
		if domain != "" {
			tidesURL = fmt.Sprintf("https://bf-tideprediction.%s/tides", domain)
		} else {
			LogAlert(&BasicLogContext{}, "No domain in environment. Using default tides URL: "+defaultTidesUrl)
			tidesURL = defaultTidesUrl
		}
	}
	return tidesURL
}

func IsPlanetPermissionsDisabled() (bool, error) {
	return strconv.ParseBool(os.Getenv(PL_DISABLE_PERMISSIONS_CHECK))
}

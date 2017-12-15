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

// Environment variables
const (
	DOMAIN                       = "DOMAIN"
	LANDSAT_HOST                 = "LANDSAT_HOST"
	SENTINEL_HOST                = "SENTINEL_HOST"
	PL_API_URL                   = "PL_API_URL"
	BF_TIDE_PREDICTION_URL       = "BF_TIDE_PREDICTION_URL"
	PL_DISABLE_PERMISSIONS_CHECK = "PL_DISABLE_PERMISSIONS_CHECK"
)

const defaultTidesURL = "https://bf-tideprediction.int.geointservices.io/tides"

// GetBeachfrontDomain returns a string for the DOMAIN environment variable
func GetBeachfrontDomain() string {
	domain, ok := os.LookupEnv(DOMAIN)
	if !ok {
		LogAlert(&BasicLogContext{}, "Didn't get domain from environment.")
	}
	return domain
}

// GetLandsatHost returns a string for the LANDSAT_HOST environment variable
func GetLandsatHost() string {
	landSatHost, ok := os.LookupEnv(LANDSAT_HOST)
	if !ok {
		LogAlert(&BasicLogContext{}, "Did not get Landsat Host URL from the environment. Landsat will not be available.")
	}
	return landSatHost
}

// GetSentinelHost returns a string for the SENTINEL_HOST environment variable
func GetSentinelHost() string {
	sentinelHost, ok := os.LookupEnv(SENTINEL_HOST)
	if !ok {
		LogAlert(&BasicLogContext{}, "Did not get Sentinel Host URL from the environment. Sentinel will not be available.")
	}
	return sentinelHost
}

// GetPlanetAPIURL returns a string for the PL_API_URL environment variable
func GetPlanetAPIURL() string {
	planetBaseURL, ok := os.LookupEnv(PL_API_URL)
	if !ok {
		LogAlert(&BasicLogContext{}, "Did not get Planet API URL from the environment. Planet API will not be available.")
	}
	return planetBaseURL
}

// GetTidesURL returns a string for the BF_TIDE_PREDICTION_URL
// environment variable or generates one if needed
func GetTidesURL() string {
	tidesURL, ok := os.LookupEnv(BF_TIDE_PREDICTION_URL)
	if !ok {
		LogInfo(&BasicLogContext{}, "Did not get explicit Tide Prediction URL from the environment. Using implied URL based on domain.")
		domain := GetBeachfrontDomain()
		if len(domain) == 0 {
			LogAlert(&BasicLogContext{}, "No domain in environment. Using default tides URL: "+defaultTidesURL)
			tidesURL = defaultTidesURL
		} else {
			tidesURL = fmt.Sprintf("https://bf-tideprediction.%s/tides", domain)
		}
	}
	return tidesURL
}

// IsPlanetPermissionsDisabled returns true if the
// PL_DISABLE_PERMISSIONS_CHECK is true
func IsPlanetPermissionsDisabled() (bool, error) {
	return strconv.ParseBool(os.Getenv(PL_DISABLE_PERMISSIONS_CHECK))
}

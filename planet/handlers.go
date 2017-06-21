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

package planet

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

const noPlanetKey = "This operation requires a Planet Labs API key."
const noPlanetImageID = "This operation requires a Planet Labs image ID."
const invalidCloudCover = "Cloud Cover value of %v is invalid."

// DiscoverHandler is a handler for /planet/discover
// @Title planetDiscoverHandler
// @Description discovers scenes from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true         "Planet Labs Item Type, e.g., rapideye or planetscope"
// @Param   bbox            query   string  false        "The bounding box, as a GeoJSON Bounding box (x1,y1,x2,y2)"
// @Param   cloudCover      query   string  false        "The maximum cloud cover, as a percentage (0-100)"
// @Param   acquiredDate    query   string  false        "The minimum (earliest) acquired date, as RFC 3339"
// @Param   maxAcquiredDate query   string  false        "The maximum acquired date, as RFC 3339"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.FeatureCollection
// @Failure 400 {object}  string
// @Router /planet/discover/{itemType} [get]
type DiscoverHandler struct {
	Context Context
}

// NewDiscoverHandler creates a new handler using configuration
// from environment variables
func NewDiscoverHandler() DiscoverHandler {
	planetBaseURL := os.Getenv("PL_API_URL")
	if planetBaseURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Planet Labs API URL from the environment. Using default.")
		planetBaseURL = "https://api.planet.com"
	}

	tidesURL := os.Getenv("BF_TIDE_PREDICTION_URL")
	if tidesURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Tide Prediction URL from the environment. Using default.")
		tidesURL = "https://bf-tideprediction.int.geointservices.io/tides"
	}

	return DiscoverHandler{
		Context: Context{
			BasePlanetURL: planetBaseURL,
			BaseTidesURL:  tidesURL,
		},
	}
}

// ServeHTTP implements the http.Handler interface for the DiscoverHandler type
func (h DiscoverHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var (
		fc         *geojson.FeatureCollection
		err        error
		itemType   string
		bytes      []byte
		bbox       geojson.BoundingBox
		ccStr      string
		cloudCover float64
	)
	util.LogAudit(&h.Context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving /discover request", Severity: util.INFO})

	if util.Preflight(writer, request, &h.Context) {
		return
	}

	h.Context.PlanetKey = request.FormValue("PL_API_KEY")
	if h.Context.PlanetKey == "" {
		util.LogSimpleErr(&h.Context, noPlanetKey, nil)
		util.HTTPError(request, writer, &h.Context, noPlanetKey, http.StatusBadRequest)
		return
	}

	tides, _ := strconv.ParseBool(request.FormValue("tides"))

	ccStr = request.FormValue("cloudCover")
	if ccStr != "" {
		if cloudCover, err = strconv.ParseFloat(ccStr, 64); err != nil {
			message := fmt.Sprintf(invalidCloudCover, ccStr)
			util.LogSimpleErr(&h.Context, message, err)
			util.HTTPError(request, writer, &h.Context, message, http.StatusBadRequest)
			return
		}
		cloudCover = cloudCover / 100.0
	}

	itemType = mux.Vars(request)["itemType"]
	switch itemType {
	case "REOrthoTile", "rapideye":
		itemType = "REOrthoTile"
	case "PSOrthoTile", "planetscope":
		itemType = "PSOrthoTile"
	case "Landsat8L1G", "landsat":
		itemType = "Landsat8L1G"
	case "Sentinel2L1C", "sentinel":
		itemType = "Sentinel2L1C"
	case "PSScene4Band":
		// No op
	default:
		message := fmt.Sprintf("The item type value of %v is invalid", itemType)
		util.LogSimpleErr(&h.Context, message, nil)
		util.HTTPError(request, writer, &h.Context, message, http.StatusBadRequest)
		return
	}

	bboxString := request.FormValue("bbox")
	if bboxString != "" {
		if bbox, err = geojson.NewBoundingBox(bboxString); err != nil {
			message := fmt.Sprintf("The bbox value of %v is invalid", bboxString)
			util.LogSimpleErr(&h.Context, message, err)
			util.HTTPError(request, writer, &h.Context, message, http.StatusBadRequest)
			return
		}
	}

	options := SearchOptions{
		ItemType:        itemType,
		CloudCover:      cloudCover,
		AcquiredDate:    request.FormValue("acquiredDate"),
		MaxAcquiredDate: request.FormValue("maxAcquiredDate"),
		Tides:           tides,
		Bbox:            bbox}

	if fc, err = GetScenes(options, &h.Context); err == nil {
		if bytes, err = geojson.Write(fc); err != nil {
			err = util.LogSimpleErr(&h.Context, fmt.Sprintf("Failed to write output GeoJSON from:\n%#v", fc), err)
			util.HTTPError(request, writer, &h.Context, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(bytes)
		util.LogAudit(&h.Context, util.LogAuditInput{Actor: "anon user", Action: request.Method + " response", Actee: request.URL.String(), Message: "Sending /discover response", Severity: util.INFO})
	} else {
		switch herr := err.(type) {
		case util.HTTPErr:
			util.HTTPError(request, writer, &h.Context, herr.Message, herr.Status)
		default:
			err = util.LogSimpleErr(&h.Context, "Failed to get Planet Labs scenes. ", err)
			util.HTTPError(request, writer, &h.Context, err.Error(), http.StatusInternalServerError)
		}
	}
}

// MetadataHandler is a handler for /planet
// @Title planetMetadataHandler
// @Description Gets image metadata from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true         "Planet Labs Item Type, e.g., rapideye or planetscope"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.Feature
// @Failure 400 {object}  string
// @Router /planet/{itemType}/{id} [get]
type MetadataHandler struct {
	Context Context
}

// NewMetadataHandler creates a new handler using configuration
// from environment variables
func NewMetadataHandler() MetadataHandler {
	planetBaseURL := os.Getenv("PL_API_URL")
	if planetBaseURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Planet Labs API URL from the environment. Using default.")
		planetBaseURL = "https://api.planet.com"
	}

	tidesURL := os.Getenv("BF_TIDE_PREDICTION_URL")
	if tidesURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Tide Prediction URL from the environment. Using default.")
		tidesURL = "https://bf-tideprediction.int.geointservices.io/tides"
	}

	return MetadataHandler{
		Context: Context{
			BasePlanetURL: planetBaseURL,
			BaseTidesURL:  tidesURL,
		},
	}
}

// ServeHTTP implements the http.Handler interface for the MetadataHandler type
func (h MetadataHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		feature *geojson.Feature
		bytes   []byte
		options MetadataOptions
		asset   Asset
	)

	util.LogAudit(&h.Context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving /planet/{itemType}/{id} request", Severity: util.INFO})

	if util.Preflight(writer, request, &h.Context) {
		return
	}
	vars := mux.Vars(request)
	options.ID = vars["id"]
	if options.ID == "" {

		util.LogSimpleErr(&h.Context, noPlanetImageID, nil)
		util.HTTPError(request, writer, &h.Context, noPlanetImageID, http.StatusBadRequest)
		return
	}

	h.Context.PlanetKey = request.FormValue("PL_API_KEY")

	if h.Context.PlanetKey == "" {
		util.LogAlert(&h.Context, noPlanetKey)
		util.HTTPError(request, writer, &h.Context, noPlanetKey, http.StatusBadRequest)
		return
	}

	options.Tides, _ = strconv.ParseBool(request.FormValue("tides"))

	itemType := vars["itemType"]
	switch itemType {
	case "REOrthoTile", "rapideye":
		options.ItemType = "REOrthoTile"
	case "PSOrthoTile", "planetscope":
		options.ItemType = "PSOrthoTile"
	case "Landsat8L1G", "landsat":
		options.ItemType = "Landsat8L1G"
	case "Sentinel2L1C", "sentinel":
		options.ItemType = "Sentinel2L1C"
	case "PSScene4Band":
		// No op
	default:
		message := fmt.Sprintf("The item type value of %v is invalid", itemType)
		util.LogSimpleErr(&h.Context, message, nil)
		util.HTTPError(request, writer, &h.Context, message, http.StatusBadRequest)
		return
	}

	if feature, err = GetMetadata(options, &h.Context); err == nil {
		if asset, err = GetAsset(options, &h.Context); err == nil {
			injectAssetIntoMetadata(feature, asset)
			if bytes, err = geojson.Write(feature); err != nil {
				err = util.LogSimpleErr(&h.Context, fmt.Sprintf("Failed to write output GeoJSON from:\n%#v", feature), err)
				util.HTTPError(request, writer, &h.Context, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(bytes)

			util.LogAudit(&h.Context, util.LogAuditInput{Actor: request.URL.String(), Action: request.Method + " response", Actee: "anon user", Message: "Sending planet/{itemType}/{id} response", Severity: util.INFO})
		} else {
			switch herr := err.(type) {
			case util.HTTPErr:
				util.HTTPError(request, writer, &h.Context, herr.Message, herr.Status)
			default:
				err = util.LogSimpleErr(&h.Context, "Failed to get Planet Labs asset information. ", err)
				util.HTTPError(request, writer, &h.Context, err.Error(), 0)
			}
		}
	} else {
		switch herr := err.(type) {
		case util.HTTPErr:
			util.HTTPError(request, writer, &h.Context, herr.Message, herr.Status)
		default:
			err = util.LogSimpleErr(&h.Context, "Failed to get Planet Labs scene metadata. ", err)
			util.HTTPError(request, writer, &h.Context, err.Error(), 0)
		}
	}
}

// ActivateHandler is a handler for /planet
// @Title planetActivateHandler
// @Description Activates a scene
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true         "Planet Labs Item Type, e.g., rapideye or planetscope"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Success 200 {object}  geojson.Feature
// @Failure 400 {object}  string
// @Router /planet/activate/{itemType}/{id} [post]
type ActivateHandler struct {
	Context Context
}

// NewActivateHandler creates a new handler using configuration
// from environment variables
func NewActivateHandler() ActivateHandler {
	planetBaseURL := os.Getenv("PL_API_URL")
	if planetBaseURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Planet Labs API URL from the environment. Using default.")
		planetBaseURL = "https://api.planet.com"
	}

	tidesURL := os.Getenv("BF_TIDE_PREDICTION_URL")
	if tidesURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Tide Prediction URL from the environment. Using default.")
		tidesURL = "https://bf-tideprediction.int.geointservices.io/tides"
	}

	return ActivateHandler{
		Context: Context{
			BasePlanetURL: planetBaseURL,
			BaseTidesURL:  tidesURL,
		},
	}
}

// ServeHTTP implements the http.Handler interface for the ActivateHandler type
func (h ActivateHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var (
		err      error
		options  MetadataOptions
		response *http.Response
	)

	util.LogAudit(&h.Context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving /planet/activate/{itemType}/{id} request", Severity: util.INFO})

	if util.Preflight(writer, request, &h.Context) {
		return
	}
	vars := mux.Vars(request)
	options.ID = vars["id"]
	if options.ID == "" {
		util.LogSimpleErr(&h.Context, noPlanetImageID, nil)
		util.HTTPError(request, writer, &h.Context, noPlanetImageID, http.StatusBadRequest)
		return
	}

	h.Context.PlanetKey = request.FormValue("PL_API_KEY")

	if h.Context.PlanetKey == "" {

		util.LogAlert(&h.Context, noPlanetKey)
		util.HTTPError(request, writer, &h.Context, noPlanetKey, http.StatusBadRequest)
		return
	}

	itemType := vars["itemType"]
	switch itemType {
	case "REOrthoTile", "rapideye":
		options.ItemType = "REOrthoTile"
	case "PSOrthoTile", "planetscope":
		options.ItemType = "PSOrthoTile"
	// case "Sentinel2L1C", "sentinel":			// Sentinel does not need activation
	// 	 options.ItemType = "Sentinel2L1C"
	case "PSScene4Band":
		// No op
	// case "landsat":		                  // LandSat does not need activation
	// 	 options.ItemType = "Landsat8L1G"
	default:
		message := fmt.Sprintf("The item type value of %v is invalid", itemType)
		util.LogSimpleErr(&h.Context, message, nil)
		util.HTTPError(request, writer, &h.Context, message, http.StatusBadRequest)
		return
	}

	if response, err = Activate(options, &h.Context); err == nil {
		defer response.Body.Close()
		writer.Header().Set("Content-Type", response.Header.Get("Content-Type"))
		if (response.StatusCode >= 200) && (response.StatusCode < 300) {
			bytes, _ := ioutil.ReadAll(response.Body)
			writer.Write(bytes)
			util.LogAudit(&h.Context, util.LogAuditInput{Actor: request.URL.String(), Action: request.Method + " response", Actee: "anon user", Message: "Sending planet/{itemType}/{id} response", Severity: util.INFO})
		} else {
			message := fmt.Sprintf("Failed to activate Planet Labs scene: " + response.Status)
			err = util.LogSimpleErr(&h.Context, message, nil)
			util.HTTPError(request, writer, &h.Context, err.Error(), response.StatusCode)
		}
	} else {
		switch herr := err.(type) {
		case util.HTTPErr:
			util.HTTPError(request, writer, &h.Context, herr.Message, herr.Status)
		default:
			err = util.LogSimpleErr(&h.Context, "Failed to activate Planet Labs scene. ", err)
			util.HTTPError(request, writer, &h.Context, err.Error(), 0)
		}
	}
}

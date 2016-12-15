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
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

const noPlanetKey = "This operation requires a Planet Labs API key."

// DiscoverHandler is a handler for /planet/discover
// @Title planetDiscoverHandler
// @Description discovers scenes from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true        "Planet Labs Item Type, e.g., REOrthoTile"
// @Param   bbox            query   string  false        "The bounding box, as a GeoJSON Bounding box (x1,y1,x2,y2)"
// @Param   acquiredDate    query   string  false        "The minimum (earliest) acquired date, as RFC 3339"
// @Param   maxAcquiredDate query   string  false        "The maximum acquired date, as RFC 3339"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.FeatureCollection
// @Failure 400 {object}  string
// @Router /planet/discover/{itemType} [get]
func DiscoverHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		fc        *geojson.FeatureCollection
		err       error
		planetKey string
		itemType  string
		bytes     []byte
		bbox      geojson.BoundingBox
	)
	if util.Preflight(writer, request) {
		return
	}

	tides, _ := strconv.ParseBool(request.FormValue("tides"))
	planetKey = request.FormValue("PL_API_KEY")
	if planetKey == "" {
		http.Error(writer, noPlanetKey, http.StatusBadRequest)
		return
	}

	itemType = mux.Vars(request)["itemType"]

	bboxString := request.FormValue("bbox")
	if bboxString != "" {
		if bbox, err = geojson.NewBoundingBox(bboxString); err != nil {
			http.Error(writer, "The bbox value of "+bboxString+" is invalid: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	options := SearchOptions{
		Bbox: bbox}
	context := Context{
		ItemType:  itemType,
		Tides:     tides,
		PlanetKey: planetKey}

	if fc, err = GetScenes(options, context); err == nil {
		if bytes, err = geojson.Write(fc); err != nil {
			http.Error(writer, util.TraceStr(err.Error()), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(bytes)
	} else {
		http.Error(writer, util.TraceStr(err.Error()), http.StatusInternalServerError)
	}
}

// AssetHandler is a handler for /planet/asset
// @Title planetAssetHandler
// @Description Gets asset information from Planet Labs; on a POST request will also trigger activation if needed
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true        "Planet Labs Item Type, e.g., REOrthoTile"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Success 200 {object}  string
// @Failure 400 {object}  string
// @Router /planet/asset/{itemType}/{id} [get,post]
func AssetHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		context Context
		result  []byte
		options AssetOptions
	)
	if util.Preflight(writer, request) {
		return
	}
	vars := mux.Vars(request)
	options.ID = vars["id"]
	if options.ID == "" {
		http.Error(writer, "This operation requires a Planet Labs image ID.", http.StatusBadRequest)
		return
	}
	context.PlanetKey = request.FormValue("PL_API_KEY")

	if context.PlanetKey == "" {
		http.Error(writer, "This operation requires a Planet Labs API key.", http.StatusBadRequest)
		return
	}

	context.ItemType = vars["itemType"]

	if request.Method == "POST" {
		options.activate = true
	}

	if result, err = GetAsset(options, context); err == nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(result)
	} else {
		http.Error(writer, "Failed to acquire activation information for "+options.ID+": "+err.Error(), http.StatusBadRequest)
	}
}

// MetadataHandler is a handler for /planet
// @Title planetMetadataHandler
// @Description Gets image metadata from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true        "Planet Labs Item Type, e.g., REOrthoTile"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Success 200 {object}  geojson.Feature
// @Failure 400 {object}  string
// @Router /planet/{itemType}/{id} [get]
func MetadataHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		context Context
		feature *geojson.Feature
		bytes   []byte
		options AssetOptions
	)
	if util.Preflight(writer, request) {
		return
	}
	vars := mux.Vars(request)
	options.ID = vars["id"]
	if options.ID == "" {
		http.Error(writer, "This operation requires a Planet Labs image ID.", http.StatusBadRequest)
		return
	}
	context.PlanetKey = request.FormValue("PL_API_KEY")

	if context.PlanetKey == "" {
		http.Error(writer, "This operation requires a Planet Labs API key.", http.StatusBadRequest)
		return
	}

	context.ItemType = vars["itemType"]

	if feature, err = GetMetadata(options, context); err == nil {
		if bytes, err = geojson.Write(feature); err != nil {
			http.Error(writer, util.TraceStr(err.Error()), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(bytes)
	} else {
		http.Error(writer, "Failed to acquire activation information for "+options.ID+": "+err.Error(), http.StatusInternalServerError)
	}
}

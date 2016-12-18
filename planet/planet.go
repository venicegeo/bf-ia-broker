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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/venicegeo/bf-ia-broker/tides"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

var baseURLString string

func init() {
	baseURLString = os.Getenv("PL_API_URL")
	if baseURLString == "" {
		baseURLString = "https://api.planet.com/"
	}
}

// Context is the context for a Planet Labs Operation
type Context struct {
	PlanetKey string
	sessionID string
}

// AppName returns an empty string
func (c *Context) AppName() string {
	return "bf-ia-broker"
}

// SessionID returns a Session ID, creating one if needed
func (c *Context) SessionID() string {
	if c.sessionID == "" {
		c.sessionID, _ = util.PsuUUID()
	}
	return c.sessionID
}

// LogRootDir returns an empty string
func (c *Context) LogRootDir() string {
	return ""
}

// SearchOptions are the search options for a quick-search request
type SearchOptions struct {
	ItemType        string
	Tides           bool
	AcquiredDate    string
	MaxAcquiredDate string
	Bbox            geojson.BoundingBox
	CloudCover      float64
}

type request struct {
	ItemTypes []string `json:"item_types"`
	Filter    filter   `json:"filter"`
}

type filter struct {
	Type   string        `json:"type"`
	Config []interface{} `json:"config"`
}

type objectFilter struct {
	Type      string      `json:"type"`
	FieldName string      `json:"field_name"`
	Config    interface{} `json:"config"`
}

type dateConfig struct {
	GTE string `json:"gte,omitempty"`
	LTE string `json:"lte,omitempty"`
	GT  string `json:"gt,omitempty"`
	LT  string `json:"lt,omitempty"`
}

type rangeConfig struct {
	GTE float64 `json:"gte,omitempty"`
	LTE float64 `json:"lte,omitempty"`
	GT  float64 `json:"gt,omitempty"`
	LT  float64 `json:"lt,omitempty"`
}

type doRequestInput struct {
	method      string
	inputURL    string // URL may be relative or absolute based on baseURLString
	body        []byte
	contentType string
}

// doRequest performs the request
func doRequest(input doRequestInput, context Context) (*http.Response, error) {
	var (
		request   *http.Request
		parsedURL *url.URL
		inputURL  string
		err       error
	)
	inputURL = input.inputURL
	if !strings.Contains(inputURL, baseURLString) {
		baseURL, _ := url.Parse(baseURLString)
		parsedRelativeURL, _ := url.Parse(input.inputURL)
		resolvedURL := baseURL.ResolveReference(parsedRelativeURL)

		if parsedURL, err = url.Parse(resolvedURL.String()); err != nil {
			err = fmt.Errorf("Failed to parse %v into a URL: %v", resolvedURL.String(), err.Error())
			util.LogAlert(&context, err.Error())
			return nil, err
		}
		inputURL = parsedURL.String()
	}
	util.LogInfo(&context, fmt.Sprintf("Calling %v at %v", input.method, inputURL))
	if request, err = http.NewRequest(input.method, inputURL, bytes.NewBuffer(input.body)); err != nil {
		err = fmt.Errorf("Failed to make a new HTTP request for %v: %v", inputURL, err.Error())
		util.LogAlert(&context, err.Error())
		return nil, err
	}
	if input.contentType != "" {
		request.Header.Set("Content-Type", input.contentType)
	}

	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(context.PlanetKey+":")))
	return util.HTTPClient().Do(request)
}

// Assets represents the assets available for a scene
type Assets struct {
	Analytic    Asset `json:"analytic"`
	AnalyticXML Asset `json:"analytic_xml"`
	UDM         Asset `json:"udm"`
	Visual      Asset `json:"visual"`
	VisualXML   Asset `json:"visual_xml"`
}

// Asset represents a single asset available for a scene
type Asset struct {
	Links       Links    `json:"_links"`
	Status      string   `json:"status"`
	Type        string   `json:"type"`
	Location    string   `json:"location,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	Permissions []string `json:"_permissions,omitempty"`
}

// Links represents the links JSON structure.
type Links struct {
	Self     string `json:"_self"`
	Activate string `json:"activate"`
	Type     string `json:"type"`
}

// GetScenes returns a FeatureCollection containing the scenes requested
func GetScenes(options SearchOptions, context Context) (*geojson.FeatureCollection, error) {
	var (
		err      error
		response *http.Response
		body     []byte
		req      request
		fc       *geojson.FeatureCollection
	)

	req.ItemTypes = append(req.ItemTypes, options.ItemType)
	req.Filter.Type = "AndFilter"
	req.Filter.Config = make([]interface{}, 0)
	if options.Bbox != nil {
		req.Filter.Config = append(req.Filter.Config, objectFilter{Type: "GeometryFilter", FieldName: "geometry", Config: options.Bbox.Polygon()})
	}
	if options.AcquiredDate != "" || options.MaxAcquiredDate != "" {
		dc := dateConfig{GTE: options.AcquiredDate, LTE: options.MaxAcquiredDate}
		req.Filter.Config = append(req.Filter.Config, objectFilter{Type: "DateRangeFilter", FieldName: "acquired", Config: dc})
	}
	if options.CloudCover > 0 {
		cc := rangeConfig{LTE: options.CloudCover}
		req.Filter.Config = append(req.Filter.Config, objectFilter{Type: "RangeFilter", FieldName: "cloud_cover", Config: cc})
	}
	if body, err = json.Marshal(req); err != nil {
		err = fmt.Errorf("Failed to marshal request object %#v: %v", req, err.Error())
		util.LogAlert(&context, err.Error())
		return nil, err
	}
	if response, err = doRequest(doRequestInput{method: "POST", inputURL: "data/v1/quick-search", body: body, contentType: "application/json"}, context); err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	if fc, err = transformSRBody(body, context); err != nil {
		return nil, err
	}
	if options.Tides {
		var context tides.Context
		if fc, err = tides.GetTides(fc, context); err != nil {
			return nil, err
		}
	}
	return fc, nil
}

// Transforms search results into a FeatureCollection for later use
func transformSRBody(body []byte, context Context) (*geojson.FeatureCollection, error) {
	var (
		result *geojson.FeatureCollection
		fc     *geojson.FeatureCollection
		fci    interface{}
		err    error
		ok     bool
	)
	if fci, err = geojson.Parse(body); err != nil {
		err = fmt.Errorf("Failed to parse GeoJSON: %v" + err.Error())
		util.LogAlert(&context, err.Error())
		util.LogInfo(&context, string(body))
		return nil, err
	}
	if fc, ok = fci.(*geojson.FeatureCollection); !ok {
		err = fmt.Errorf("Expected a FeatureCollection and got %T", fci)
		util.LogAlert(&context, err.Error())
		util.LogInfo(&context, string(body))
		return nil, err
	}
	features := make([]*geojson.Feature, len(fc.Features))
	for inx, curr := range fc.Features {
		features[inx] = transformSRFeature(curr)
	}
	result = geojson.NewFeatureCollection(features)
	return result, nil
}

func transformSRFeature(feature *geojson.Feature) *geojson.Feature {
	properties := make(map[string]interface{})
	properties["cloudCover"] = feature.Properties["cloud_cover"].(float64)
	id := feature.IDStr()
	properties["resolution"] = feature.Properties["gsd"].(float64)
	adString := feature.Properties["acquired"].(string)
	properties["acquiredDate"] = adString
	properties["fileFormat"] = "geotiff"
	properties["sensorName"] = feature.Properties["satellite_id"].(string)
	result := geojson.NewFeature(feature.Geometry, id, properties)
	result.Bbox = result.ForceBbox()
	return result
}

// AssetOptions are the options for the Asset func
type AssetOptions struct {
	ID       string
	activate bool
	ItemType string
}

// GetAsset returns the status of the analytic asset and
// attempts to activate it if needed
func GetAsset(options AssetOptions, context Context) ([]byte, error) {
	var (
		response *http.Response
		err      error
		body     []byte
		assets   Assets
	)
	if response, err = doRequest(doRequestInput{method: "GET", inputURL: "data/v1/item-types/" + options.ItemType + "/items/" + options.ID + "/assets/"}, context); err != nil {
		err = fmt.Errorf("Failed to get asset metadata for %v: %v", options.ID, err.Error())
		util.LogAlert(&context, err.Error())
		return nil, err
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	if err = json.Unmarshal(body, &assets); err != nil {
		err = fmt.Errorf("Failed to Unmarshal response: %v", err.Error())
		util.LogAlert(&context, err.Error())
		util.LogInfo(&context, string(body))
		return nil, err
	}
	if options.activate && (assets.Analytic.Status == "inactive") {
		go doRequest(doRequestInput{method: "POST", inputURL: assets.Analytic.Links.Activate}, context)
	}
	return json.Marshal(assets.Analytic)
}

// GetMetadata returns the Beachfront metadata for a single scene
func GetMetadata(options AssetOptions, context Context) (*geojson.Feature, error) {
	var (
		response *http.Response
		err      error
		body     []byte
		feature  geojson.Feature
	)
	util.LogInfo(&context, "Calling Planet Labs to get metadata for scene "+options.ID)
	input := doRequestInput{method: "GET", inputURL: "data/v1/item-types/" + options.ItemType + "/items/" + options.ID}
	if response, err = doRequest(input, context); err != nil {
		err = fmt.Errorf("Failed to get metadata for %v: %v", options.ID, err.Error())
		util.LogAlert(&context, err.Error())
		return nil, err
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	if err = json.Unmarshal(body, &feature); err != nil {
		err = fmt.Errorf("Failed to Unmarshal response: %v", err.Error())
		util.LogAlert(&context, err.Error())
		util.LogInfo(&context, string(body))
		return nil, err
	}
	return transformSRFeature(&feature), nil
}

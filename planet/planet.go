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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/venicegeo/bf-ia-broker/landsat"
	"github.com/venicegeo/bf-ia-broker/tides"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

var disablePermissionsCheck bool

func init() {
	disablePermissionsCheck, _ = strconv.ParseBool(os.Getenv("PL_DISABLE_PERMISSIONS_CHECK"))
	if disablePermissionsCheck {
		util.LogInfo(&util.BasicLogContext{}, "Disabling Planet Labs permissions check")
	}
}

// Context is the context for a Planet Labs Operation
type Context struct {
	BasePlanetURL string
	BaseTidesURL  string
	PlanetKey     string
	sessionID     string
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

type searchResults struct {
	Features []feature `json:"features"`
}

type feature struct {
	Links       Links    `json:"_links"`
	Permissions []string `json:"_permissions"`
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

type doRequestInput struct {
	method      string
	inputURL    string // URL may be relative or absolute based on baseURLString
	body        []byte
	contentType string
}

// MetadataOptions are the options for the Asset func
type MetadataOptions struct {
	ID       string
	Tides    bool
	ItemType string
}

// GetScenes returns a FeatureCollection containing the scenes requested
func GetScenes(options SearchOptions, context *Context) (*geojson.FeatureCollection, error) {
	var (
		err          error
		response     *http.Response
		requestBody  []byte
		responseBody []byte
		req          request
		fc           *geojson.FeatureCollection
	)

	req.ItemTypes = append(req.ItemTypes, options.ItemType)
	req.Filter.Type = "AndFilter"
	req.Filter.Config = make([]interface{}, 0)
	if options.Bbox != nil {
		req.Filter.Config = append(req.Filter.Config, objectFilter{Type: "GeometryFilter", FieldName: "geometry", Config: options.Bbox.Geometry()})
	}
	if options.AcquiredDate != "" || options.MaxAcquiredDate != "" {
		dc := dateConfig{GTE: options.AcquiredDate, LTE: options.MaxAcquiredDate}
		req.Filter.Config = append(req.Filter.Config, objectFilter{Type: "DateRangeFilter", FieldName: "acquired", Config: dc})
	}
	if options.CloudCover > 0 {
		cc := rangeConfig{LTE: options.CloudCover}
		req.Filter.Config = append(req.Filter.Config, objectFilter{Type: "RangeFilter", FieldName: "cloud_cover", Config: cc})
	}
	if requestBody, err = json.Marshal(req); err != nil {
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to marshal request object %#v.", req), err)
		return nil, err
	}
	if response, err = doRequest(doRequestInput{method: "POST", inputURL: "data/v1/quick-search", body: requestBody, contentType: "application/json"}, context); err != nil {
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to complete Planet Labs request %#v.", requestBody), err)
		return nil, err
	}
	switch {
	case (response.StatusCode >= 400) && (response.StatusCode < 500):
		message := fmt.Sprintf("Failed to discover scenes from Planet Labs: %v. ", response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
	case response.StatusCode >= 500:
		err = util.LogSimpleErr(context, "Failed to discover scenes from Planet Labs.", errors.New(response.Status))
		return nil, err
	default:
		//no op
	}

	defer response.Body.Close()
	responseBody, _ = ioutil.ReadAll(response.Body)

	if fc, err = transformSRBody(responseBody, context); err != nil {
		return nil, err
	}
	if options.Tides {
		tidesContext := tides.Context{TidesURL: context.BaseTidesURL}
		if fc, err = tides.GetTides(fc, &tidesContext); err != nil {
			return nil, err
		}
	}
	return fc, nil
}

// GetAsset returns the status of the analytic asset and
// attempts to activate it if needed
func GetAsset(options MetadataOptions, context *Context) (Asset, error) {
	var (
		result   Asset
		response *http.Response
		err      error
		body     []byte
		assets   Assets
	)
	// Note: trailing `/` is needed here to avoid a redirect which causes a Go 1.7 redirect bug issue
	inputURL := "data/v1/item-types/" + options.ItemType + "/items/" + options.ID + "/assets/"
	if response, err = doRequest(doRequestInput{method: "GET", inputURL: inputURL}, context); err != nil {
		return result, err
	}
	switch {
	case (response.StatusCode >= 400) && (response.StatusCode < 500):
		message := fmt.Sprintf("Failed to get asset information for scene %v: %v. ", options.ID, response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return result, err
	case response.StatusCode >= 500:
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to get asset information for scene %v. ", options.ID), errors.New(response.Status))
		return result, err
	default:
		//no op
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	if err = json.Unmarshal(body, &assets); err != nil {
		plErr := util.Error{LogMsg: "Failed to Unmarshal response from Planet Labs data request: " + err.Error(),
			SimpleMsg:  "Planet Labs returned an unexpected response for this request. See log for further details.",
			Response:   string(body),
			URL:        inputURL,
			HTTPStatus: response.StatusCode}
		err = plErr.Log(context, "")
		return result, err
	}
	return assets.Analytic, nil
}

// GetMetadata returns the Beachfront metadata for a single scene
func GetMetadata(options MetadataOptions, context *Context) (*geojson.Feature, error) {
	var (
		response *http.Response
		err      error
		body     []byte
		feature  geojson.Feature
	)
	inputURL := "data/v1/item-types/" + options.ItemType + "/items/" + options.ID
	input := doRequestInput{method: "GET", inputURL: inputURL}
	if response, err = doRequest(input, context); err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	switch {
	case (response.StatusCode >= 400) && (response.StatusCode < 500):
		message := fmt.Sprintf("Failed to find metadata for scene %v: %v. ", options.ID, response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
	case response.StatusCode >= 500:
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to retrieve metadata for scene %v. ", options.ID), errors.New(response.Status))
		return nil, err
	default:
		//no op
	}
	if err = json.Unmarshal(body, &feature); err != nil {
		plErr := util.Error{LogMsg: "Failed to Unmarshal response from Planet Labs data request: " + err.Error(),
			SimpleMsg:  "Planet Labs returned an unexpected response for this request. See log for further details.",
			Response:   string(body),
			URL:        inputURL,
			HTTPStatus: response.StatusCode}
		err = plErr.Log(context, "")
		return nil, err
	}
	feature = *transformSRFeature(&feature, context)
	if options.Tides {
		var (
			tc tides.Context
		)
		tc.TidesURL = context.BaseTidesURL
		fc := geojson.NewFeatureCollection([]*geojson.Feature{&feature})
		if fc, err = tides.GetTides(fc, &tc); err != nil {
			return nil, err
		}
		feature = *fc.Features[0]
	}

	return &feature, nil
}

// Activate retrieves and activates the analytic asset.
func Activate(options MetadataOptions, context *Context) (*http.Response, error) {
	var (
		asset Asset
		err   error
	)
	if asset, err = GetAsset(options, context); err != nil {
		return nil, err
	}
	return doRequest(doRequestInput{method: "POST", inputURL: asset.Links.Activate}, context)
}

// doRequest performs the request
func doRequest(input doRequestInput, context *Context) (*http.Response, error) {
	var (
		request   *http.Request
		parsedURL *url.URL
		inputURL  string
		err       error
	)
	inputURL = input.inputURL
	if !strings.Contains(inputURL, context.BasePlanetURL) {
		baseURL, _ := url.Parse(context.BasePlanetURL)
		parsedRelativeURL, _ := url.Parse(input.inputURL)
		resolvedURL := baseURL.ResolveReference(parsedRelativeURL)

		if parsedURL, err = url.Parse(resolvedURL.String()); err != nil {
			err = util.LogSimpleErr(context, fmt.Sprintf("Failed to parse %v into a URL.", resolvedURL.String()), err)
			return nil, err
		}
		inputURL = parsedURL.String()
	}
	message := "Requesting data from Planet Labs"
	bodyStr := string(input.body)
	if bodyStr != "" {
		message += ": " + bodyStr
	}
	if request, err = http.NewRequest(input.method, inputURL, bytes.NewBuffer(input.body)); err != nil {
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to make a new HTTP request for %v.", inputURL), err)
		return nil, err
	}
	if input.contentType != "" {
		request.Header.Set("Content-Type", input.contentType)
	}

	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(context.PlanetKey+":")))
	message = fmt.Sprintf("%v\nHeader:\n%#v", message, request.Header)
	util.LogAudit(context, util.LogAuditInput{Actor: "planet/doRequest", Action: input.method, Actee: inputURL, Message: message, Severity: util.INFO})
	util.LogAudit(context, util.LogAuditInput{Actor: inputURL, Action: input.method + " response", Actee: "planet/doRequest", Message: "Receiving data from Planet Labs", Severity: util.INFO})
	return util.HTTPClient().Do(request)
}

func scontains(input []string, check string) bool {
	for _, curr := range input {
		if curr == check {
			return true
		}
	}
	return false
}

// Transforms search results into a FeatureCollection for later use
func transformSRBody(body []byte, context util.LogContext) (*geojson.FeatureCollection, error) {
	var (
		result    *geojson.FeatureCollection
		fc        *geojson.FeatureCollection
		fci       interface{}
		err       error
		ok        bool
		features  []*geojson.Feature
		plResults searchResults
	)
	if fci, err = geojson.Parse(body); err != nil {
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to parse GeoJSON.\n%v", string(body)), err)
		return nil, err
	}
	if fc, ok = fci.(*geojson.FeatureCollection); !ok {
		plErr := util.Error{SimpleMsg: fmt.Sprintf("Expected a FeatureCollection and got %T", fci),
			Response: string(body)}
		err = plErr.Log(context, "")
		return nil, err
	}
	if err = json.Unmarshal(body, &plResults); err != nil {
		return result, err
	}
	for inx, curr := range fc.Features {

		// We need to suppress scenes that we don't have permissions for
		if disablePermissionsCheck || scontains(plResults.Features[inx].Permissions, "assets.analytic:download") {
			features = append(features, transformSRFeature(curr, context))
			// } else {
			// 	util.LogInfo(context, fmt.Sprintf("Skipping scene %v due to lack of permissions.", curr.IDStr()))
		}
	}
	result = geojson.NewFeatureCollection(features)
	return result, nil
}

func transformSRFeature(feature *geojson.Feature, context util.LogContext) *geojson.Feature {
	properties := make(map[string]interface{})
	if cc, ok := feature.Properties["cloud_cover"].(float64); ok {
		properties["cloudCover"] = cc * 100.0
	} else {
		properties["cloudCover"] = -1.0
	}
	id := feature.IDStr()
	properties["resolution"], _ = feature.Properties["gsd"].(float64)
	adString, _ := feature.Properties["acquired"].(string)
	properties["acquiredDate"] = adString
	properties["fileFormat"] = "geotiff"
	properties["sensorName"], _ = feature.Properties["satellite_id"].(string)

	if landsat.IsValidLandSatID(id) {
		dataType, _ := feature.Properties["data_type"].(string)
		err := addLandsatS3BandsToProperties(id, dataType, &properties)
		if err != nil {
			util.LogAlert(context, err.Error()+" :: in LandSat feature: "+feature.String())
		}
	}

	if isSentinelFeature(id) {
		properties["fileFormat"] = "jpeg2000"
		err := addSentinelS3BandsToProperties(id, &properties)
		if err != nil {
			util.LogAlert(context, err.Error()+" :: in Sentinel-2 feature: "+feature.String())
		}
	}

	result := geojson.NewFeature(feature.Geometry, id, properties)
	result.Bbox = result.ForceBbox()
	return result
}

func injectAssetIntoMetadata(feature *geojson.Feature, asset Asset) {
	if asset.ExpiresAt != "" {
		feature.Properties["expires_at"] = asset.ExpiresAt
	}
	if asset.Location != "" {
		feature.Properties["location"] = asset.Location
	}
	if len(asset.Permissions) > 0 {
		feature.Properties["permissions"] = asset.Permissions
	}
	if asset.Status != "" {
		feature.Properties["status"] = asset.Status
	}
	if asset.Type != "" {
		feature.Properties["type"] = asset.Type
	}
}

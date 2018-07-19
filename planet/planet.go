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
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/tides"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

var addTidesToSearchResults = tides.AddTidesToSearchResults

// GetScenes returns a FeatureCollection containing the scenes requested
func GetScenes(options SearchOptions, context *Context) (*geojson.FeatureCollection, error) {
	var (
		err          error
		response     *http.Response
		requestBody  []byte
		responseBody []byte
		req          request
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
	if response, err = planetRequest(planetRequestInput{method: "POST", inputURL: "data/v1/quick-search", body: requestBody, contentType: "application/json"}, context); err != nil {
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to complete Planet API request %#v.", string(requestBody)), err)
		return nil, err
	}
	switch {
	case (response.StatusCode == http.StatusUnauthorized) || (response.StatusCode == http.StatusForbidden):
		message := fmt.Sprintf("Specified API key is invalid or has inadequate permissions. (%v) ", response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
	case (response.StatusCode >= 400) && (response.StatusCode < 500):
		message := fmt.Sprintf("Failed to discover scenes from Planet API: %v. ", response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
	case response.StatusCode >= 500:
		err = util.LogSimpleErr(context, "Failed to discover scenes from Planet API.", errors.New(response.Status))
		return nil, err
	default:
		//no op
	}

	defer response.Body.Close()
	responseBody, _ = ioutil.ReadAll(response.Body)

	results, err := parseSearchResults(context, responseBody)
	if err != nil {
		return nil, err
	}
	if options.Tides {
		tidesContext := tides.Context{TidesURL: context.BaseTidesURL}
		if err = addTidesToSearchResults(&tidesContext, results); err != nil {
			return nil, err
		}
	}

	featureCreators := make([]model.GeoJSONFeatureCreator, len(results))
	for i, result := range results {
		featureCreators[i] = result
	}

	return model.MultiBrokerResult{FeatureCreators: featureCreators}.GeoJSONFeatureCollection()
}

// GetPlanetAssets returns the asset metadata related to a particular item
func GetPlanetAssets(options MetadataOptions, context *Context) (*model.PlanetAssetMetadata, error) {
	var (
		response *http.Response
		err      error
		body     []byte
		assets   Assets
	)
	// Note: trailing `/` is needed here to avoid a redirect which causes a Go 1.7 redirect bug issue
	inputURL := "data/v1/item-types/" + options.ItemType + "/items/" + options.ID + "/assets/"
	if response, err = planetRequest(planetRequestInput{method: "GET", inputURL: inputURL}, context); err != nil {
		return nil, err
	}
	switch {
	case (response.StatusCode == http.StatusUnauthorized) || (response.StatusCode == http.StatusForbidden):
		message := fmt.Sprintf("Specified API key is invalid or has inadequate permissions. (%v) ", response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
	case (response.StatusCode >= 400) && (response.StatusCode < 500):
		message := fmt.Sprintf("Failed to get asset information for scene %v: %v. ", options.ID, response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
	case response.StatusCode >= 500:
		err = util.LogSimpleErr(context, fmt.Sprintf("Failed to get asset information for scene %v. ", options.ID), errors.New(response.Status))
		return nil, err
	default:
		//no op
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	if err = json.Unmarshal(body, &assets); err != nil {
		plErr := util.Error{LogMsg: "Failed to Unmarshal response from Planet API data request: " + err.Error(),
			SimpleMsg:  "Planet Labs returned an unexpected response for this request. See log for further details.",
			Response:   string(body),
			URL:        inputURL,
			HTTPStatus: response.StatusCode}
		err = plErr.Log(context, "")
		return nil, err
	}

	assetMetadata, err := planetAssetMetadataFromAssets(assets)

	log.Print("XXXXXXXX")
	log.Print(string(body))
	log.Print(assets)
	log.Print(assetMetadata)
	log.Print(err)

	if err == nil && itemTypeRequiresActivation(options.ItemType) {
		if assetMetadata == nil {
			err = errors.New("Found no asset data in response for item type requiring asset activation")
		} else if assetMetadata.ActivationURL.String() == "" {
			err = errors.New("Found no asset activation URL for item type requiring asset activation")
		} else if assetMetadata.Status == "active" {
			if assetMetadata.AssetURL.String() == "" {
				err = errors.New("Found no asset URL for supposedly active item")
			} else if assetMetadata.ExpiresAt.IsZero() {
				err = errors.New("Found no expiration time for supposedly active item")
			}
		}
	}

	if err != nil {
		plErr := util.Error{LogMsg: "Invalid data from Planet API asset request: " + err.Error(),
			SimpleMsg:  "Planet Labs returned invalid metadata for this scene's assets.",
			Response:   string(body),
			URL:        inputURL,
			HTTPStatus: response.StatusCode}
		err = plErr.Log(context, "")
		return assetMetadata, util.HTTPErr{Status: http.StatusBadGateway, Message: plErr.SimpleMsg}
	}

	return assetMetadata, nil
}

// GetPlanetItem returns the Beachfront metadata for a single scene
func GetPlanetItem(options MetadataOptions, context *Context) (*model.BrokerSearchResult, error) {
	var (
		response      *http.Response
		err           error
		body          []byte
		planetFeature geojson.Feature
	)
	inputURL := "data/v1/item-types/" + options.ItemType + "/items/" + options.ID
	input := planetRequestInput{method: "GET", inputURL: inputURL}
	if response, err = planetRequest(input, context); err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	switch {
	case (response.StatusCode == http.StatusUnauthorized) || (response.StatusCode == http.StatusForbidden):
		message := fmt.Sprintf("Specified API key is invalid or has inadequate permissions. (%v) ", response.Status)
		err := util.HTTPErr{Status: response.StatusCode, Message: message}
		util.LogAlert(context, message)
		return nil, err
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
	if err = json.Unmarshal(body, &planetFeature); err != nil {
		plErr := util.Error{LogMsg: "Failed to Unmarshal response from Planet API data request: " + err.Error(),
			SimpleMsg:  "Planet Labs returned an unexpected response for this request. See log for further details.",
			Response:   string(body),
			URL:        inputURL,
			HTTPStatus: response.StatusCode}
		err = plErr.Log(context, "")
		return nil, err
	}

	result, err := planetSearchBrokerResultFromFeature(&planetFeature)
	if err != nil {
		return nil, err
	}

	if options.Tides {
		// Hacky way to use the multi-tides query for a single query
		singleSearchResultForTides := []model.BrokerSearchResult{model.BrokerSearchResult{
			BasicBrokerResult: result.BasicBrokerResult,
		}}
		tidesContext := tides.Context{TidesURL: context.BaseTidesURL}
		if err = tides.AddTidesToSearchResults(&tidesContext, singleSearchResultForTides); err != nil {
			return nil, err
		}
		result.TidesData = singleSearchResultForTides[0].TidesData
	}

	return result, nil
}

// Activate retrieves and activates the analytic asset.
func Activate(options MetadataOptions, context *Context) (*http.Response, error) {
	var (
		assetMetadata *model.PlanetAssetMetadata
		err           error
	)
	if assetMetadata, err = GetPlanetAssets(options, context); err != nil {
		return nil, err
	}
	return planetRequest(planetRequestInput{method: "POST", inputURL: assetMetadata.ActivationURL.String()}, context)
}

// planetRequest performs the request
func planetRequest(input planetRequestInput, context *Context) (*http.Response, error) {
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
	util.LogAudit(context, util.LogAuditInput{Actor: inputURL, Action: input.method + " response", Actee: "planet/doRequest", Message: "Receiving data from Planet API", Severity: util.INFO})
	return util.HTTPClient().Do(request)
}

func itemTypeRequiresActivation(itemType string) bool {
	switch itemType {
	case "REOrthoTile":
		fallthrough
	case "PSOrthoTile":
		return true
	default:
		return false
	}
}

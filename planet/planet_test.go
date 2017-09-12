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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

func makeTestingContext(planetServer *httptest.Server, tidesServer *httptest.Server) Context {
	return Context{
		BasePlanetURL: planetServer.URL,
		BaseTidesURL:  tidesServer.URL,
		PlanetKey:     testingValidKey,
	}
}

func TestPlanetNoParameters(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	body := `{
     "item_types":[
        "REOrthoTile"
     ],
     "filter":{
        "type":"AndFilter",
        "config":[
          ]
        }
  }`

	requestInput := doRequestInput{
		method:      "POST",
		inputURL:    "data/v1/quick-search",
		body:        []byte(body),
		contentType: "application/json",
	}

	_, err := doRequest(requestInput, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)
}

func TestGetScenesBoundingBox(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	var options SearchOptions
	bbox, err := geojson.NewBoundingBox("139,50,140,51")
	assert.Nil(t, err, "Failed creating bounding box %v", err)
	options.Bbox = bbox

	_, err = GetScenes(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)
}

func TestGetScenesCloudCover(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	options := SearchOptions{CloudCover: 0.1}

	_, err := GetScenes(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)
}

func TestGetScenesAcquiredDate(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	options := SearchOptions{AcquiredDate: "2016-01-01T00:00:00Z"}

	_, err := GetScenes(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)
}

func TestGetScenesTides(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	options := SearchOptions{Tides: true}

	_, err := GetScenes(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)

}

func TestGetMetadata(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	options := SearchOptions{Tides: true}

	scenes, err := GetScenes(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)

	aOptions := MetadataOptions{ID: scenes.Features[0].IDStr(), Tides: true, ItemType: "REOrthoTile"}
	feature, err := GetMetadata(aOptions, &context)
	assert.Nil(t, err, "Failed to get asset metadata; received: %v", err)

	assert.Equal(t, aOptions.ID, feature.IDStr())
}

func TestGetAsset(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	options := SearchOptions{Tides: true}

	scenes, err := GetScenes(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)

	aOptions := MetadataOptions{ID: scenes.Features[0].IDStr(), Tides: true, ItemType: "REOrthoTile"}
	_, err = GetAsset(aOptions, &context)
	assert.Nil(t, err, "Failed to get asset; received %v", err)
}

func TestGetMetadataBadAssetID(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)
	aOptions := MetadataOptions{ID: "X-BAD-ID-X", Tides: true, ItemType: "PSOrthoTile"}

	_, err := GetMetadata(aOptions, &context)
	assert.NotNil(t, err, "Expected invalid ID asset to fail, but it succeeded.")
	if _, ok := err.(util.HTTPErr); err != nil && !ok {
		t.Errorf("Expected an HTTPErr, got a %T", err)
	}
}

func TestGetAssetNoDataFromPlanet(t *testing.T) {
	// Sometimes planet responds with a plain `{}` to a scene asset request
	// This needs to be handled properly, returning a 502 error
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	var aOptions MetadataOptions
	var err error

	// Should 502 for RapidEye missing asset data
	aOptions = MetadataOptions{ID: testingValidSceneIDWithNoMetadata, Tides: true, ItemType: "REOrthoTile"}
	_, err = GetAsset(aOptions, &context)
	if httpErr, ok := err.(util.HTTPErr); err != nil && !ok {
		t.Errorf("Expected an HTTPErr, got a %T", err)
	} else {
		assert.Equal(t, http.StatusBadGateway, httpErr.Status, "Expected error 502 but got: %v", httpErr)
	}

	// Should 502 for PlanetScope missing asset data
	aOptions = MetadataOptions{ID: testingValidSceneIDWithNoMetadata, Tides: true, ItemType: "PSOrthoTile"}
	_, err = GetAsset(aOptions, &context)
	if httpErr, ok := err.(util.HTTPErr); err != nil && !ok {
		t.Errorf("Expected an HTTPErr, got a %T", err)
	} else {
		assert.Equal(t, http.StatusBadGateway, httpErr.Status, "Expected error 502 but got: %v", httpErr)
	}

	// Should NOT error for Landsat missing asset data (it's supposed to not have it)
	aOptions = MetadataOptions{ID: testingValidSceneIDWithNoMetadata, Tides: true, ItemType: "Landsat8L1G"}
	_, err = GetAsset(aOptions, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)

	// Should NOT error for Sentinel missing asset data (it's supposed to not have it)
	aOptions = MetadataOptions{ID: testingValidSceneIDWithNoMetadata, Tides: true, ItemType: "Sentinel2L1C"}
	_, err = GetAsset(aOptions, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)
}

func TestGetMetadataBadKey(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)
	context.PlanetKey = "garbage"
	aOptions := MetadataOptions{ID: "foobar123", Tides: true, ItemType: "PSOrthoTile"}
	_, err := GetMetadata(aOptions, &context)
	assert.NotNil(t, err, "Expected invalid API key to fail, but it succeeded.")
	if httpErr, ok := err.(util.HTTPErr); err != nil && !ok {
		t.Errorf("Expected an HTTPErr, got a %T", err)
	} else {
		assert.Equal(t, 401, httpErr.Status, "Expected error 401 but got %v", httpErr.Status)
	}
}

func TestGetMetadataSentinel(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)

	options := MetadataOptions{
		ID:       "S2A_MSIL1C_20160513T183921_N0204_R070_T11SKD_20160513T185132",
		ItemType: "Sentinel2L1C",
	}

	_, err := GetMetadata(options, &context)
	assert.Nil(t, err, "Expected request to succeed; received: %v", err)
}

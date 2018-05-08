package model

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/geojson-go/geojson"
)

// General test mocks and utils

var mockPolygon = geojson.NewPolygon([][][]float64{[][]float64{
	[]float64{30, 10}, []float64{40, 40}, []float64{20, 40}, []float64{10, 20}, []float64{30, 10},
}})

var mockBasicBrokerResult = BasicBrokerResult{
	AcquiredDate: time.Unix(123, 0),
	CloudCover:   50.123,
	FileFormat:   JPEG2000,
	Geometry:     mockPolygon,
	ID:           "test-id-123",
	Resolution:   10.123,
	SensorName:   "test-sensor",
}

var mockPlanetAssetMetadata = PlanetAssetMetadata{
	AssetURL:      url.URL{Scheme: "https", Host: "example.localhost", Path: "/asset.JP2"},
	ActivationURL: url.URL{Scheme: "https", Host: "example.localhost", Path: "/activate"},
	ExpiresAt:     time.Unix(123, 0),
	Permissions:   []string{"a", "b", "c"},
	Status:        "active",
	Type:          "test",
}

var mockTidesData = TidesData{
	Current: 123.123,
	Min24h:  111.111,
	Max24h:  222.222,
}

var mockLandsatS3Bands = LandsatS3Bands{
	Coastal:      url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B1.TIF"},
	Blue:         url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B2.TIF"},
	Green:        url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B3.TIF"},
	Red:          url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B4.TIF"},
	NIR:          url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B5.TIF"},
	SWIR1:        url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B6.TIF"},
	SWIR2:        url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B7.TIF"},
	Panchromatic: url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B8.TIF"},
	Cirrus:       url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B9.TIF"},
	TIRS1:        url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B10.TIF"},
	TIRS2:        url.URL{Scheme: "https", Host: "example.localhost", Path: "LC8TEST123_B11.TIF"},
}

func assertFeatureContainsBasicBrokerResult(t *testing.T, feature *geojson.Feature, result BasicBrokerResult) {
	assert.Equal(t, result.ID, feature.IDStr())
	assert.Equal(t, result.SensorName, feature.PropertyString("sensorName"))
	assert.Equal(t, result.AcquiredDate.Format(time.RFC3339), feature.PropertyString("acquiredDate"))
	assert.Equal(t, result.CloudCover, feature.PropertyFloat("cloudCover"))
	assert.Equal(t, result.Resolution, feature.PropertyFloat("resolution"))
}

func assertFeatureContainsPlanetAssetMetadata(t *testing.T, feature *geojson.Feature, data PlanetAssetMetadata) {
	assert.Equal(t, data.AssetURL.String(), feature.PropertyString("location"))
	assert.Equal(t, data.ExpiresAt.Format(time.RFC3339), feature.PropertyString("expires_at"))
	assert.Equal(t, data.Permissions, feature.PropertyStringSlice("permissions"))
	assert.Equal(t, data.Status, feature.PropertyString("status"))
	assert.Equal(t, data.Type, feature.PropertyString("type"))
}

func assertFeatureContainsTidesData(t *testing.T, feature *geojson.Feature, tides TidesData) {
	assert.Equal(t, tides.Current, feature.PropertyFloat("currentTide"))
	assert.Equal(t, tides.Min24h, feature.PropertyFloat("minimumTide24Hours"))
	assert.Equal(t, tides.Max24h, feature.PropertyFloat("maximumTide24Hours"))
}

func assertFeatureContainsLandsatBands(t *testing.T, feature *geojson.Feature, bands LandsatS3Bands) {
	assert.IsType(t, map[string]string{}, feature.Properties["bands"])
	featureBands := feature.Properties["bands"].(map[string]string)

	assert.Equal(t, bands.Coastal.String(), featureBands["coastal"])
	assert.Equal(t, bands.Blue.String(), featureBands["blue"])
	assert.Equal(t, bands.Green.String(), featureBands["green"])
	assert.Equal(t, bands.Red.String(), featureBands["red"])
	assert.Equal(t, bands.NIR.String(), featureBands["nir"])
	assert.Equal(t, bands.SWIR1.String(), featureBands["swir1"])
	assert.Equal(t, bands.SWIR2.String(), featureBands["swir2"])
	assert.Equal(t, bands.Panchromatic.String(), featureBands["panchromatic"])
	assert.Equal(t, bands.Cirrus.String(), featureBands["cirrus"])
	assert.Equal(t, bands.TIRS1.String(), featureBands["tirs1"])
	assert.Equal(t, bands.TIRS2.String(), featureBands["tirs2"])
}

// Actual tests

func TestBasicBrokerResult_GeoJSONFeature(t *testing.T) {
	// Mock
	result := mockBasicBrokerResult

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, feature)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assert.Nil(t, feature.Bbox.Valid())
}

func TestPlanetActivateableBrokerResult_GeoJSONFeature_NoTides(t *testing.T) {
	// Mock
	result := PlanetActivateableBrokerResult{
		BasicBrokerResult:   mockBasicBrokerResult,
		PlanetAssetMetadata: mockPlanetAssetMetadata,
	}

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, feature)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assertFeatureContainsPlanetAssetMetadata(t, feature, mockPlanetAssetMetadata)
	assert.Empty(t, feature.PropertyString("currentTide"))
	assert.Empty(t, feature.PropertyString("minimumTide24Hours"))
	assert.Empty(t, feature.PropertyString("maximumTide24Hours"))
	assert.Nil(t, feature.Bbox.Valid())
}

func TestPlanetActivateableBrokerResult_GeoJSONFeature_WithTides(t *testing.T) {
	// Mock
	result := PlanetActivateableBrokerResult{
		BasicBrokerResult:   mockBasicBrokerResult,
		PlanetAssetMetadata: mockPlanetAssetMetadata,
		TidesData:           &mockTidesData,
	}

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, feature)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assertFeatureContainsPlanetAssetMetadata(t, feature, mockPlanetAssetMetadata)
	assertFeatureContainsTidesData(t, feature, mockTidesData)
	assert.Nil(t, feature.Bbox.Valid())
}

func TestPlanetLandsatBrokerResult_GeoJsonFeature_NoTides(t *testing.T) {
	// Mock
	result := PlanetLandsatBrokerResult{
		BasicBrokerResult: mockBasicBrokerResult,
		LandsatS3Bands:    mockLandsatS3Bands,
	}

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assertFeatureContainsLandsatBands(t, feature, mockLandsatS3Bands)
	assert.Empty(t, feature.PropertyString("currentTide"))
	assert.Empty(t, feature.PropertyString("minimumTide24Hours"))
	assert.Empty(t, feature.PropertyString("maximumTide24Hours"))
	assert.Nil(t, feature.Bbox.Valid())
}

func TestPlanetLandsatBrokerResult_GeoJsonFeature_WithTides(t *testing.T) {
	// Mock
	result := PlanetLandsatBrokerResult{
		BasicBrokerResult: mockBasicBrokerResult,
		LandsatS3Bands:    mockLandsatS3Bands,
		TidesData:         &mockTidesData,
	}

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assertFeatureContainsLandsatBands(t, feature, mockLandsatS3Bands)
	assertFeatureContainsTidesData(t, feature, mockTidesData)
	assert.Nil(t, feature.Bbox.Valid())
}

func TestIndexedLandsatBrokerResult_GeoJsonFeature_NoTides(t *testing.T) {
	// Mock
	result := IndexedLandsatBrokerResult{
		BasicBrokerResult: mockBasicBrokerResult,
		LandsatS3Bands:    mockLandsatS3Bands,
	}

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assertFeatureContainsLandsatBands(t, feature, mockLandsatS3Bands)
	assert.Empty(t, feature.PropertyString("currentTide"))
	assert.Empty(t, feature.PropertyString("minimumTide24Hours"))
	assert.Empty(t, feature.PropertyString("maximumTide24Hours"))
	assert.Nil(t, feature.Bbox.Valid())
}

func TestIndexedLandsatBrokerResult_GeoJsonFeature_WithTides(t *testing.T) {
	// Mock
	result := IndexedLandsatBrokerResult{
		BasicBrokerResult: mockBasicBrokerResult,
		LandsatS3Bands:    mockLandsatS3Bands,
		TidesData:         &mockTidesData,
	}

	// Tested code
	feature, err := result.GeoJSONFeature()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	assertFeatureContainsLandsatBands(t, feature, mockLandsatS3Bands)
	assertFeatureContainsTidesData(t, feature, mockTidesData)
	assert.Nil(t, feature.Bbox.Valid())
}

func TestMultiBrokerResult_GeoJSONFeatureCollection(t *testing.T) {
	// Mock
	result := MultiBrokerResult{
		FeatureCreators: []GeoJSONFeatureCreator{mockBasicBrokerResult, mockBasicBrokerResult, mockBasicBrokerResult},
	}

	// Tested coordinate
	fc, err := result.GeoJSONFeatureCollection()

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, fc)
	assert.Len(t, fc.Features, 3)
	for _, feature := range fc.Features {
		assertFeatureContainsBasicBrokerResult(t, feature, mockBasicBrokerResult)
	}
}

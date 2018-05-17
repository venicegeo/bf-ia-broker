package model

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/geojson-go/geojson"
)

func TestPlanetAssetMetadata_Apply(t *testing.T) {
	// Mock
	feature := geojson.NewFeature(nil, "test-id", nil)
	activationURL, _ := url.Parse("https://example.localdomain/path/to/activate")
	assetURL, _ := url.Parse("https://example.localdomain/path/to/asset.JP2")

	data := PlanetAssetMetadata{
		ActivationURL: *activationURL,
		AssetURL:      *assetURL,
		ExpiresAt:     time.Unix(123, 0),
		Permissions:   []string{"a", "b", "c"},
		Status:        "active",
		Type:          "test",
	}

	// Tested code
	err := data.Apply(feature)

	// Asserts
	assert.Nil(t, err)
	assert.Equal(t, time.Unix(123, 0).Format(PlanetTimeFormat), feature.PropertyString("expires_at"))
	assert.Equal(t, "https://example.localdomain/path/to/asset.JP2", feature.PropertyString("location"))
	assert.Equal(t, []string{"a", "b", "c"}, feature.PropertyStringSlice("permissions"))
	assert.Equal(t, "active", feature.PropertyString("status"))
	assert.Equal(t, "test", feature.PropertyString("type"))
}

func TestTidesData_Apply(t *testing.T) {
	// Mock
	feature := geojson.NewFeature(nil, "test-id", nil)
	data := TidesData{
		Current: 123.123,
		Min24h:  111.111,
		Max24h:  222.222,
	}

	// Tested code
	err := data.Apply(feature)

	// Asserts
	assert.Nil(t, err)
	assert.Equal(t, 123.123, feature.PropertyFloat("currentTide"))
	assert.Equal(t, 111.111, feature.PropertyFloat("minimumTide24Hours"))
	assert.Equal(t, 222.222, feature.PropertyFloat("maximumTide24Hours"))
}

func TestNewLandsatS3Bands_Success(t *testing.T) {
	// Tested code
	bands, err := NewLandsatS3Bands("https://s3.example.localdomain/landsat/", "LC8TEST123")

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, bands)
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B1.TIF", bands.Coastal.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B2.TIF", bands.Blue.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B3.TIF", bands.Green.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B4.TIF", bands.Red.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B5.TIF", bands.NIR.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B6.TIF", bands.SWIR1.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B7.TIF", bands.SWIR2.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B8.TIF", bands.Panchromatic.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B9.TIF", bands.Cirrus.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B10.TIF", bands.TIRS1.String())
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B11.TIF", bands.TIRS2.String())
}

func TestNewLandsatS3Bands_Error(t *testing.T) {
	// Tested code
	_, err := NewLandsatS3Bands("", "LC8TEST123")

	// Asserts
	assert.NotNil(t, err)
}

func TestLandsatS3Bands_Apply(t *testing.T) {
	// Mock
	feature := geojson.NewFeature(nil, "test-id", nil)
	bands, _ := NewLandsatS3Bands("https://s3.example.localdomain/landsat/", "LC8TEST123")

	// Tested code
	err := bands.Apply(feature)

	// Asserts
	assert.Nil(t, err)
	assert.IsType(t, map[string]string{}, feature.Properties["bands"])
	featureBands := feature.Properties["bands"].(map[string]string)

	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B1.TIF", featureBands["coastal"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B2.TIF", featureBands["blue"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B3.TIF", featureBands["green"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B4.TIF", featureBands["red"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B5.TIF", featureBands["nir"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B6.TIF", featureBands["swir1"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B7.TIF", featureBands["swir2"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B8.TIF", featureBands["panchromatic"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B9.TIF", featureBands["cirrus"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B10.TIF", featureBands["tirs1"])
	assert.Equal(t, "https://s3.example.localdomain/landsat/LC8TEST123_B11.TIF", featureBands["tirs2"])
}

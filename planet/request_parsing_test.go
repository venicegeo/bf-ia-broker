package planet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/geojson-go/geojson"
)

func TestPlanetAssetMetadataFromAssets_Success(t *testing.T) {
	// Mock
	mockExpiresAt := time.Unix(123, 0).UTC()
	validAssets := Assets{
		Analytic: Asset{
			Location:    "https://example.localdomain/path/to/asset.JP2",
			ExpiresAt:   mockExpiresAt.Format(model.PlanetTimeFormat),
			Permissions: []string{"a", "b", "c"},
			Status:      "active",
			Type:        "test",
			Links: Links{
				Activate: "https://example.localdomain/path/to/activate",
			},
		},
	}

	// Tested code
	data, err := planetAssetMetadataFromAssets(validAssets)

	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "https://example.localdomain/path/to/asset.JP2", data.AssetURL.String())
	assert.Equal(t, mockExpiresAt, data.ExpiresAt)
	assert.Equal(t, []string{"a", "b", "c"}, data.Permissions)
	assert.Equal(t, "active", data.Status)
	assert.Equal(t, "test", data.Type)
	assert.Equal(t, "https://example.localdomain/path/to/activate", data.ActivationURL.String())
}

func TestPlanetAssetMetadataFromAssets_Error(t *testing.T) {
	// Mock
	emptyAssets := Assets{}
	badTimeAssets := Assets{
		Analytic: Asset{
			Type:      "REOrthoTile",
			Location:  "https://example.localdomain/path/to/asset.JP2",
			ExpiresAt: "this-is-not-a-time-format",
			Links: Links{
				Activate: "https://example.localdomain/path/to/activate",
			},
		},
	}
	noLocationAssets := Assets{
		Analytic: Asset{
			Type:      "REOrthoTile",
			Location:  "",
			ExpiresAt: time.Unix(123, 0).Format(model.PlanetTimeFormat),
			Links: Links{
				Activate: "https://example.localdomain/path/to/activate",
			},
		},
	}
	noActivationAssets := Assets{
		Analytic: Asset{
			Type:      "REOrthoTile",
			Location:  "https://example.localdomain/path/to/asset.JP2",
			ExpiresAt: time.Unix(123, 0).Format(model.PlanetTimeFormat),
			Links:     Links{},
		},
	}

	// Tested code
	emptyResult, emptyErr := planetAssetMetadataFromAssets(emptyAssets)
	_, badTimeErr := planetAssetMetadataFromAssets(badTimeAssets)
	_, noLocationErr := planetAssetMetadataFromAssets(noLocationAssets)
	_, noActivationErr := planetAssetMetadataFromAssets(noActivationAssets)

	// Asserts
	assert.Nil(t, emptyResult)
	assert.Nil(t, emptyErr)
	assert.NotNil(t, badTimeErr)
	assert.NotNil(t, noLocationErr)
	assert.NotNil(t, noActivationErr)
}

func TestBasicBrokerResultFromPlanetFeature_MissingCloudCover(t *testing.T) {
	// Mock
	mockAcquired := time.Unix(123, 0).UTC()
	feature := geojson.NewFeature([]float64{}, "test-id", map[string]interface{}{
		"acquired": mockAcquired.Format(model.PlanetTimeFormat),
	})

	// Tested Code
	result, err := basicBrokerResultFromPlanetFeature(feature, model.JPEG2000)

	// Asserts
	assert.NotNil(t, result)
	assert.Nil(t, err)
	assert.Equal(t, -1.0, result.CloudCover)
}

func TestBasicBrokerResultFromPlanetFeature_ZeroCloudCover(t *testing.T) {
	// Mock
	mockAcquired := time.Unix(123, 0).UTC()
	feature := geojson.NewFeature([]float64{}, "test-id", map[string]interface{}{
		"acquired":    mockAcquired.Format(model.PlanetTimeFormat),
		"cloud_cover": 0.0,
	})

	// Tested Code
	result, err := basicBrokerResultFromPlanetFeature(feature, model.JPEG2000)

	// Asserts
	assert.NotNil(t, result)
	assert.Nil(t, err)
	assert.Equal(t, 0.0, result.CloudCover)
}

func TestBasicBrokerResultFromPlanetFeature_NonZeroCloudCover(t *testing.T) {
	// Mock
	mockAcquired := time.Unix(123, 0).UTC()
	feature := geojson.NewFeature([]float64{}, "test-id", map[string]interface{}{
		"acquired":    mockAcquired.Format(model.PlanetTimeFormat),
		"cloud_cover": 0.123,
	})

	// Tested Code
	result, err := basicBrokerResultFromPlanetFeature(feature, model.JPEG2000)

	// Asserts
	assert.NotNil(t, result)
	assert.Nil(t, err)
	assert.Equal(t, 12.3, result.CloudCover)
}

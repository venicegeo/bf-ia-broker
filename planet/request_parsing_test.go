package planet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlanetAssetMetadataFromAssets_Success(t *testing.T) {
	// Mock
	validAssets := Assets{
		Analytic: Asset{
			Location:    "https://example.localdomain/path/to/asset.JP2",
			ExpiresAt:   time.Unix(123, 0).Format(time.RFC3339),
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
	assert.Equal(t, time.Unix(123, 0), data.ExpiresAt)
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
			Location:  "https://example.localdomain/path/to/asset.JP2",
			ExpiresAt: "this-is-not-a-time-format",
			Links: Links{
				Activate: "https://example.localdomain/path/to/activate",
			},
		},
	}
	noLocationAssets := Assets{
		Analytic: Asset{
			Location:  "",
			ExpiresAt: time.Unix(123, 0).Format(time.RFC3339),
			Links: Links{
				Activate: "https://example.localdomain/path/to/activate",
			},
		},
	}
	noActivationAssets := Assets{
		Analytic: Asset{
			Location:  "https://example.localdomain/path/to/asset.JP2",
			ExpiresAt: time.Unix(123, 0).Format(time.RFC3339),
			Links:     Links{},
		},
	}

	// Tested code
	_, emptyErr := planetAssetMetadataFromAssets(emptyAssets)
	_, badTimeErr := planetAssetMetadataFromAssets(badTimeAssets)
	_, noLocationErr := planetAssetMetadataFromAssets(noLocationAssets)
	_, noActivationErr := planetAssetMetadataFromAssets(noActivationAssets)

	// Asserts
	assert.NotNil(t, emptyErr)
	assert.NotNil(t, badTimeErr)
	assert.NotNil(t, noLocationErr)
	assert.NotNil(t, noActivationErr)
}

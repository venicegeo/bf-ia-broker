package model

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

// PlanetAssetMetadata is a mixin containing metadata for an activate-able
// asset retrieved from the Planet API
type PlanetAssetMetadata struct {
	AssetURL      url.URL
	ActivationURL url.URL
	ExpiresAt     time.Time
	Permissions   []string
	Status        string
	Type          string
}

// Apply implements the GeoJSONFeatureMixin interface
func (pam PlanetAssetMetadata) Apply(feature *geojson.Feature) error {
	feature.Properties["expires_at"] = pam.ExpiresAt.Format(PlanetTimeFormat)
	feature.Properties["location"] = pam.AssetURL.String()
	feature.Properties["permissions"] = pam.Permissions
	feature.Properties["status"] = pam.Status
	feature.Properties["type"] = pam.Type
	return nil
}

// TidesData is a mixin containing optional tides data from bf-tideprediction
type TidesData struct {
	Current float64
	Min24h  float64
	Max24h  float64
}

// Apply implements the GeoJSONFeatureMixin interface
func (td TidesData) Apply(feature *geojson.Feature) error {
	feature.Properties["currentTide"] = td.Current
	feature.Properties["minimumTide24Hours"] = td.Min24h
	feature.Properties["maximumTide24Hours"] = td.Max24h
	return nil
}

// LandsatS3Bands is a mixin containing data about the bands of a Landsat8 result
type LandsatS3Bands struct {
	Coastal      url.URL
	Blue         url.URL
	Green        url.URL
	Red          url.URL
	NIR          url.URL
	SWIR1        url.URL
	SWIR2        url.URL
	Panchromatic url.URL
	Cirrus       url.URL
	TIRS1        url.URL
	TIRS2        url.URL
}

type landsatSuffixDestination struct {
	BandSuffix  string
	Destination *url.URL
}

// NewLandsatS3Bands creates a new LandsatS3Bands by inferring the bands based on
// Landsat bucket info
func NewLandsatS3Bands(bucketFolderURL string, id string) (*LandsatS3Bands, error) {
	baseURL, err := url.Parse(bucketFolderURL)
	if baseURL == nil || baseURL.String() == "" {
		err = errors.New("No base Landsat S3 bucker folder could be parsed")
	}
	if err != nil {
		return nil, err
	}

	bands := LandsatS3Bands{}

	suffixes := []landsatSuffixDestination{
		landsatSuffixDestination{"B1", &bands.Coastal},
		landsatSuffixDestination{"B2", &bands.Blue},
		landsatSuffixDestination{"B3", &bands.Green},
		landsatSuffixDestination{"B4", &bands.Red},
		landsatSuffixDestination{"B5", &bands.NIR},
		landsatSuffixDestination{"B6", &bands.SWIR1},
		landsatSuffixDestination{"B7", &bands.SWIR2},
		landsatSuffixDestination{"B8", &bands.Panchromatic},
		landsatSuffixDestination{"B9", &bands.Cirrus},
		landsatSuffixDestination{"B10", &bands.TIRS1},
		landsatSuffixDestination{"B11", &bands.TIRS2},
	}

	for _, dest := range suffixes {
		filename := fmt.Sprintf("%s_%s.TIF", id, dest.BandSuffix)
		fileURL, _ := url.Parse("./" + filename)
		*dest.Destination = *baseURL.ResolveReference(fileURL)
	}

	return &bands, nil
}

// Apply implements the GeoJSONFeatureMixin interface
func (lsb LandsatS3Bands) Apply(feature *geojson.Feature) error {
	feature.Properties["bands"] = map[string]string{
		"coastal":      lsb.Coastal.String(),
		"blue":         lsb.Blue.String(),
		"green":        lsb.Green.String(),
		"red":          lsb.Red.String(),
		"nir":          lsb.NIR.String(),
		"swir1":        lsb.SWIR1.String(),
		"swir2":        lsb.SWIR2.String(),
		"panchromatic": lsb.Panchromatic.String(),
		"cirrus":       lsb.Cirrus.String(),
		"tirs1":        lsb.TIRS1.String(),
		"tirs2":        lsb.TIRS2.String(),
	}
	return nil
}

package model

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
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
	feature.Properties["expires_at"] = pam.ExpiresAt.Format(StandardTimeLayout)
	feature.Properties["location"] = pam.AssetURL.String()
	feature.Properties["permissions"] = pam.Permissions
	feature.Properties["status"] = pam.Status
	feature.Properties["type"] = pam.Type
	feature.Properties["srcHorizontalAccuracy"] = "<10m RMSE"
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
func NewLandsatS3Bands(bucketFolderURL string, filePrefix string) (*LandsatS3Bands, error) {
	baseURL, err := url.Parse(bucketFolderURL)
	if baseURL == nil || baseURL.String() == "" {
		err = errors.New("No base Landsat S3 bucket folder could be parsed")
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
		filename := fmt.Sprintf("%s_%s.TIF", filePrefix, dest.BandSuffix)
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
	feature.Properties["srcHorizontalAccuracy"] = "12m CE90"
	return nil
}

// SentinelS3Bands is a mixin containing data about the bands of a Sentinel-2 result
type SentinelS3Bands struct {
	Coastal    url.URL
	Blue       url.URL
	Green      url.URL
	Red        url.URL
	RedEdge1   url.URL
	RedEdge2   url.URL
	RedEdge3   url.URL
	NIR        url.URL
	WaterVapor url.URL
	Cirrus     url.URL
	SWIR1      url.URL
	SWIR2      url.URL
}

type sentinelFilenameDestination struct {
	FileName    string
	Destination *url.URL
}

// https://earth.esa.int/web/sentinel/user-guides/sentinel-2-msi/naming-convention
// TODO: add support for old-style product IDs (which do not contain MGRS info in them)
var sentinelIDPattern = regexp.MustCompile("S2(A|B)_MSIL1C_([0-9]{4})([0-9]{2})([0-9]{2})T[0-9]+_[A-Z0-9]+_[A-Z0-9]+_T([0-9]+)([A-Z])([A-Z]+)_[0-9]{8}T[0-9]")

// Inputs: hostUrl, mgrs1, mgrs2, mgrs3, year, month, day, filename
const sentinelAWSURLFormat = "%s/tiles/%s/%s/%s/%d/%d/%d/0/%s"

// NewSentinelS3Bands creates a SentinelS3Bands object based on the given sentinel ID
func NewSentinelS3Bands(bucketFolderURL string, sentinelID string) (*SentinelS3Bands, error) {
	if !sentinelIDPattern.MatchString(sentinelID) {
		return nil, fmt.Errorf("Product ID but did not match expected Sentinel-2 format: %s", sentinelID)
	}

	m := sentinelIDPattern.FindStringSubmatch(sentinelID)
	m = m[2:] // Skip over whole string match and satellite A/B match

	var (
		year, month, day int
		mgrs1            = m[3]
		mgrs2            = m[4]
		mgrs3            = m[5]
	)
	year, _ = strconv.Atoi(m[0])
	month, _ = strconv.Atoi(m[1])
	day, _ = strconv.Atoi(m[2])

	bands := SentinelS3Bands{}
	fileNameMap := []sentinelFilenameDestination{
		sentinelFilenameDestination{"B01.jp2", &bands.Coastal},
		sentinelFilenameDestination{"B02.jp2", &bands.Blue},
		sentinelFilenameDestination{"B03.jp2", &bands.Green},
		sentinelFilenameDestination{"B04.jp2", &bands.Red},
		sentinelFilenameDestination{"B05.jp2", &bands.RedEdge1},
		sentinelFilenameDestination{"B06.jp2", &bands.RedEdge2},
		sentinelFilenameDestination{"B07.jp2", &bands.RedEdge3},
		sentinelFilenameDestination{"B08.jp2", &bands.NIR},
		sentinelFilenameDestination{"B09.jp2", &bands.WaterVapor},
		sentinelFilenameDestination{"B10.jp2", &bands.Cirrus},
		sentinelFilenameDestination{"B11.jp2", &bands.SWIR1},
		sentinelFilenameDestination{"B12.jp2", &bands.SWIR2},
	}

	for _, dest := range fileNameMap {
		s3URLString := fmt.Sprintf(sentinelAWSURLFormat, bucketFolderURL, mgrs1, mgrs2, mgrs3, year, month, day, dest.FileName)
		s3URL, _ := url.Parse(s3URLString)
		*dest.Destination = *s3URL
	}

	return &bands, nil
}

// Apply implements the GeoJSONFeatureMixin interface
func (ssb SentinelS3Bands) Apply(feature *geojson.Feature) error {
	feature.Properties["bands"] = map[string]string{
		"coastal":    ssb.Coastal.String(),
		"blue":       ssb.Blue.String(),
		"green":      ssb.Green.String(),
		"red":        ssb.Red.String(),
		"rededge1":   ssb.RedEdge1.String(),
		"rededge2":   ssb.RedEdge2.String(),
		"rededge3":   ssb.RedEdge3.String(),
		"nir":        ssb.NIR.String(),
		"watervapor": ssb.WaterVapor.String(),
		"cirrus":     ssb.Cirrus.String(),
		"swir1":      ssb.SWIR1.String(),
		"swir2":      ssb.SWIR2.String(),
	}
	feature.Properties["srcHorizontalAccuracy"] = "12.5m with GCPs"
	return nil
}

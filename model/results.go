package model

import (
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

// BasicBrokerResult holds the fields common to all bf-ia-broker single results
type BasicBrokerResult struct {
	ID           string
	Geometry     interface{}
	CloudCover   float64
	Resolution   float64
	AcquiredDate time.Time
	SensorName   string
	FileFormat   BrokerFileFormat
}

// GeoJSONFeature implements the GeoJSONFeatureCreator interface
func (br BasicBrokerResult) GeoJSONFeature() (*geojson.Feature, error) {
	f := geojson.NewFeature(br.Geometry, br.ID, map[string]interface{}{
		"cloudCover":   br.CloudCover,
		"resolution":   br.Resolution,
		"acquiredDate": br.AcquiredDate.Format(PlanetTimeFormat),
		"sensorName":   br.SensorName,
	})
	f.Bbox = f.ForceBbox()
	return f, nil
}

// BrokerSearchResult is contains a barebones broker search result -- basic
// data, plus optional tides data
type BrokerSearchResult struct {
	BasicBrokerResult
	*TidesData
}

// GeoJSONFeature implements the GeoJSONFeatureCreator interface
func (result BrokerSearchResult) GeoJSONFeature() (*geojson.Feature, error) {
	feature, err := result.BasicBrokerResult.GeoJSONFeature()
	if err != nil {
		return nil, err
	}

	if result.TidesData != nil {
		err = result.TidesData.Apply(feature)
		if err != nil {
			return nil, err
		}
	}

	return feature, nil
}

// PlanetActivateableBrokerResult represents a Planet result that may require
// asset activation (RapidEye, PlanetScope, Sentinel-2)
type PlanetActivateableBrokerResult struct {
	BasicBrokerResult
	PlanetAssetMetadata
	*TidesData
}

// GeoJSONFeature implements the GeoJSONFeatureCreator interface
func (result PlanetActivateableBrokerResult) GeoJSONFeature() (*geojson.Feature, error) {
	feature, err := result.BasicBrokerResult.GeoJSONFeature()
	if err != nil {
		return nil, err
	}

	err = result.PlanetAssetMetadata.Apply(feature)
	if err != nil {
		return nil, err
	}

	if result.TidesData != nil {
		err = result.TidesData.Apply(feature)
		if err != nil {
			return nil, err
		}
	}

	return feature, nil
}

// PlanetLandsatBrokerResult represents a Planet result referencing an external
// Landsat8 archive, requiring no activation
type PlanetLandsatBrokerResult struct {
	BasicBrokerResult
	LandsatS3Bands
	*TidesData
}

// GeoJSONFeature implements the GeoJSONFeatureCreator interface
func (result PlanetLandsatBrokerResult) GeoJSONFeature() (*geojson.Feature, error) {
	feature, err := result.BasicBrokerResult.GeoJSONFeature()
	if err != nil {
		return nil, err
	}

	err = result.LandsatS3Bands.Apply(feature)
	if err != nil {
		return nil, err
	}

	if result.TidesData != nil {
		err = result.TidesData.Apply(feature)
		if err != nil {
			return nil, err
		}
	}

	return feature, nil
}

// IndexedLandsatBrokerResult represents a local-index result containing Landsat8 data
type IndexedLandsatBrokerResult struct {
	BasicBrokerResult
	LandsatS3Bands
	*TidesData
}

// GeoJSONFeature implements the GeoJSONFeatureCreator interface
func (result IndexedLandsatBrokerResult) GeoJSONFeature() (*geojson.Feature, error) {
	feature, err := result.BasicBrokerResult.GeoJSONFeature()
	if err != nil {
		return nil, err
	}

	err = result.LandsatS3Bands.Apply(feature)
	if err != nil {
		return nil, err
	}

	if result.TidesData != nil {
		err = result.TidesData.Apply(feature)
		if err != nil {
			return nil, err
		}
	}

	return feature, nil
}

// MultiBrokerResult is a container type for bundling multiple results together,
// e.g. as results from a search endpoint
type MultiBrokerResult struct {
	FeatureCreators []GeoJSONFeatureCreator
}

// GeoJSONFeatureCollection implements the GeoJSONFeatureCollectionCreator interface
func (result MultiBrokerResult) GeoJSONFeatureCollection() (*geojson.FeatureCollection, error) {
	var err error
	features := make([]*geojson.Feature, len(result.FeatureCreators))
	for i, creator := range result.FeatureCreators {
		features[i], err = creator.GeoJSONFeature()
		if err != nil {
			return nil, err
		}
	}

	return geojson.NewFeatureCollection(features), nil
}

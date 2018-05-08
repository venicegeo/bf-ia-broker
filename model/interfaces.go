package model

import "github.com/venicegeo/geojson-go/geojson"

// BrokerFileFormat is an enum type for recognized file input types
type BrokerFileFormat string

// GeoTIFF corresponds to .TIF files with geospatial info
const GeoTIFF BrokerFileFormat = "geotiff"

// JPEG2000 corresponds to .JP2 files
const JPEG2000 BrokerFileFormat = "jpeg2000"

// GeoJSONFeatureCreator is an interface for data that can convert itself to a GeoJSON feature
type GeoJSONFeatureCreator interface {
	GeoJSONFeature() (*geojson.Feature, error)
}

// GeoJSONFeatureCollectionCreator is an interface for data that can convert itself to a GeoJSON feature collection
type GeoJSONFeatureCollectionCreator interface {
	GeoJSONFeatureCollection() (*geojson.FeatureCollection, error)
}

// GeoJSONFeatureMixin is an interface for data that can be used to augment an existing GeoJSON feature
type GeoJSONFeatureMixin interface {
	Apply(*geojson.Feature) error
}

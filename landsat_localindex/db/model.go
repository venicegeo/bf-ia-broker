package db

import (
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

// SingleOrMultiPolygon is an interface that supports both geojson.Polygon and geojson.MultiPolygon
type SingleOrMultiPolygon interface {
	ForceBbox() geojson.BoundingBox
	Map() map[string]interface{}
	String() string
	WKT() string
}

// LandsatLocalIndexScene contains data pertaining to a Landsat8 scene indexed locally
type LandsatLocalIndexScene struct {
	ProductID       string
	AcquisitionDate time.Time
	CloudCover      float64
	SceneURLString  string
	Bounds          SingleOrMultiPolygon
	BoundingBox     geojson.BoundingBox
}

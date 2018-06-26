package db

import (
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

// LandsatLocalIndexScene contains data pertaining to a Landsat8 scene indexed locally
type LandsatLocalIndexScene struct {
	ProductID       string
	AcquisitionDate time.Time
	CloudCover      float64
	SceneURLString  string
	Bounds          *geojson.Polygon
	BoundingBox     geojson.BoundingBox
}

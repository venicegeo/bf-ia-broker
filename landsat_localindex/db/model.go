package db

import (
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

type LandsatLocalIndexScene struct {
	ProductID       string
	AcquisitionDate time.Time
	CloudCover      float64
	SceneURLString  string
	Bounds          *geojson.Polygon
}

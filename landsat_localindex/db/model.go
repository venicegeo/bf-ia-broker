package db

import "time"

type LandsatLocalIndexScene struct {
	ProductID       string
	AcquisitionDate time.Time
	CloudCover      float64
	SceneURLString  string
	Bounds          []byte
}

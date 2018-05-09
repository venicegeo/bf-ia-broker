package tides

import (
	"time"

	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

// Context is the context for this operation
type Context struct {
	TidesURL  string
	sessionID string
}

// AppName returns an empty string
func (c *Context) AppName() string {
	return "bf-ia-broker"
}

// SessionID returns a Session ID, creating one if needed
func (c *Context) SessionID() string {
	if c.sessionID == "" {
		c.sessionID, _ = util.PsuUUID()
	}
	return c.sessionID
}

// LogRootDir returns an empty string
func (c *Context) LogRootDir() string {
	return ""
}

type InputLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	Dtg string  `json:"dtg"`
}

type Input struct {
	Locations []InputLocation `json:"locations"`
}

type OutputData struct {
	MinTide  float64 `json:"minimumTide24Hours"`
	MaxTide  float64 `json:"maximumTide24Hours"`
	CurrTide float64 `json:"currentTide"`
}

type OutputLocation struct {
	Lat     float64    `json:"lat"`
	Lon     float64    `json:"lon"`
	Dtg     string     `json:"dtg"`
	Results OutputData `json:"results"`
}

type Output struct {
	Locations []OutputLocation `json:"locations"`
}

func InputLocationForFeature(feature *geojson.Feature, acquiredDate time.Time) InputLocation {
	center := feature.ForceBbox().Centroid()
	return InputLocation{
		Lon: center.Coordinates[0],
		Lat: center.Coordinates[1],
		Dtg: acquiredDate.Format("2006-01-02-15-04"),
	}
}

func InputForBasicBrokerResults(results []model.BasicBrokerResult) (*Input, error) {
	locations := make([]InputLocation, len(results))
	for i, result := range results {
		feature, err := result.GeoJSONFeature()
		if err != nil {
			return nil, err
		}
		locations[i] = InputLocationForFeature(feature, result.AcquiredDate)
	}
	return &Input{Locations: locations}, nil
}

func OutputToTidesData(output Output) []model.TidesData {
	tidesData := make([]model.TidesData, len(output.Locations))
	for i, location := range output.Locations {
		tidesData[i] = model.TidesData{
			Current: location.Results.CurrTide,
			Min24h:  location.Results.MinTide,
			Max24h:  location.Results.MaxTide,
		}
	}
	return tidesData
}

package landsatlocalindex

import (
	"database/sql"

	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db"
	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/util"
)

// Context is the context for a Planet Labs Operation
type Context struct {
	DB           *sql.DB
	BaseTidesURL string
	sessionID    string
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

func indexedLandsatBrokerResultFromBrokerSearchResult(original model.BrokerSearchResult, sceneURLString string) (*model.IndexedLandsatBrokerResult, error) {
	result := model.IndexedLandsatBrokerResult{BasicBrokerResult: original.BasicBrokerResult, TidesData: original.TidesData}

	bands, err := model.NewLandsatS3Bands(sceneURLString, result.BasicBrokerResult.ID)
	if err != nil {
		return nil, err
	}
	result.LandsatS3Bands = *bands

	return &result, nil
}

func brokerSearchResultFromScene(scene db.LandsatLocalIndexScene) model.BrokerSearchResult {
	return model.BrokerSearchResult{
		BasicBrokerResult: model.BasicBrokerResult{
			ID:           scene.ProductID,
			AcquiredDate: scene.AcquisitionDate,
			CloudCover:   scene.CloudCover,
			Resolution:   0,              // No data available for this
			SensorName:   "Landsat8L1TP", // XXX: hardcoded
			FileFormat:   model.GeoTIFF,
			Geometry:     scene.Bounds,
		},
	}
}

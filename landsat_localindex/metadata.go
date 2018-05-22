package landsatlocalindex

import (
	"database/sql"

	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db"
	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/tides"
)

func getMetadata(tx *sql.Tx, ctx Context, sceneID string, withTides bool) (model.GeoJSONFeatureCreator, error) {
	scene, err := db.GetSceneByID(tx, sceneID)
	if err != nil {
		return nil, err
	}

	result := model.IndexedLandsatBrokerResult{}
	result.BasicBrokerResult = model.BasicBrokerResult{
		ID:           scene.ProductID,
		AcquiredDate: scene.AcquisitionDate,
		CloudCover:   scene.CloudCover,
		Resolution:   0,              // No data available for this
		SensorName:   "Landsat8L1TP", // XXX: hardcoded
		FileFormat:   model.GeoTIFF,
		Geometry:     scene.Bounds,
	}

	bands, err := model.NewLandsatS3Bands(scene.SceneURLString, sceneID)
	if err != nil {
		return nil, err
	}
	result.LandsatS3Bands = *bands

	if withTides {
		tidesContext := &tides.Context{TidesURL: ctx.BaseTidesURL}
		if tidesData, err := tides.GetSingleTidesData(tidesContext, result.BasicBrokerResult); err == nil {
			result.TidesData = tidesData
		} else {
			return nil, err
		}
	}

	return result, nil
}

package landsatlocalindex

import (
	"database/sql"
	"time"

	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db"
	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/tides"
	"github.com/venicegeo/geojson-go/geojson"
)

func discoverScenes(tx *sql.Tx, ctx Context, bbox geojson.BoundingBox,
	maxCloudCover float64, minAcquiredDate time.Time, maxAcquiredDate time.Time, withTides bool) (model.GeoJSONFeatureCollectionCreator, error) {
	scenes, err := db.SearchScenes(tx, bbox, maxCloudCover, minAcquiredDate, maxAcquiredDate)
	if err != nil {
		return nil, err
	}

	searchResults := make([]model.BrokerSearchResult, len(scenes))
	for i, scene := range scenes {
		searchResults[i] = brokerSearchResultFromScene(scene)
	}

	if withTides {
		tidesContext := &tides.Context{TidesURL: ctx.BaseTidesURL}
		if err = tides.AddTidesToSearchResults(tidesContext, searchResults); err != nil {
			return nil, err
		}
	}

	multiResult := model.MultiBrokerResult{
		FeatureCreators: make([]model.GeoJSONFeatureCreator, len(searchResults)),
	}

	for i, result := range searchResults {
		if multiResult.FeatureCreators[i], err = indexedLandsatBrokerResultFromBrokerSearchResult(result, scenes[i].SceneURLString); err != nil {
			return nil, err
		}
	}

	return multiResult, nil
}

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

	searchResult := brokerSearchResultFromScene(*scene)

	if withTides {
		tidesContext := &tides.Context{TidesURL: ctx.BaseTidesURL}
		inPlaceEditableSearchResults := []model.BrokerSearchResult{searchResult}
		if err = tides.AddTidesToSearchResults(tidesContext, inPlaceEditableSearchResults); err == nil {
			searchResult = inPlaceEditableSearchResults[0]
		} else {
			return nil, err
		}
	}

	result, err := indexedLandsatBrokerResultFromBrokerSearchResult(searchResult, scene.SceneURLString)
	if err != nil {
		return nil, err
	}

	return result, nil
}

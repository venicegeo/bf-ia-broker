package planet

import (
	"errors"

	landsat "github.com/venicegeo/bf-ia-broker/landsat_planet"
	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/tides"
	"github.com/venicegeo/geojson-go/geojson"
)

func GetItemWithAssetMetadata(context *Context, options MetadataOptions) (*geojson.Feature, error) {
	var (
		err           error
		searchResult  *model.BrokerSearchResult
		assetMetadata *model.PlanetAssetMetadata
	)
	if searchResult, err = GetPlanetItem(options, context); err != nil {
		return nil, err
	}

	basicResult := searchResult.BasicBrokerResult
	if options.ItemType == "Sentinel2L1C" {
		// Sentinel-2 uses JPEG2000 imagery
		basicResult.FileFormat = model.JPEG2000
	}
	if err != nil {
		return nil, err
	}

	var tidesData *model.TidesData
	if options.Tides {
		tidesContext := tides.Context{TidesURL: context.BaseTidesURL}
		if tidesData, err = tides.GetSingleTidesData(&tidesContext, basicResult); err != nil {
			return nil, err
		}
	}

	var result model.GeoJSONFeatureCreator

	switch options.ItemType {
	case "REOrthoTile", "PSOrthoTile", "Sentinel2L1C", "PSScene4Band":
		// These are sources with activateable imagery hosted by Planet itself
		if assetMetadata, err = GetPlanetAssets(options, context); err != nil {
			return nil, err
		}
		result = model.PlanetActivateableBrokerResult{
			BasicBrokerResult:   basicResult,
			PlanetAssetMetadata: *assetMetadata,
			TidesData:           tidesData,
		}

	case "Landsat8L1G":
		// Landsat imagery is hosted on an external S3 archive
		folderURL, _, err := landsat.GetSceneFolderURL(basicResult.ID, basicResult.DataType)
		if err != nil {
			return nil, err
		}
		landsatBands, err := model.NewLandsatS3Bands(folderURL, basicResult.ID)
		if err != nil {
			return nil, err
		}
		result = model.PlanetLandsatBrokerResult{
			BasicBrokerResult: basicResult,
			LandsatS3Bands:    *landsatBands,
			TidesData:         tidesData,
		}
	default:
		err = errors.New("Unrecognized item type:" + options.ItemType)
	}
	return result.GeoJSONFeature()
}

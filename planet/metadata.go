package planet

import (
	"fmt"

	landsat "github.com/venicegeo/bf-ia-broker/landsat_planet"
	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/util"
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
	tidesData := searchResult.TidesData
	if options.ImagerySource == sentinelFromS3 {
		// Sentinel-2 in S3 uses JPEG2000 imagery
		basicResult.FileFormat = model.JPEG2000
	}
	// TODO: check if Sentinel-2 in Planet returns as GeoTIFF or JPEG2000
	if err != nil {
		return nil, err
	}

	var result model.GeoJSONFeatureCreator

	switch options.ImagerySource {
	case rapidEye, planetScope, sentinelFromPlanet:
		// These are sources with activateable imagery hosted by Planet itself
		if assetMetadata, err = GetPlanetAssets(options, context); err != nil {
			return nil, err
		}
		result = model.PlanetActivateableBrokerResult{
			BasicBrokerResult:   basicResult,
			PlanetAssetMetadata: *assetMetadata,
			TidesData:           tidesData,
		}

	case landsatFromS3:
		// Landsat imagery is hosted on an external S3 archive
		folderURL, prefix, err := landsat.GetSceneFolderURL(basicResult.ID, basicResult.DataType)
		if err != nil {
			return nil, err
		}
		landsatBands, err := model.NewLandsatS3Bands(folderURL, prefix)
		if err != nil {
			return nil, err
		}
		result = model.PlanetLandsatBrokerResult{
			BasicBrokerResult: basicResult,
			LandsatS3Bands:    *landsatBands,
			TidesData:         tidesData,
		}

	case sentinelFromS3:
		// Sentinel-2 imagery is hosted on an external S3 archive
		sentinelBands, err := model.NewSentinelS3Bands(util.GetSentinelHost(), basicResult.ID)
		if err != nil {
			return nil, err
		}
		result = model.PlanetSentinelBrokerResult{
			BasicBrokerResult: basicResult,
			SentinelS3Bands:   *sentinelBands,
			TidesData:         tidesData,
		}

	default:
		return nil, fmt.Errorf("Unrecognized imagery source (%v), type: %s", options.ImagerySource, options.ItemType)
	}
	return result.GeoJSONFeature()
}

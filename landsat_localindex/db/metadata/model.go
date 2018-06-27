package metadata

import (
	"github.com/venicegeo/geojson-go/geojson"
)

// LandsatSceneMetadata contains metadata recovered from a Landsat S3 MTL file
type LandsatSceneMetadata struct {
	Bounds *geojson.Polygon
}

type sceneMTL struct {
	L1MetadataFile struct {
		ProductMetadata struct {
			CornerUpperLeftLon  float64 `json:"CORNER_UL_LON_PRODUCT"`
			CornerUpperLeftLat  float64 `json:"CORNER_UL_LAT_PRODUCT"`
			CornerUpperRightLon float64 `json:"CORNER_UR_LON_PRODUCT"`
			CornerUpperRightLat float64 `json:"CORNER_UR_LAT_PRODUCT"`
			CornerLowerLeftLon  float64 `json:"CORNER_LL_LON_PRODUCT"`
			CornerLowerLeftLat  float64 `json:"CORNER_LL_LAT_PRODUCT"`
			CornerLowerRightLon float64 `json:"CORNER_LR_LON_PRODUCT"`
			CornerLowerRightLat float64 `json:"CORNER_LR_LAT_PRODUCT"`
		} `json:"PRODUCT_METADATA"`
	} `json:"L1_METADATA_FILE"`
}

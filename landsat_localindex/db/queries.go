package db

import (
	"database/sql"
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

// GetSceneByID looks up a single scene by its product ID
func GetSceneByID(tx *sql.Tx, productID string) (*LandsatLocalIndexScene, error) {
	var wrsBoundsBytes []byte
	var mtlBoundsBytes []byte
	scene := LandsatLocalIndexScene{}

	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds), 
		       ST_AsGeoJSON(ST_MakePolygon(ST_MakeLine(ARRAY[corner_ul, corner_ur, corner_lr, corner_ll, corner_ul])))
		FROM public.scenes
		WHERE product_id=$1
			AND corner_ll IS NOT NULL 
		LIMIT 1`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &wrsBoundsBytes, &mtlBoundsBytes)
	if err != nil {
		return nil, err
	}

	mtlBounds, err := geojson.PolygonFromBytes(mtlBoundsBytes)
	if err != nil {
		return nil, err
	}
	scene.Bounds = mtlBounds
	scene.BoundingBox = mtlBounds.ForceBbox()

	return &scene, nil
}

// SearchScenes does a lookup in indexed scenes based on a bounding box, cloud cover, and time window
// Note: Any cloud cover that is <0 is usually corrupt in some way, and shall be excluded
func SearchScenes(tx *sql.Tx, bbox geojson.BoundingBox, maxCloudCover float64, minAcquiredDate time.Time, maxAcquiredDate time.Time) ([]LandsatLocalIndexScene, error) {
	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds), 
		       ST_AsGeoJSON(bounds)
		FROM public.scenes
		WHERE cloud_cover >= 0
			AND cloud_cover < $1
			AND acquisition_date > $2
			AND acquisition_date < $3
			AND corner_ll IS NOT NULL 
			AND ST_Intersects(bounds, ST_MakeEnvelope($4, $5, $6, $7, 4326))
		ORDER BY acquisition_date DESC
		LIMIT 100`,
		maxCloudCover*100, // Cloud cover is imported as 0-100, not as 0-1
		minAcquiredDate, maxAcquiredDate,
		bbox[0], bbox[1], bbox[2], bbox[3],
	)
	if err != nil {
		return nil, err
	}

	results := []LandsatLocalIndexScene{}
	for rows.Next() {
		var wrsBoundsBytes []byte
		var mtlBoundsBytes []byte
		scene := LandsatLocalIndexScene{}
		if err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &wrsBoundsBytes, &mtlBoundsBytes); err != nil {
			return nil, err
		}

		mtlBounds, err := geojson.PolygonFromBytes(mtlBoundsBytes)
		if err != nil {
			return nil, err
		}
		scene.Bounds = mtlBounds
		scene.BoundingBox = mtlBounds.ForceBbox()

		results = append(results, scene)
	}

	return results, nil
}

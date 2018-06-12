package db

import (
	"database/sql"
	"time"

	"github.com/venicegeo/geojson-go/geojson"
)

func GetSceneByID(tx *sql.Tx, productID string) (*LandsatLocalIndexScene, error) {
	var boundsBytes []byte
	scene := LandsatLocalIndexScene{}

	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds)
		FROM public.scenes
		WHERE product_id=$1
		LIMIT 1`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &boundsBytes)
	if err != nil {
		return nil, err
	}

	scene.Bounds, err = geojson.PolygonFromBytes(boundsBytes)
	if err != nil {
		return nil, err
	}

	return &scene, nil
}

func SearchScenes(tx *sql.Tx, bbox geojson.BoundingBox, maxCloudCover float64, minAcquiredDate time.Time, maxAcquiredDate time.Time) ([]LandsatLocalIndexScene, error) {
	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, ST_AsGeoJSON(bounds)
		FROM public.scenes
		WHERE cloud_cover < $1
		      AND acquisition_date > $2
					AND acquisition_date < $3
					AND ST_Intersects(bounds, ST_MakeEnvelope($4, $5, $6, $7, 4326))
		ORDER BY acquisition_date DESC
		LIMIT 100`,
		maxCloudCover, minAcquiredDate, maxAcquiredDate,
		bbox[0], bbox[1], bbox[2], bbox[3],
	)
	if err != nil {
		return nil, err
	}

	results := []LandsatLocalIndexScene{}
	for rows.Next() {
		var boundsBytes []byte
		scene := LandsatLocalIndexScene{}
		if err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &boundsBytes); err != nil {
			return nil, err
		}

		if scene.Bounds, err = geojson.PolygonFromBytes(boundsBytes); err != nil {
			return nil, err
		}
		results = append(results, scene)
	}

	return results, nil
}

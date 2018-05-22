package db

import (
	"database/sql"
)

func GetSceneByID(tx *sql.Tx, productID string) (*LandsatLocalIndexScene, error) {
	scene := LandsatLocalIndexScene{}

	rows, err := tx.Query(`
		SELECT product_id, acquisition_date, cloud_cover, scene_url, bounds
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
	err = rows.Scan(&scene.ProductID, &scene.AcquisitionDate, &scene.CloudCover, &scene.SceneURLString, &scene.Bounds)
	if err != nil {
		return nil, err
	}

	return &scene, nil
}

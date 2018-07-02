package main

import (
	"database/sql"
	"log"

	"github.com/venicegeo/bf-ia-broker/landsat_localindex/db/metadata"
	"github.com/venicegeo/bf-ia-broker/util"
	cli "gopkg.in/urfave/cli.v1"
)

const missingDataQuery = `
SELECT product_id, scene_url FROM scenes WHERE 
corner_ul IS NULL OR 
corner_ur IS NULL OR  
corner_ll IS NULL OR 
corner_lr IS NULL 
`

const insertSQL = `
UPDATE scenes SET 
	corner_ul = st_setSRID( st_MakePoint($2, $3), 4326), 
	corner_ur = st_setSRID( st_MakePoint($4, $5), 4326),
	corner_lr = st_setSRID(st_MakePoint($6, $7), 4326),
	corner_ll = st_setSRID( st_MakePoint($8, $9), 4326)
WHERE product_id = $1
`

type sceneRow struct {
	productID string
	url       string
}

type sceneMetadata struct {
	productID string
	metadata  *metadata.LandsatSceneMetadata
}

func landsatPopulateMetadata(*cli.Context) {
	conn, err := getDbConnection(&util.BasicLogContext{})

	if err != nil {
		log.Fatal("Error opening db connection.")
	}
	defer conn.Close()

	//Create the statement to inser the info into the database.
	insertStmt, err := conn.Prepare(insertSQL)
	if err != nil {
		log.Fatal("Error preparing insert statement: " + err.Error())
	}
	defer insertStmt.Close()

	rows, err := conn.Query(missingDataQuery)
	if err != nil {
		log.Fatal("Error with query.")
	}
	defer rows.Close()

	numWorkers := 10
	scenesQueue := make(chan *sceneRow, numWorkers)
	responseQueue := make(chan *sceneMetadata, numWorkers)
	workerCompleteChan := make(chan bool, 1)

	//Start some workers.
	for i := 0; i < numWorkers; i++ {
		go downloadWorker(scenesQueue, responseQueue, workerCompleteChan)
	}

	//Listen for their exit.
	go func() {
		workersDone := 0
		for workersDone < numWorkers {
			<-workerCompleteChan
			workersDone++
		}
		close(responseQueue)
	}()

	//Launch a process to write all of the rows into the channel where
	//the workers will listen for them.
	go func() {
		for rows.Next() {
			var theRow sceneRow
			scanErr := rows.Scan(&theRow.productID, &theRow.url)
			if scanErr != nil {
				log.Printf("Failure reading row.")
				continue
			}

			scenesQueue <- &theRow
		}
		close(scenesQueue)
		log.Printf("Sql rows done.")
	}()

	//Read the responses and write them into the database.
	for meta := range responseQueue {
		insertMetadata(insertStmt, meta)
		// Commented to squelch log spam
		//log.Printf("Done " + meta.productID)
	}

	log.Printf("Done")
}

func insertMetadata(stmt *sql.Stmt, scene *sceneMetadata) {
	_, err := stmt.Exec(scene.productID,
		scene.metadata.Bounds.Coordinates[0][0][0], scene.metadata.Bounds.Coordinates[0][0][1],
		scene.metadata.Bounds.Coordinates[0][1][0], scene.metadata.Bounds.Coordinates[0][1][1],
		scene.metadata.Bounds.Coordinates[0][2][0], scene.metadata.Bounds.Coordinates[0][2][1],
		scene.metadata.Bounds.Coordinates[0][3][0], scene.metadata.Bounds.Coordinates[0][3][1],
	)
	if err != nil {
		log.Printf("Error inserting values.")
	}
}

func downloadWorker(scenesChan chan *sceneRow, responseChan chan *sceneMetadata, completeChan chan bool) {
	for scene := range scenesChan {
		result := sceneMetadata{
			productID: scene.productID,
		}

		// Commented to squelch log spam
		//log.Printf("Getting scene metadata %s", scene.productID)
		var err error
		result.metadata, err = metadata.GetLandsatS3SceneMetadata(scene.productID, scene.url)

		if err != nil {
			log.Printf("Error getting metadata.")
		}

		responseChan <- &result
	}
	completeChan <- true
	log.Printf("Worker exited.")
}

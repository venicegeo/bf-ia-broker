package main

import (
	"compress/gzip"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/venicegeo/bf-ia-broker/landsataws"

	_ "github.com/lib/pq"
	cli "gopkg.in/urfave/cli.v1"
)

const connectionStringEnv = "postgresConnectionString"
const scenesFileEnv = "scenesCsvUrl"

func landsatIngestAction(*cli.Context) {
	scenesURL := os.Getenv(scenesFileEnv)
	scenesIsGzip := strings.HasSuffix(strings.ToLower(scenesURL), "gz")

	var mainReader io.Reader
	sourceReader, err := openReader(scenesURL)
	if err != nil {
		log.Fatal("Could not open the source file/url.")
	}
	defer sourceReader.Close()
	mainReader = sourceReader

	if scenesIsGzip {
		archiveReader, zipErr := gzip.NewReader(mainReader)
		if zipErr != nil {
			log.Fatal("Error opening gzip archive.", zipErr)
		}
		defer archiveReader.Close()
		mainReader = archiveReader
	}

	database, err := getDbConnection()
	if err != nil {
		log.Fatal("Could not open database connection.")
	}
	defer database.Close()

	landsataws.Ingest(mainReader, database)
}

func openReader(scenesURL string) (io.ReadCloser, error) {
	if strings.HasPrefix(scenesURL, "http://") || strings.HasPrefix(scenesURL, "https://") {
		log.Println("Requesting url:", scenesURL)
		archiveResponse, netErr := http.Get("https://landsat-pds.s3.amazonaws.com/c1/L8/scene_list.gz")
		if netErr != nil {
			return nil, netErr
		}
		return archiveResponse.Body, nil
	}

	//Treat this as a file.
	cleanPath := filepath.Clean(scenesURL)
	log.Println("Opening file", cleanPath)
	file, err := os.Open(cleanPath)
	return file, err
}

func getDbConnection() (*sql.DB, error) {
	connStr := os.Getenv(connectionStringEnv)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

package main

import (
	"log"

	"github.com/pressly/goose"
	cli "gopkg.in/urfave/cli.v1"

	_ "github.com/venicegeo/bf-ia-broker/migrations"
	"github.com/venicegeo/bf-ia-broker/util"
)

func migrateDatabaseAction(*cli.Context) {
	database, err := getDbConnectionFunc(&util.BasicLogContext{})
	if err != nil {
		log.Fatal("Could not open database connection.")
	}
	defer database.Close()

	goose.Run("up", database, ".")
}

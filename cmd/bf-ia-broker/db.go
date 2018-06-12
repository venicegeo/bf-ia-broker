package main

import (
	"database/sql"
	"errors"
	"os"

	"github.com/venicegeo/bf-ia-broker/util"
)

const connectionStringEnv = "DATABASE_URL"
const vcapServicesEnv = "VCAP_SERVICES"
const pzPostgresService = "pz-postgres"

//getDbConnection opens a new database connection.
func getDbConnection(ctx util.LogContext) (*sql.DB, error) {
	connStr := os.Getenv(connectionStringEnv)
	if connStr == "" {
		util.LogInfo(ctx, "No DB connection found in DATABASE_URL, checking VCAP_SERVICES")
		services, err := util.ParseVcapServices([]byte(os.Getenv(vcapServicesEnv)))
		if err != nil {
			return nil, errors.New("Could not get DB connection from DATABASE_URL or VCAP_SERVICES (no valid VCAP_SERVICES found): " + err.Error())
		}
		service := services.FindServiceByName(pzPostgresService)
		if service == nil {
			return nil, errors.New("Could not get DB connection from DATABASE_URL or VCAP_SERVICES ('pz-postgres' service not found)")
		}
		connStr, err = service.Credentials.String("uri")
		if err != nil {
			return nil, errors.New("Could not get DB connection from DATABASE_URL or VCAP_SERVICES (error getting URI string): " + err.Error())
		}
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

var getDbConnectionFunc = getDbConnection

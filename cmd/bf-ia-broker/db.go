package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
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
			return nil, fmt.Errorf("Could not get DB connection from DATABASE_URL or VCAP_SERVICES ('pz-postgres' service not found); available services: %v",
				services.GetServiceNames())
		}
		connStr, err = service.Credentials.String("uri")
		if err != nil {
			return nil, errors.New("Could not get DB connection from DATABASE_URL or VCAP_SERVICES (error getting URI string): " + err.Error())
		}
	}

	// XXX: pq expects SSL to be enabled if not explicitly disabled; we need to explicitly disable it
	dbURI, _ := url.Parse(connStr)
	params := dbURI.Query()
	params.Set("sslmode", "disable")
	dbURI.RawQuery = params.Encode()

	util.LogInfo(ctx, fmt.Sprintf("Creating database connection at: `%s`", dbURI.String()))
	db, err := sql.Open("postgres", dbURI.String())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

var getDbConnectionFunc = getDbConnection

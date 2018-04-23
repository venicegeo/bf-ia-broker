// Copyright 2018, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/landsat"
	"github.com/venicegeo/bf-ia-broker/planet"
	"github.com/venicegeo/bf-ia-broker/util"
	cli "gopkg.in/urfave/cli.v1"
)

func getPortStr() string {
	if port, ok := os.LookupEnv("PORT"); ok {
		return ":" + port
	}
	return ":8080"
}

func createRouter(ctx util.LogContext) (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		util.LogAudit(ctx, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving / request", Severity: util.INFO})
		writer.Write([]byte("Hi"))
		util.LogAudit(ctx, util.LogAuditInput{Actor: request.URL.String(), Action: request.Method + " response", Actee: "anon user", Message: "Sending / response", Severity: util.INFO})
	})
	router.Handle("/planet/discover/{itemType}", planet.NewDiscoverHandler())
	router.Handle("/planet/{itemType}/{id}", planet.NewMetadataHandler())
	router.Handle("/planet/activate/{itemType}/{id}", planet.NewActivateHandler())
	return
}

func serveAction(*cli.Context) {
	logContext := &(util.BasicLogContext{})

	portStr := getPortStr()

	if len(util.GetLandsatHost()) != 0 {
		util.LogInfo(logContext, fmt.Sprintf("Starting Landsat scene list query loop for host: '%s'", util.GetLandsatHost()))
		go landsat.UpdateSceneMapOnTicker(30*time.Minute, logContext)
	} else {
		util.LogAlert(logContext, "No Landsat host found, not starting Landsat scene list query loop")
	}
	launchServerFunc(portStr, createRouter(logContext))
}

var launchServerFunc = launchServer

func launchServer(portStr string, router *mux.Router) {
	server := http.Server{
		Addr:    portStr,
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
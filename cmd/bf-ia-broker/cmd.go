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
	cli "gopkg.in/urfave/cli.v1"
)

var commands = cli.Commands{
	cli.Command{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "Launch the bf-ia-broker webserver",
		Action:  serveAction,
	},
	cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print the version number of the Broker CLI",
		Action:  versionAction,
	},
	cli.Command{
		Name:    "landsat_ingest",
		Aliases: []string{"l"},
		Usage:   "Update the database with the latest landsat entries",
		Action:  landsatIngestAction,
	},
	cli.Command{
		Name:    "migrate",
		Aliases: []string{"m"},
		Usage:   "Update database schema",
		Action:  migrateDatabaseAction,
	},
}

func createCliApp() (app *cli.App) {
	app = cli.NewApp()
	app.Name = "bf-ia-broker"
	app.Usage = "Launch a bf-ia-broker process"
	app.Commands = commands
	return
}

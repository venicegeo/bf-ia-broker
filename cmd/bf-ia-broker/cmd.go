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
}

func createCliApp() (app *cli.App) {
	app = cli.NewApp()
	app.Name = "bf-ia-broker"
	app.Usage = "Launch a bf-ia-broker process"
	app.Commands = commands
	return
}

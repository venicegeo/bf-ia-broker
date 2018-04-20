package main

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
)

func versionAction(*cli.Context) {
	fmt.Println("bf-ia-broker v2.0")
}

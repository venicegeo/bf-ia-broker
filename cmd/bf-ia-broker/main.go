package main

import (
	"fmt"
	"os"

	"github.com/venicegeo/bf-ia-broker/util"
)

func main() {
	util.LogAudit(&(util.BasicLogContext{}), util.LogAuditInput{Actor: "main()", Action: "startup", Actee: "self", Message: "Application Startup", Severity: util.INFO})
	err := createCliApp().Run(os.Args)
	if err != nil {
		util.LogAlert(&(util.BasicLogContext{}), fmt.Sprintf("Error executing CLI app: %v", err))
	}
}

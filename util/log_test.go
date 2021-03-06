// Copyright 2016, RadiantBlue Technologies, Inc.
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

package util

import (
	"errors"
	"testing"
)

func TestLog(t *testing.T) {
	context := &BasicLogContext{}
	message := "Here is an info."
	LogAlert(context, "Here is an alert.")
	LogSimpleErr(context, "Here is an error.", errors.New("Error happened."))
	LogInfo(context, message)
	LogAudit(context, LogAuditInput{Actor: "log_test", Action: "Test Audit", Actee: "LogAudit", Message: message, Severity: DEBUG})
}

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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

var httpClient *http.Client

// HTTPErr represents any HTTP error
type HTTPErr struct {
	Status  int
	Message string
}

func (err HTTPErr) Error() string {
	return fmt.Sprintf("%d: %v", err.Status, err.Message)
}

// HTTPClient is a factory method for a http.Client suitable for common operations
func HTTPClient() *http.Client {
	if httpClient == nil {
		transport := &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		}

		httpClient = &http.Client{Transport: transport}
	}
	return httpClient
}

// SetHTTPClient is used to set the current http client.  This is mostly useful
// for testing purposes
func SetHTTPClient(newClient *http.Client) {
	httpClient = newClient
}

// RequestKnownJSON submits an http request where the response is assumed to be JSON
// for which the format is known.  Given an object of the appropriate format for
// said response JSON, an address to call and an authKey to send, it will submit
// the get request, unmarshal the result into the given object, and return. It
// returns the response buffer, in case it is needed for debugging purposes.
func RequestKnownJSON(method, bodyStr, address, authKey string, outpObj interface{}) ([]byte, error) {

	resp, err := SubmitSinglePart(method, bodyStr, address, authKey)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		if resp != nil {
			errByt, _ := ioutil.ReadAll(resp.Body)
			return errByt, err
		}
		return nil, err
	}
	return ReadBodyJSON(outpObj, resp.Body)
}

// ReqByObjJSON is much like RequestKnownJSON, except that it takes an interface (which
// it then json-marshals) as its input, rather than an already-marshaled string
func ReqByObjJSON(method, addr, authKey string, inpObj, outpObj interface{}) ([]byte, error) {
	byts, err := json.Marshal(inpObj)
	if err != nil {
		return nil, &Error{LogMsg: "Json Marshal attempt failed."}

	}
	byts, pErr := RequestKnownJSON(method, string(byts), addr, "", outpObj)
	if pErr != nil {
		return nil, pErr
	}
	return byts, nil
}

// SubmitMultipart sends a multi-part POST call, including an optional uploaded file,
// and returns the response.  Primarily intended to support Ingest calls.
func SubmitMultipart(bodyStr, address, filename, authKey string, fileData []byte) (*http.Response, error) {

	var (
		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
		client = HTTPClient()
		err    error
	)

	err = writer.WriteField("data", bodyStr)
	if err != nil {
		return nil, &Error{LogMsg: "Could not write string " + bodyStr + "to message body: " + err.Error(), SimpleMsg: "Internal Error on file upload.  See logs."}
	}

	if fileData != nil {
		var part io.Writer
		part, err = writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, &Error{LogMsg: "Error on CreateFormFile: " + err.Error(), SimpleMsg: "Internal Error on file upload.  See logs."}
		}
		if part == nil {
			return nil, &Error{LogMsg: "CreateFormFile returned empty form.", SimpleMsg: "Internal Error on file upload.  See logs."}
		}

		_, err = io.Copy(part, bytes.NewReader(fileData))
		if err != nil {
			return nil, &Error{LogMsg: "Error on file data Copy: " + err.Error(), SimpleMsg: "Internal Error on file upload.  See logs."}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, &Error{LogMsg: "Error on Writer close: " + err.Error(), SimpleMsg: "Internal Error on file upload.  See logs."}
	}

	fileReq, err := http.NewRequest("POST", address, body)
	if err != nil {
		return nil, &Error{LogMsg: "Error on Request creation: " + err.Error(), SimpleMsg: "Internal Error on file upload.  See logs."}
	}

	fileReq.Header.Add("Content-Type", writer.FormDataContentType())
	fileReq.Header.Add("Authorization", authKey)

	resp, err := client.Do(fileReq)
	if err != nil {
		return nil, &Error{LogMsg: "Error on POST multipart: " + err.Error(), URL: address, Request: bodyStr, SimpleMsg: "HTTP error on file upload.  See logs."}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		errByt, _ := ioutil.ReadAll(resp.Body)
		outMsg := "Received " + http.StatusText(resp.StatusCode) + " on multipart POST call to " + address + ".  Further details logged."
		return resp, &Error{LogMsg: "Failed multipart HTTP request", URL: address, Request: bodyStr, Response: string(errByt), HTTPStatus: resp.StatusCode, SimpleMsg: outMsg}
	}
	return resp, nil
}

// SubmitSinglePart sends a single-part GET/POST/PUT/DELETE call to the target URL
// and returns the result.  Includes the necessary headers.
func SubmitSinglePart(method, bodyStr, url, authKey string) (*http.Response, error) {

	var (
		fileReq *http.Request
		err     error
		client  = HTTPClient()
	)

	if method == "" || url == "" {
		return nil, &Error{LogMsg: `method:"` + method + `", url:"` + url + `".  You must have both.`}
	}

	if bodyStr != "" {
		fileReq, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(bodyStr)))
		if err != nil {
			return nil, &Error{LogMsg: err.Error()}
		}
		fileReq.Header.Add("Content-Type", "application/json")
	} else {
		fileReq, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, &Error{LogMsg: err.Error()}
		}
	}

	fileReq.Header.Add("Authorization", authKey)

	resp, err := client.Do(fileReq)
	if err != nil {
		return nil, &Error{LogMsg: err.Error(), Request: bodyStr}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		errByt, _ := ioutil.ReadAll(resp.Body)

		outMsg := "Received " + http.StatusText(resp.StatusCode) + " on call to " + url + ".  Further details logged."
		return resp, &Error{LogMsg: "Failed HTTP request", Request: bodyStr, Response: string(errByt), URL: url, HTTPStatus: resp.StatusCode, SimpleMsg: outMsg}
	}

	return resp, nil
}

// ReadBodyJSON takes the body of either a request object or a response
// object, pulls out the body, and attempts to interpret it as JSON into
// the given interface format.  It's mostly there as a minor simplifying
// function.
func ReadBodyJSON(output interface{}, body io.ReadCloser) ([]byte, error) {
	rBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, &Error{LogMsg: "Could not read HTTP response."}
	}
	err = json.Unmarshal(rBytes, output)
	if err != nil {
		return nil, &Error{LogMsg: "Unmarshal failed: " + err.Error() + ".  Original input: " + string(rBytes) + ".", SimpleMsg: "JSON error when reading HTTP Response.  See log."}
	}
	return rBytes, nil
}

// CheckAuth verifies that the given API key is valid for the given
// Piazza address
func CheckAuth(pzAddr, pzAuth string) *Error {
	_, err := SubmitSinglePart("GET", "", pzAddr+"/service", pzAuth)
	if err != nil {
		return &Error{LogMsg: "Could not confirm user authorization."}
	}
	return nil
}

// HTTPOut outputs the given string on the given responseWriter
// with the given http code.  It is nearly identical in behavior
// to http.Error, except that it doesn't modify the headers
// otherwise, allowing us to maintain the Content-Type of
// application/json, and make things a bit easier for our
// service consumers to digest.
func HTTPOut(w http.ResponseWriter, output string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(output))
}

// Preflight sets up the CORS stuff and
// returns TRUE if this is an OPTIONS request
func Preflight(w http.ResponseWriter, r *http.Request, lc LogContext) bool {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	if r.Method == "OPTIONS" {
		LogAudit(lc, LogAuditInput{Actor: r.URL.String(), Action: r.Method, Actee: "anon user", Message: "Returning Preflight", Severity: INFO})
		return true
	}
	return false
}

// PrintJSON marshals the given object, turns it into a string, and feeds it to
// the given ResponseWriter.
func PrintJSON(w http.ResponseWriter, output interface{}, httpStatus int) []byte {
	outBuf, err := json.Marshal(output)
	if err != nil {
		HTTPOut(w, `{"Errors":"JSON marshal failure on output: `+err.Error()+`"}`, http.StatusInternalServerError)
	} else {
		HTTPOut(w, string(outBuf), httpStatus)
	}
	return outBuf
}

// HTTPError provides an error message that conceals the detailed information for security reasons
func HTTPError(r *http.Request, w http.ResponseWriter, context LogContext, message string, status int) {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	LogAudit(context, LogAuditInput{Actor: r.URL.String(), Action: r.Method + " response", Actee: "anon user", Message: "Returning error response", Severity: INFO})
	if message == "" {
		message = fmt.Sprintf("An error occurred. Please contact your system administrator. Session ID: %v", context.SessionID())
	}
	http.Error(w, message, status)
}

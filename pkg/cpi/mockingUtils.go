// +build !release

package cpi

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

//GetCPIFunctionMockResponse -Generate mock response payload for different CPI functions
func GetCPIFunctionMockResponse(functionName, testType string) (*http.Response, error) {
	switch functionName {
	case "DeployIntegrationDesigntimeArtifact":
		if testType == "Positive" {
			return GetEmptyHttpResponseBody()
		}
		res := http.Response{
			StatusCode: 500,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
						"code": "Internal Server Error",
						"message": {
						"@lang": "en",
						"#text": "Cannot deploy artifact with Id 'flow1'!"
						}
					}`))),
		}
		return &res, errors.New("Internal Server Error")

	case "UpdateIntegrationArtifactConfiguration":
		if testType == "Positive" {
			return GetEmptyHttpResponseBody()
		}
		res := http.Response{
			StatusCode: 404,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
						"code": "Not Found",
						"message": {
						"@lang": "en",
						"#text": "Parameter key 'Parameter1' not found."
						}
					}`))),
		}
		return &res, errors.New("Not found - either wrong version for the given Id or wrong parameter key")

	default:
		res := http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}
		return &res, errors.New("Service not Found")
	}
}

//GetEmptyHttpResponseBody -Empty http respose body
func GetEmptyHttpResponseBody() (*http.Response, error) {
	res := http.Response{
		StatusCode: 202,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
	}
	return &res, nil
}

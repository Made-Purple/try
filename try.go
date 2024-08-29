package try

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

type TestCase struct {
	TestName           string
	Request            Request
	RequestBody        interface{}
	RequestReader      io.Reader
	RequestContentType string
	RequestCookies     []*http.Cookie
	RequestHeaders     map[string]string
	Expected           ExpectedResponse
	AccessToken        string
	Setup              func(testCase *TestCase)
	Teardown           func(testCase *TestCase, res *HijackableResponseRecorder)
	DisplayResponse    bool
}

type Request struct {
	Method string
	Url    string
}

type ExpectedResponse struct {
	StatusCode       int
	BodyPart         string
	BodyParts        []string
	BodyPartMissing  string
	BodyPartsMissing []string
}

func GenerateRequest(testCase *TestCase) (*http.Request, error) {

	reqJson, err := json.Marshal(testCase.RequestBody)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	if testCase.RequestReader != nil {
		req, err = http.NewRequest(testCase.Request.Method, testCase.Request.Url, testCase.RequestReader)
	} else {
		req, err = http.NewRequest(testCase.Request.Method, testCase.Request.Url, bytes.NewBuffer(reqJson))
	}

	if err != nil {
		return nil, err
	}

	// Should always request JSON
	if testCase.RequestContentType != "" {
		req.Header.Set(echo.HeaderContentType, testCase.RequestContentType)
	} else {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	req.Header.Set(echo.HeaderXRealIP, "127.0.0.0")

	// Add cookies in if present
	if len(testCase.RequestCookies) > 0 {
		for _, cookie := range testCase.RequestCookies {
			req.AddCookie(cookie)
		}
	}

	// Add headers if present
	if len(testCase.RequestHeaders) > 0 {
		for headerKey, headerValue := range testCase.RequestHeaders {
			// Set requires to overrise content type. May need to be add if you need multiple headers
			// with the same key.
			req.Header.Set(headerKey, headerValue)
		}
	}

	return req, nil
}

func ExecuteRequest(e *echo.Echo, req *http.Request) *HijackableResponseRecorder {
	// Create a new recorder then process request with server.
	rr := NewHijackableRecorder(nil)
	e.ServeHTTP(rr, req)
	return rr
}

func ValidateResults(t *testing.T, test *TestCase, res *HijackableResponseRecorder) {

	if test.DisplayResponse {
		fmt.Println("Request Output: ")
		fmt.Println(res.Body.String())
	}

	if res.Code != 0 {
		assert.Equal(t, test.Expected.StatusCode, res.Code)
	}

	if test.Expected.BodyPart != "" {
		assert.Contains(t, res.Body.String(), test.Expected.BodyPart)
	}
	if len(test.Expected.BodyParts) > 0 {
		for _, expectedText := range test.Expected.BodyParts {
			assert.Contains(t, res.Body.String(), expectedText)
		}
	}
	if test.Expected.BodyPartMissing != "" {
		assert.NotContains(t, res.Body.String(), test.Expected.BodyPartMissing)
	}

	if len(test.Expected.BodyPartsMissing) > 0 {
		for _, expectedText := range test.Expected.BodyPartsMissing {
			assert.NotContains(t, res.Body.String(), expectedText)
		}
	}
}

func ExecuteTest(t *testing.T, e *echo.Echo, testCase *TestCase) {

	// Run any setup required before we execute the request
	if testCase.Setup != nil {
		testCase.Setup(testCase)
	}
	req, err := GenerateRequest(testCase)
	if err != nil {
		t.Fatalf("unable to Generate Request: %v", err)
	}
	res := ExecuteRequest(e, req)
	ValidateResults(t, testCase, res)

	if testCase.Teardown != nil {
		testCase.Teardown(testCase, res)
	}
}

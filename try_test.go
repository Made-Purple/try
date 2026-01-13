package try

import (
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type user struct {
	Name string
}

func TestExecuteRequest(t *testing.T) {

	tc := &TestCase{
		TestName: "Can create User",
		Request: Request{
			Method: http.MethodPost,
			Url:    "/users",
		},
	}

	req, err := GenerateRequest(tc)
	if err != nil {
		t.Fatalf("Error generating request: %v", err)
	}

	e := echo.New()
	res := ExecuteRequest(e, req)

	assert.Equal(t, `{"message":"Not Found"}`+"\n", res.Body.String())
}

func TestExecuteRequestAdditional(t *testing.T) {
	r := strings.NewReader("my request")
	tc := &TestCase{
		TestName:           "Can create User",
		RequestReader:      r,
		RequestContentType: "something",
		Request: Request{
			Method: http.MethodPost,
			Url:    "/users",
		},
	}

	req, err := GenerateRequest(tc)
	if err != nil {
		t.Fatalf("Error generating request: %v", err)
	}

	e := echo.New()
	res := ExecuteRequest(e, req)

	res.Closed()
	res.Hijack()

	assert.Equal(t, `{"message":"Not Found"}`+"\n", res.Body.String())
}

func TestValidateResults(t *testing.T) {
	expectedCallbackCalled := false
	tc := &TestCase{
		TestName: "Can create User",
		Request: Request{
			Method: http.MethodPost,
			Url:    "/users",
		},
		Expected: ExpectedResponse{
			StatusCode:       404,
			BodyPart:         "Not Found",
			BodyParts:        []string{"Not Found"},
			BodyPartMissing:  "This is Not Returned",
			BodyPartsMissing: []string{"This is Not Returned"},
			ExpectedCallBack: func(res *HijackableResponseRecorder) {
				expectedCallbackCalled = true
			},
		},
	}

	req, err := GenerateRequest(tc)
	if err != nil {
		t.Fatalf("Error generating request: %v", err)
	}

	e := echo.New()
	res := ExecuteRequest(e, req)

	ValidateResults(t, tc, res)

	assert.True(t, expectedCallbackCalled)
}

func TestExecuteTest(t *testing.T) {
	adminRefreshCookie := &http.Cookie{
		Name: "test cookie",
	}

	u := &user{Name: "Matt Nelson"}

	tc := &TestCase{
		TestName: "Can create User",
		Request: Request{
			Method: http.MethodPost,
			Url:    "/users",
		},
		Setup:           func(testCase *TestCase) {},
		Teardown:        func(testCase *TestCase, res *HijackableResponseRecorder) {},
		RequestBody:     u,
		RequestCookies:  []*http.Cookie{adminRefreshCookie},
		RequestHeaders:  map[string]string{"test-header": "header"},
		DisplayResponse: true,
		Expected: ExpectedResponse{
			StatusCode:       404,
			BodyPart:         "Not Found",
			BodyParts:        []string{"Not Found"},
			BodyPartMissing:  "This is Not Returned",
			BodyPartsMissing: []string{"This is Not Returned"},
			Headers:          map[string]string{"Content-Type": "application/json"},
		},
	}
	e := echo.New()
	ExecuteTest(t, e, tc)

}
func TestGenerateRequest_BadData(t *testing.T) {

	tc := &TestCase{
		TestName: "Can create User",
		Request: Request{
			Method: http.MethodPost,
			Url:    "/users",
		},
		RequestBody: func() {},
	}

	_, err := GenerateRequest(tc)
	if err == nil {
		t.Fatal("Expecting error for bad data.")
	}

}

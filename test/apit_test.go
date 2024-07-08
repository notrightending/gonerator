package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/notrightending/gonerator/example"
)

var (
	client = &http.Client{Timeout: time.Second}
)

type Case struct {
	Method string
	Path   string
	Query  string
	Auth   bool
	Status int
	Result interface{}
}

const (
	ApiUserCreate  = "/user/create"
	ApiUserProfile = "/user/profile"
)

type CR map[string]interface{}

func TestMain(m *testing.M) {
	fmt.Printf("MY_API_KEY: %s\n", os.Getenv("MY_API_KEY"))
	fmt.Printf("OTHER_API_KEY: %s\n", os.Getenv("OTHER_API_KEY"))

	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		os.Exit(1)
	}

	// Build the generator
	buildCmd := exec.Command("go", "build", "-o", "generator", "./cmd/generator")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err = buildCmd.Run()
	if err != nil {
		fmt.Printf("Error building generator: %v\n", err)
		os.Exit(1)
	}

	// Run the generator
	genCmd := exec.Command("./generator", "example/api.go", "example/generated_api.go")
	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr
	err = genCmd.Run()
	if err != nil {
		fmt.Printf("Error generating API handlers: %v\n", err)
		os.Exit(1)
	}

	// Set up environment variables for testing
	os.Setenv("MY_API_KEY", "test_my_api_key")
	os.Setenv("OTHER_API_KEY", "test_other_api_key")

	// Run tests
	code := m.Run()

	// Clean up
	os.Unsetenv("MY_API_KEY")
	os.Unsetenv("OTHER_API_KEY")

	os.Exit(code)
}

func TestMyApi(t *testing.T) {
	ts := httptest.NewServer(example.NewMyApi())
	defer ts.Close()

	cases := []Case{
		{
			Path:   ApiUserProfile,
			Query:  "login=rvasily",
			Status: http.StatusOK,
			Result: CR{
				"error": "",
				"response": CR{
					"id":        42,
					"login":     "rvasily",
					"full_name": "Vasily Romanov",
					"status":    20,
				},
			},
		},
		{
			Path:   ApiUserProfile,
			Method: http.MethodPost,
			Query:  "login=rvasily",
			Status: http.StatusOK,
			Result: CR{
				"error": "",
				"response": CR{
					"id":        42,
					"login":     "rvasily",
					"full_name": "Vasily Romanov",
					"status":    20,
				},
			},
		},
		{
			Path:   ApiUserProfile,
			Query:  "",
			Status: http.StatusBadRequest,
			Result: CR{
				"error": "login must be not empty",
			},
		},
		{
			Path:   ApiUserProfile,
			Query:  "login=bad_user",
			Status: http.StatusInternalServerError,
			Result: CR{
				"error": "bad user",
			},
		},
		{
			Path:   ApiUserProfile,
			Query:  "login=not_exist_user",
			Status: http.StatusNotFound,
			Result: CR{
				"error": "user not exist",
			},
		},
		{
			Path:   "/user/unknown",
			Query:  "login=not_exist_user",
			Status: http.StatusNotFound,
			Result: CR{
				"error": "unknown method",
			},
		},
		{
			Path:   ApiUserCreate,
			Method: http.MethodPost,
			Query:  "login=mr.moderator&age=32&status=moderator&full_name=Ivan_Ivanov",
			Status: http.StatusOK,
			Auth:   true,
			Result: CR{
				"error": "",
				"response": CR{
					"id": 43,
				},
			},
		},
		{
			Path:   ApiUserProfile,
			Query:  "login=mr.moderator",
			Status: http.StatusOK,
			Result: CR{
				"error": "",
				"response": CR{
					"id":        43,
					"login":     "mr.moderator",
					"full_name": "Ivan_Ivanov",
					"status":    10,
				},
			},
		},
		{
			Path:   ApiUserCreate,
			Method: http.MethodGet,
			Query:  "login=mr.moderator&age=32&status=moderator&full_name=GetMethod",
			Status: http.StatusNotAcceptable,
			Auth:   true,
			Result: CR{
				"error": "bad method",
			},
		},
		{
			Path:   ApiUserCreate,
			Method: http.MethodPost,
			Query:  "any_params=123",
			Status: http.StatusForbidden,
			Auth:   false,
			Result: CR{
				"error": "unauthorized",
			},
		},
		// Add more test cases as needed
	}

	runTests(t, ts, cases)
}

func TestOtherApi(t *testing.T) {
	ts := httptest.NewServer(example.NewOtherApi())
	defer ts.Close()

	cases := []Case{
		{
			Path:   ApiUserCreate,
			Method: http.MethodPost,
			Query:  "username=I3apBap&level=1&class=barbarian&account_name=Vasily",
			Status: http.StatusBadRequest,
			Auth:   true,
			Result: CR{
				"error": "class must be one of [warrior, sorcerer, rouge]",
			},
		},
		{
			Path:   ApiUserCreate,
			Method: http.MethodPost,
			Query:  "username=I3apBap&level=1&class=warrior&account_name=Vasily",
			Status: http.StatusOK,
			Auth:   true,
			Result: CR{
				"error": "",
				"response": CR{
					"id":        12,
					"login":     "I3apBap",
					"full_name": "Vasily",
					"level":     1,
				},
			},
		},
	}

	runTests(t, ts, cases)
}

func runTests(t *testing.T, ts *httptest.Server, cases []Case) {
	for idx, item := range cases {
		var (
			err      error
			result   interface{}
			expected interface{}
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Query)

		if item.Method == http.MethodPost {
			reqBody := strings.NewReader(item.Query)
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
		}

		if item.Auth {
			var authKey string
			if strings.Contains(item.Path, "/user/create") && item.Method == http.MethodPost {
				if _, ok := ts.Config.Handler.(*example.OtherApi); ok {
					authKey = os.Getenv("OTHER_API_KEY")
				} else {
					authKey = os.Getenv("MY_API_KEY")
				}
			} else {
				authKey = os.Getenv("MY_API_KEY")
			}
			req.Header.Add("X-Auth", authKey)
			fmt.Printf("Setting X-Auth header to: %s for path: %s\n", authKey, item.Path) // Debug print
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != item.Status {
			t.Errorf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Errorf("[%s] cant unpack json: %v", caseName, err)
			continue
		}

		data, err := json.Marshal(item.Result)
		json.Unmarshal(data, &expected)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("[%d] results not match\nGot: %#v\nExpected: %#v", idx, result, item.Result)
			continue
		}
	}
}

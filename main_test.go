package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

type Request struct {
	//Method string
	Path string
	//Body string
}

type Response struct {
	//Code int
	Body   string
	Header http.Header
}

type TestCase struct {
	Request    Request
	Control    Response
	Experiment Response
	Output     string
}

func (tc TestCase) Run(t *testing.T) {

	// set up control andd experiment servers and their responses
	control := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tc.Control.Body)
	}))
	defer control.Close()
	controlURL, _ := url.Parse(control.URL)
	experiment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tc.Experiment.Body)
	}))
	defer experiment.Close()
	experimentURL, _ := url.Parse(experiment.URL)

	// set up science server proxying both control and experiment
	var output bytes.Buffer
	scienceServer := httptest.NewServer(Science{
		ControlProxy:      httputil.NewSingleHostReverseProxy(controlURL),
		ExperimentProxy:   httputil.NewSingleHostReverseProxy(experimentURL),
		ExperimentPercent: 100.0,
		DiffLog:           log.New(&output, "", 0),
	})
	defer scienceServer.Close()

	// send the test request to the science server
	resp, err := http.Get(scienceServer.URL + tc.Request.Path)
	if err != nil {
		t.Fatal(err)
	}

	// verify we got a response from the control
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatal(err)
	} else if strings.TrimSpace(string(body)) != tc.Control.Body {
		t.Fatalf("response body not proxied from control: got '%s', expected '%s'", string(body), tc.Control.Body)
	}
	for k, v := range tc.Control.Header {
		if !reflect.DeepEqual(v, resp.Header[k]) {
			t.Fatalf("response headers not proxied from control: got '%s: %s', expected '%s: %s'", k, v, k, resp.Header[k])
		}
	}

	// verify that the output of the science logic is what we expect
	templ := template.Must(template.New("output").Parse(tc.Output))
	var expectedbuf bytes.Buffer
	if err := templ.Execute(&expectedbuf, map[string]string{
		"Host": strings.Replace(scienceServer.URL, "http://", "", -1),
	}); err != nil {
		t.Fatal(err)
	}
	expected := expectedbuf.String()
	outputstr := strings.Replace(output.String(), "\r\n", "\n", -1)
	if outputstr != expected {
		t.Fatalf("expected:\n'%s'\ngot:\n'%s'\n", expected, outputstr)
	}
}

func TestScience(t *testing.T) {
	for _, tc := range []TestCase{

		// same responses
		TestCase{
			Control: Response{
				Body: "Server A",
			},
			Experiment: Response{
				Body: "Server A",
			},
			Output: ``,
		},

		// different response bodies
		TestCase{
			Control: Response{
				Body: "Server A",
			},
			Experiment: Response{
				Body: "Server B",
			},
			Output: `=== diff ===
GET / HTTP/1.1
Host: {{.Host}}
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


--- control ---
HTTP/0.0 200 OK
Content-Type: text/plain; charset=utf-8

Server A

--- experiment ---
HTTP/0.0 200 OK
Content-Type: text/plain; charset=utf-8

Server B

============
`,
		},
	} {
		tc.Run(t)
	}
}

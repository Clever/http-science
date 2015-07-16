package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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
	Body    string
	Headers map[string]string
}

type TestCase struct {
	Request    Request
	Control    Response
	Experiment Response
	Output     string
}

func (tc TestCase) Run(t *testing.T) {
	control := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-HTTP-Science") != "1" {
			t.Fatalf("Did not have X-HTTP-Science header")
		}
		fmt.Fprintln(w, tc.Control.Body)
	}))
	defer control.Close()
	experiment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-HTTP-Science") != "1" {
			t.Fatalf("Did not have X-HTTP-Science header")
		}
		fmt.Fprintln(w, tc.Experiment.Body)
	}))
	defer experiment.Close()
	var output bytes.Buffer
	scienceServer := httptest.NewServer(Science{
		ControlDial:    strings.Replace(control.URL, "http://", "", -1),
		ExperimentDial: strings.Replace(experiment.URL, "http://", "", -1),
		DiffLog:        log.New(&output, "", 0),
	})
	defer scienceServer.Close()
	_, err := http.Get(scienceServer.URL + tc.Request.Path)
	if err != nil {
		t.Fatal(err)
	}

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


---
HTTP/1.1 200 OK
Content-Length: 9
Content-Type: text/plain; charset=utf-8

Server A

---
HTTP/1.1 200 OK
Content-Length: 9
Content-Type: text/plain; charset=utf-8

Server B

============
`,
		},
	} {
		tc.Run(t)
	}
}

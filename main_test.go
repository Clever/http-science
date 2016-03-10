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
	Diff       Diff
}

func (tc TestCase) Run(t *testing.T) {
	control := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tc.Control.Body)
	}))
	defer control.Close()
	experiment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tc.Experiment.Body)
	}))
	defer experiment.Close()
	var output bytes.Buffer
	if tc.Diff == nil {
		tc.Diff = TextDiff
	}
	scienceServer := httptest.NewServer(Science{
		ControlDial:    strings.Replace(control.URL, "http://", "", -1),
		ExperimentDial: strings.Replace(experiment.URL, "http://", "", -1),
		DiffLog:        log.New(&output, "", 0),
		Diff:           tc.Diff,
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
	return
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
			Output: `===============
--- request ---
GET / HTTP/1.1
Host: {{.Host}}
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


---  diff   ---
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

---
===============
`,
		},

		TestCase{
			Control: Response{
				Body: `{"a":[1,2,3], "b": "asdf"}`,
			},
			Experiment: Response{
				Body: `{"a":[1,2], "b": "a"}`,
			},
			Diff: JSONDiff,
			Output: `===============
--- request ---
GET / HTTP/1.1
Host: {{.Host}}
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


---  diff   ---
["Content-Length"][0]: "27" != "22"
["a"]: []interface {}[3] != []interface {}[2]
["b"]: "asdf" != "a"

===============
`,
		},
	} {
		tc.Run(t)
	}
}

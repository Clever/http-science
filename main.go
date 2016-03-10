package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/kr/pretty"
)

// IgnoreHeaders is a list of headers that should be ignored when diffing two responses.
var IgnoreHeaders = []string{"Date", "Etag", "Content-Length", "Transfer-Encoding"}

// Science is an http.Handler that forwards requests it receives to two places, logging any
// difference in response.
type Science struct {
	ControlDial    string
	ExperimentDial string
	DiffLog        *log.Logger
	Diff           Diff
}

// Diff performs a diff on two http responses and returns a string diff.
type Diff func(control *http.Response, experiment *http.Response) (string, error)

// TextDiff does an equality check on the response text.
func TextDiff(control *http.Response, experiment *http.Response) (string, error) {
	controlDump, err := httputil.DumpResponse(control, true)
	if err != nil {
		return "", fmt.Errorf("error dumping response: %s", err)
	}
	experimentDump, err := httputil.DumpResponse(experiment, true)
	if err != nil {
		return "", fmt.Errorf("error dumping response: %s", err)
	}
	if string(controlDump) != string(experimentDump) {
		return fmt.Sprintf(`---
%s
---
%s
---`, string(controlDump), string(experimentDump)), nil
	}
	return "", nil

}

// JSONDiff does a text diff of headers, but parses the body as JSON.
func JSONDiff(control *http.Response, experiment *http.Response) (string, error) {
	var controlBody, experimentBody map[string]interface{}
	controlDec := json.NewDecoder(control.Body)
	if err := controlDec.Decode(&controlBody); err != nil {
		return "", fmt.Errorf("error decoding control: %s", err)
	}
	experimentDec := json.NewDecoder(experiment.Body)
	if err := experimentDec.Decode(&experimentBody); err != nil {
		return "", fmt.Errorf("error decoding experiment: %s", err)
	}

	diff := ""
	if control.Status != experiment.Status {
		diff += fmt.Sprintf("status diff: '%s' vs. '%s'\n", control.Status, experiment.Status)
	}
	if control.Proto != experiment.Proto {
		diff += fmt.Sprintf("proto diff: '%s' vs. '%s'\n", control.Proto, experiment.Proto)
	}
	if headerDiff := pretty.Diff(control.Header, experiment.Header); len(headerDiff) > 0 {
		diff += strings.Join(headerDiff, "\n") + "\n"
	}
	if bodyDiff := pretty.Diff(controlBody, experimentBody); len(bodyDiff) > 0 {
		diff += strings.Join(bodyDiff, "\n") + "\n"
	}
	return diff, nil
}

// forwardRequest forwards a request to an http server and returns the raw HTTP response.
// It also removes the Date header from the returned response data so you can diff it against other
// responses.
func forwardRequest(r *http.Request, addr string) (net.Conn, *http.Response, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("error establishing tcp connection to %s: %s", addr, err)
	}
	read := bufio.NewReader(conn)
	if err = r.WriteProxy(conn); err != nil {
		return nil, nil, fmt.Errorf("error initializing write proxy to %s: %s", addr, err)
	}
	resp, err := http.ReadResponse(read, r)
	return conn, resp, err
	/*
		res, err := http.ReadResponse(read, r)
		if err != nil {
			return "", fmt.Errorf("error reading response from %s: %s", addr, err)
		}
		defer res.Body.Close()
		delete(res.Header, "Date")
		resDump, err := httputil.DumpResponse(res, true)
		if err != nil {
			return "", fmt.Errorf("error dumping response from %s: %s", addr, err)
		}
		return string(resDump), nil
	*/
}

func (s Science) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// save request for potential diff logging
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
	}
	req := string(reqDump)

	// forward requests to control and experiment, diff response
	var resControl *http.Response
	var resExperiment *http.Response
	var connControl net.Conn
	var connExperiment net.Conn
	if connControl, resControl, err = forwardRequest(r, s.ControlDial); err != nil {
		log.Printf("error forwarding request to control: %s", err)
		return
	}
	defer resControl.Body.Close()
	defer connControl.Close()
	if connExperiment, resExperiment, err = forwardRequest(r, s.ExperimentDial); err != nil {
		log.Printf("error forwarding request to experiment: %s", err)
		return
	}
	defer resExperiment.Body.Close()
	defer connExperiment.Close()

	for _, header := range IgnoreHeaders {
		resControl.Header.Del(header)
		resExperiment.Header.Del(header)
	}

	diff, err := s.Diff(resControl, resExperiment)
	if err != nil {
		log.Printf("error generating diff: %s", err)
		return
	}

	if diff != "" {
		// log diff
		s.DiffLog.Printf(`===============
--- request ---
%s
---  diff   ---
%s
===============
`, req, diff)

	}

	// return 200 no matter what
	fmt.Fprintf(w, "OK")
}

func main() {
	for _, env := range []string{"CONTROL", "EXPERIMENT"} {
		if os.Getenv(env) == "" {
			log.Fatalf("%s required", env)
		}
	}
	diff := TextDiff
	if os.Getenv("DIFF") == "json" {
		diff = JSONDiff
	}
	log.Fatal(http.ListenAndServe(":80", Science{
		ControlDial:    os.Getenv("CONTROL"),
		ExperimentDial: os.Getenv("EXPERIMENT"),
		Diff:           diff,
		DiffLog:        log.New(os.Stdout, "", 0),
	}))
}

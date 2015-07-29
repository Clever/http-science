package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
)

// Science is an http.Handler that forwards requests it receives to two places, logging any
// difference in response.
type Science struct {
	ControlDial    string
	ExperimentDial string
	DiffLog        *log.Logger
}

// forwardRequest forwards a request to an http server and returns the raw HTTP response.
// It also removes the Date header from the returned response data so you can diff it against other
// responses.
func forwardRequest(r *http.Request, addr string) (string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("error establishing tcp connection to %s: %s", addr, err)
	}
	defer conn.Close()
	read := bufio.NewReader(conn)
	if err = r.WriteProxy(conn); err != nil {
		return "", fmt.Errorf("error initializing write proxy to %s: %s", addr, err)
	}
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
}

func (s Science) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// save request for potential diff logging
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
	}
	req := string(reqDump)

	// forward requests to control and experiment, diff response
	var resControl string
	var resExperiment string
	if resControl, err = forwardRequest(r, s.ControlDial); err != nil {
		log.Printf("error forwarding request to control: %s", err)
		return
	}
	if resExperiment, err = forwardRequest(r, s.ExperimentDial); err != nil {
		log.Printf("error forwarding request to experiment: %s", err)
		return
	}

	if resControl != resExperiment {
		s.DiffLog.Printf(`=== diff ===
%s
---
%s
---
%s
============
`, req, resControl, resExperiment)
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
	log.Fatal(http.ListenAndServe(":80", Science{
		ControlDial:    os.Getenv("CONTROL"),
		ExperimentDial: os.Getenv("EXPERIMENT"),
		DiffLog:        log.New(os.Stdout, "", 0),
	}))
}

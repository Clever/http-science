package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Clever/http-science/config"
	"github.com/Clever/http-science/finalize"
	"github.com/Clever/http-science/getfiles"
	"github.com/Clever/http-science/gor"
	"github.com/Clever/http-science/science"
	"github.com/Clever/http-science/validate"
	"gopkg.in/Clever/kayvee-go.v3/logger"
)

func main() {
	var err error
	var handler http.Handler

	if len(os.Args) != 2 {
		finalize.LogAndExitIfErr(fmt.Errorf("Expected 2 args, got %d", len(os.Args)), "incorrect-number-args", os.Args)
	}
	payloadBuffer := []byte(os.Args[1])
	payload := new(config.Payload)

	err = json.Unmarshal(payloadBuffer, payload)
	finalize.LogAndExitIfErr(err, "unmarshal-payload-failed", string(payloadBuffer))

	payload, err = validate.Payload(payload)
	finalize.LogAndExitIfErr(err, "invalid-payload", payload)

	switch payload.JobType {
	case "load":
		handler = setupLoad(payload)
	case "correctness":
		handler, err = setupCorrectness(payload)
	}
	finalize.LogAndExitIfErr(err, "setup-failed", payload)

	doScience(handler, payload)
}

// setupCorrectness returns the handler for a correctness test
func setupCorrectness(payload *config.Payload) (http.Handler, error) {
	f, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return nil, err
	}

	science.Res = science.Results{
		Reqs:    0,
		Codes:   map[int]map[int]int{},
		Mutex:   &sync.Mutex{},
		Diffs:   0,
		DiffLog: f,
	}
	handler := science.CorrectnessTest{
		ControlURL:    payload.ControlURL,
		ExperimentURL: payload.ExperimentURL,
	}
	return handler, nil
}

// setupLoad returns the handler for a load test
func setupLoad(payload *config.Payload) http.Handler {
	science.Res = science.Results{
		Reqs:  0,
		Mutex: &sync.Mutex{},
	}
	handler := science.LoadTest{
		URL: payload.LoadURL,
	}
	return handler
}

// doScience sends the stored requests to the provided handler
func doScience(handler http.Handler, payload *config.Payload) {
	startTime := time.Now()

	// Start up server to handle the requests coming from gor
	go func() {
		err := http.ListenAndServe(":8000", handler)
		finalize.LogAndExitIfErr(err, "server-crashed", nil)
	}()

	// Keep a 10 file buffer for gor
	files := make(chan string, 10)
	go func() {
		err := getfiles.AddFilesToChan(payload, files)
		finalize.LogAndExitIfErr(err, "getting-files-failed", nil)
		// Out of files. Wait until chan empty and then exit
		waitAndExit(startTime, files, payload)
	}()

	// Run gor on those files
	for {
		curFile := <-files
		err := gor.RunGor(curFile, payload)
		finalize.LogAndExitIfErr(err, "gor-failed", nil)
		config.KV.InfoD("progress", logger.M{
			"exp_url":      payload.ExperimentURL,
			"control_url":  payload.ControlURL,
			"load_url":     payload.LoadURL,
			"reqs":         science.Res.Reqs,
			"diffs":        science.Res.Diffs,
			"last_gorfile": curFile,
		})
		if science.Res.Reqs >= payload.Reqs {
			err := finalize.LogResults(startTime, payload)
			finalize.LogAndExitIfErr(err, "logging-results-failed", nil)
			os.Exit(0)
		}
	}
}

func waitAndExit(startTime time.Time, files chan string, payload *config.Payload) {
	for len(files) > 0 {
		time.Sleep(1 * time.Second)
	}
	err := finalize.LogResults(startTime, payload)
	finalize.LogAndExitIfErr(err, "no-files-logging-results-failed", nil)
	finalize.LogAndExitIfErr(fmt.Errorf("Ran out of files"), "out-of-files", nil)
}

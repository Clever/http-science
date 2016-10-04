package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Clever/http-science/config"
	"github.com/Clever/http-science/email"
	"github.com/Clever/http-science/getfiles"
	"github.com/Clever/http-science/gor"
	"github.com/Clever/http-science/science"
	"github.com/Clever/http-science/validate"
	"gopkg.in/Clever/kayvee-go.v3/logger"
	"gopkg.in/Clever/pathio.v3"
)

func main() {
	var err error
	var handler http.Handler

	if len(os.Args) != 2 {
		config.LogAndExitIfErr(fmt.Errorf("Expected 2 args, got %d", len(os.Args)), "incorrect-number-args", os.Args)
	}
	payloadBuffer := []byte(os.Args[1])
	payload := new(config.Payload)

	err = json.Unmarshal(payloadBuffer, payload)
	config.LogAndExitIfErr(err, "unmarshal-payload-failed", string(payloadBuffer))

	payload, err = validate.Payload(payload)
	config.LogAndExitIfErr(err, "invalid-payload", payload)

	switch payload.JobType {
	case "load":
		handler = setupLoad(payload)
	case "correctness":
		handler, err = setupCorrectness(payload)
	}
	config.LogAndExitIfErr(err, "setup-failed", payload)

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
		config.LogAndExitIfErr(err, "server-crashed", nil)
	}()

	// Keep a 10 file buffer for gor
	files := make(chan string, 10)
	go func() {
		err := getfiles.AddFilesToChan(payload, files)
		config.LogAndExitIfErr(err, "getting-files-failed", nil)
		// Out of files. Wait until chan empty and then exit
		waitAndExit(startTime, files, payload)
	}()

	// Run gor on those files
	for {
		err := gor.RunGor(<-files, payload)
		config.LogAndExitIfErr(err, "gor-failed", nil)
		config.KV.InfoD("progress", logger.M{
			"exp_url":     payload.ExperimentURL,
			"control_url": payload.ControlURL,
			"load_url":    payload.LoadURL,
			"reqs":        science.Res.Reqs,
			"diffs":       science.Res.Diffs,
		})
		if science.Res.Reqs >= payload.Reqs {
			err := logResults(startTime, payload)
			config.LogAndExitIfErr(err, "logging-results-failed", nil)
			os.Exit(0)
		}
	}
}

func waitAndExit(startTime time.Time, files chan string, payload *config.Payload) {
	for len(files) > 0 {
		time.Sleep(1 * time.Second)
	}
	err := logResults(startTime, payload)
	config.LogAndExitIfErr(err, "no-files-logging-results-failed", nil)
	config.LogAndExitIfErr(fmt.Errorf("Ran out of files"), "out-of-files", nil)
}

func logResults(startTime time.Time, payload *config.Payload) error {
	log.Printf("%d reqs in %v seconds", science.Res.Reqs, time.Since(startTime))

	if payload.JobType == "correctness" {
		science.Res.Mutex.Lock()
		log.Printf("Results %#v", science.Res.Codes)
		science.Res.Mutex.Unlock()
		log.Printf("%d Diffs using weak compare: %t", science.Res.Diffs, config.WeakCompare)

		// Assert difflog is a file - we use the fact that it is a ReadWriter in the tests
		diffLog, ok := science.Res.DiffLog.(*os.File)
		if !ok {
			config.LogAndExitIfErr(fmt.Errorf("Could not assert to be file"), "type-assertion-failed", nil)
		}
		// Close to prevent data being written during the request
		err := diffLog.Close()
		config.LogAndExitIfErr(err, "closing-file-failed", nil)
		// Open for reading
		diffLog, err = os.Open(diffLog.Name())
		config.LogAndExitIfErr(err, "open-difflog-failed", nil)
		err = pathio.WriteReader(payload.DiffLoc, diffLog)
		config.LogAndExitIfErr(err, "pathio-write-failed", nil)
	}

	if payload.Email != "" {
		err := email.SendEmail(payload, time.Since(startTime), science.Res)
		if err != nil {
			return err
		}
	}
	return nil
}

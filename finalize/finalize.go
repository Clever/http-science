package finalize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Clever/http-science/config"
	"github.com/Clever/http-science/science"
	"gopkg.in/Clever/pathio.v3"
)

// LogResults logs the results of science to the places given in the config
func LogResults(startTime time.Time, payload *config.Payload) error {
	log.Printf("%d reqs in %v seconds", science.Res.Reqs, time.Since(startTime))

	if payload.JobType == "correctness" {
		science.Res.Mutex.Lock()
		defer science.Res.Mutex.Unlock()
		log.Printf("Results %#v", science.Res.Codes)
		log.Printf("%d Diffs using weak compare: %t", science.Res.Diffs, config.WeakCompare)

		// Assert difflog is a file - we use the fact that it is a ReadWriter in the tests
		diffLog, ok := science.Res.DiffLog.(*os.File)
		if !ok {
			LogAndExitIfErr(fmt.Errorf("Could not assert to be file"), "type-assertion-failed", nil)
		}
		// Close to prevent data being written during the request
		err := diffLog.Close()
		LogAndExitIfErr(err, "closing-file-failed", nil)
		// Open for reading
		diffLog, err = os.Open(diffLog.Name())
		LogAndExitIfErr(err, "open-difflog-failed", nil)
		err = pathio.WriteReader(payload.DiffLoc, diffLog)
		LogAndExitIfErr(err, "pathio-write-failed", nil)
	}

	if payload.Email != "" {
		err := sendEmail(payload, time.Since(startTime), science.Res)
		if err != nil {
			return err
		}
	}

	if payload.PutbackURL != "" && payload.JobType == "correctness" {
		b, err := json.Marshal(config.PutbackResponse{
			Diffs: science.Res.Diffs,
		})
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPut, payload.PutbackURL, bytes.NewReader(b))
		if err != nil {
			return err
		}
		http.DefaultClient.Do(req)
	}
	return nil
}

package science

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/Clever/http-science/config"
	"gopkg.in/Clever/kayvee-go.v3/logger"
)

// CorrectnessTest is the interface to run correctness tests with
type CorrectnessTest struct {
	ControlURL    string
	ExperimentURL string
}

var errorForwardingControl = []byte("Error forwarding request Control")
var errorForwardingExperiment = []byte("Error forwarding request Experiment")

func (c CorrectnessTest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// save request for potential diff logging
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
	}

	// Body can only be read once so we need to duplicate it
	rControl, rExperiment, err := duplicateRequest(r)
	if err != nil {
		config.KV.ErrorD("duplicating-request-failed", logger.M{"err": err.Error()})
		return
	}

	// These headers can differ in inconsequential ways so we remove them from the response while forwarding
	cleanup := []string{
		"Date", "Content-Length", "Transfer-Encoding", "X-Request-Id", "Etag",
		"Ot-Tracer-Sampled", "Ot-Tracer-Spanid", "Ot-Tracer-Traceid",
	}
	control, err := forwardRequest(rControl, c.ControlURL, cleanup)
	handleForwardErr(control, "control", err)
	experiment, err := forwardRequest(rExperiment, c.ExperimentURL, cleanup)
	handleForwardErr(experiment, "experiment", err)

	hasDiff := !codesAreEqual(control.code, experiment.code) || !headersAreEqual(control.header, experiment.header) || !bodiesAreEqual(control.body, experiment.body)

	Res.Mutex.Lock()
	defer Res.Mutex.Unlock()
	Res.Reqs++

	if hasDiff {
		updateCodes(control.code, experiment.code)
		Res.Diffs++
		Res.DiffLog.Write(
			[]byte(fmt.Sprintf("=== diff ===\n%s\n---\n%s\n---\n%s\n============\n", string(reqDump), control.dump, experiment.dump)),
		)
	}
}

func duplicateRequest(r *http.Request) (*http.Request, *http.Request, error) {
	b1, b2 := new(bytes.Buffer), new(bytes.Buffer)
	w := io.MultiWriter(b1, b2)
	_, err := io.Copy(w, r.Body)
	if err != nil {
		return nil, nil, err
	}
	defer r.Body.Close()

	r1, r2 := *r, *r
	r1.Body = ioutil.NopCloser(b1)
	r2.Body = ioutil.NopCloser(b2)

	return &r1, &r2, err
}

func updateCodes(codeControl, codeExperiment int) {
	if _, ok := Res.Codes[codeExperiment][codeControl]; !ok {
		if _, ok := Res.Codes[codeExperiment]; !ok {
			Res.Codes[codeExperiment] = map[int]int{}
		}
		Res.Codes[codeExperiment][codeControl] = 0
	}
	Res.Codes[codeExperiment][codeControl]++
}

func handleForwardErr(res *forwardedRequest, which string, err error) {
	if err != nil {
		config.KV.ErrorD(fmt.Sprintf("forwarding-to-%s", which), logger.M{"err": err.Error()})
		res.code = -1
		if which == "control" {
			res.body = errorForwardingControl
			res.dump = string(errorForwardingControl)
		} else {
			res.body = errorForwardingExperiment
			res.dump = string(errorForwardingExperiment)
		}
	}
}

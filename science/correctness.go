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

var errorForwardingControl = "Error forwarding request Control"
var errorForwardingExperiment = "Error forwarding request Experiment"

func (c CorrectnessTest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// save request for potential diff logging
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
	}
	var resControl, resExperiment string
	var codeControl, codeExperiment int
	cleanup := []string{"Date", "Content-Length", "Transfer-Encoding"}

	rControl, rExp, err := duplicateRequest(r)
	if err != nil {
		config.KV.ErrorD("duplicating-request-failed", logger.M{"err": err.Error()})
		return
	}

	if resControl, codeControl, err = forwardRequest(rControl, c.ControlURL, cleanup); err != nil {
		config.KV.ErrorD("forwarding-to-control", logger.M{"err": err.Error()})
		resControl = errorForwardingControl
		codeControl = -1
	}
	if resExperiment, codeExperiment, err = forwardRequest(rExp, c.ExperimentURL, cleanup); err != nil {
		config.KV.ErrorD("forwarding-to-exp", logger.M{"err": err.Error()})
		resExperiment = errorForwardingExperiment
		codeExperiment = -1
	}

	Res.Mutex.Lock()
	defer Res.Mutex.Unlock()
	Res.Reqs++

	if resControl != resExperiment || codeControl != codeExperiment {
		if _, ok := Res.Codes[codeExperiment][codeControl]; !ok {
			if _, ok := Res.Codes[codeExperiment]; !ok {
				Res.Codes[codeExperiment] = map[int]int{}
			}
			Res.Codes[codeExperiment][codeControl] = 0
		}
		Res.Codes[codeExperiment][codeControl]++

		Res.Diffs++
		Res.DiffLog.Write(
			[]byte(fmt.Sprintf("=== diff ===\n%s\n---\n%s\n---\n%s\n============\n", string(reqDump), resControl, resExperiment)),
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

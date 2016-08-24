package science

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// CorrectnessTest is the interface to run correctness tests with
type CorrectnessTest struct {
	ControlURL    string
	ExperimentURL string
}

func (c CorrectnessTest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// save request for potential diff logging
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
	}

	var resControl, resExperiment string
	var codeControl, codeExperiment int
	cleanup := []string{"Date", "Content-Length", "Transfer-Encoding"}

	if resControl, codeControl, err = forwardRequest(r, c.ControlURL, cleanup); err != nil {
		resControl = "Error forwarding request"
	}
	if resExperiment, codeExperiment, err = forwardRequest(r, c.ExperimentURL, cleanup); err != nil {
		resExperiment = "Error forwarding request"
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

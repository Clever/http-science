package config

import (
	"os"
	"sync"

	"gopkg.in/Clever/kayvee-go.v3/logger"
)

// KV is this worker's Kayvee logger
var KV = logger.New("http-science")

// WeakCompare if set to true by the payload allows arrays to be out of order in the json comparison
var WeakCompare = false

// IgnoredHeaders are the headers we ignore diffs on
var IgnoredHeaders []string

// Concurrency is the max number of concurrent requests and a mutex. Ignored if value < 0
var Concurrency = struct {
	Value int
	Mutex *sync.Mutex
}{
	Value: -1,
	Mutex: &sync.Mutex{},
}

// Payload is the payload specifiying info for a load test
type Payload struct {
	// Required
	JobType     string `json:"job_type"`
	ServiceName string `json:"service_name"`
	// Only Correctness
	ExperimentEnv  string   `json:"experiment_env"`
	ControlEnv     string   `json:"control_env"`
	ExperimentURL  string   // initialized in validate.go
	ControlURL     string   // initialized in validate.go
	DiffLoc        string   `json:"diff_loc"`
	WeakCompare    bool     `json:"weak_equal"`
	IgnoredHeaders []string `json:"ignored_headers"`
	// Only Load
	LoadEnv string `json:"load_env"`
	LoadURL string // initialized in validate.go
	Speed   int    `json:"speed"`
	// Optional
	Concurrency      int    `json:"concurrency"`
	Reqs             int    `json:"reqs"`
	JobNumber        int    `json:"job_number"`
	TotalJobs        int    `json:"total_jobs"`
	StartBefore      string `json:"start_before"`
	Methods          string `json:"methods"`
	Email            string `json:"email"`
	DisallowURLRegex string `json:"disallow_url_regex"`
	AllowURLRegex    string `json:"allow_url_regex"`
}

// LogAndExitIfErr KV logs and exits with code 1 if there is an error
func LogAndExitIfErr(err error, title string, extra interface{}) {
	if err != nil {
		KV.ErrorD(title, logger.M{
			"payload": extra,
			"error":   err.Error(),
		})
		os.Exit(1)
	}
}

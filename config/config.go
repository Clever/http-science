package config

import (
	"gopkg.in/Clever/kayvee-go.v3/logger"
	"os"
)

// KV is this worker's Kayvee logger
var KV = logger.New("http-science")

// WeakCompare if set to true by the payload allows arrays to be out of order in the json comparison
var WeakCompare = false

// Payload is the payload specifiying info for a load test
type Payload struct {
	// Required
	JobType  string `json:"job_type"`
	S3Bucket string `json:"s3_bucket"`
	// Only Correctness
	ExperimentURL string `json:"experiment_url"`
	ControlURL    string `json:"control_url"`
	DiffLoc       string `json:"diff_loc"`
	WeakCompare   bool   `json:"weak_equal"`
	// Only Load
	LoadURL string `json:"load_url"`
	// Optional
	FilePrefix  string `json:"file_prefix"`
	Reqs        int    `json:"reqs"`
	Speed       int    `json:"speed"`
	JobNumber   int    `json:"job_number"`
	TotalJobs   int    `json:"total_jobs"`
	StartBefore string `json:"start_before"`
	Methods     string `json:"methods"`
	Email       string `json:"email"`
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

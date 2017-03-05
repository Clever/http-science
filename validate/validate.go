package validate

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Clever/http-science/config"
)

// Payload validates the payload
func Payload(payload *config.Payload) (*config.Payload, error) {

	// Must have job_type and the appropriate urls
	switch payload.JobType {
	case "load":
		if payload.LoadURL == "" {
			return nil, fmt.Errorf("Payload must contain 'load_url' if job_type is load")
		}
		if payload.Speed != 0 && payload.Concurrency != 0 {
			return nil, fmt.Errorf("Payload can't contain both speed an concurrency")
		}
	case "correctness":
		if payload.ExperimentURL == "" || payload.ControlURL == "" {
			return nil, fmt.Errorf("Payload must contain 'experiment_url' and 'control_url' if job_type is correctness")
		}
		if payload.DiffLoc == "" {
			return nil, fmt.Errorf("Payload must contain 'diff_loc' if job_type is correctness")
		}
		if payload.Speed != 0 {
			return nil, fmt.Errorf("Payload can't contain speed if job_type is correctness. Use concurrency")
		}

	default:
		return nil, fmt.Errorf("Payload.job_type must be 'load' or 'correctness', got %s", payload.JobType)
	}

	// s3 bucket required, and must just be the bucket
	if payload.S3Bucket == "" {
		return nil, fmt.Errorf("Payload must contain 's3_bucket'")
	}
	match, err := regexp.MatchString("/", payload.S3Bucket)
	if err != nil {
		return nil, err
	} else if match {
		return nil, fmt.Errorf("s3_bucket should not contain /. Just want the bucket name. Got: %s", payload.S3Bucket)
	}

	// file prefix should not start or end with /
	if payload.FilePrefix != "" && (payload.FilePrefix[0] == '/' || payload.FilePrefix[len(payload.FilePrefix)-1] == '/') {
		return nil, fmt.Errorf("file_prefix should not start or end with slash if used with 's3_bucket'")
	}

	// Set default speed
	if payload.Concurrency != 0 {
		payload.Speed = 10000 // set really high speed, we will control with concurrency
		config.Concurrency.Value = payload.Concurrency
	} else if payload.Speed == 0 {
		payload.Speed = 100
	}
	// Set default reqs
	if payload.Reqs == 0 {
		payload.Reqs = 1000
	}
	// Only replay GETs unless specified
	if payload.Methods == "" {
		payload.Methods = "GET"
	}

	// Set job_number and total_jobs to 1 if they are unset, return an error if one is set and the other not
	if (payload.JobNumber == 0) != (payload.TotalJobs == 0) {
		return nil, fmt.Errorf("Can't have JobNumber without TotalJobs and viceVersa: JobNumber %d, TotalJobs %d",
			payload.JobNumber, payload.TotalJobs)
	} else if payload.JobNumber == 0 {
		payload.JobNumber = 1
		payload.TotalJobs = 1
	}

	// Set StartBefore to the future if not set, return error if not in the correct format
	if payload.StartBefore == "" {
		payload.StartBefore = "9999/99/99:99"
	}
	match, err = regexp.MatchString("^[0-9]{4}/[0-9]{2}/[0-9]{2}:[0-9]{2}$", payload.StartBefore)
	if err != nil {
		return nil, err
	} else if !match {
		return nil, fmt.Errorf("start_before not in correct format. Expected 'yyyy/mm/dd:hh', got: %s", payload.StartBefore)
	}

	// If email set, need mandrill key
	if payload.Email != "" && os.Getenv("MANDRILL_KEY") == "" {
		return nil, fmt.Errorf("email given but no MANDRILL_KEY")
	}

	config.WeakCompare = payload.WeakCompare
	config.IgnoredHeaders = payload.IgnoredHeaders
	return payload, nil
}

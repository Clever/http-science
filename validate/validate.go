package validate

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Clever/http-science/config"
)

// Payload validates the payload
func Payload(payload *config.Payload) (*config.Payload, error) {
	if payload.ServiceName == "" {
		return nil, fmt.Errorf("Payload must contain 'service_name'")
	}

	// Must have job_type and the appropriate urls
	switch payload.JobType {
	case "load":
		if payload.LoadEnv == "" {
			return nil, fmt.Errorf("Payload must contain 'load_env' if job_type is load")
		}
		if payload.Speed != 0 && payload.Concurrency != 0 {
			return nil, fmt.Errorf("Payload can't contain both speed an concurrency")
		}
		podID := ""
		if payload.PodID != "" {
			podID = fmt.Sprintf("--%s", payload.PodID)
		}

		payload.LoadURL = fmt.Sprintf("https://%s--%s%s.int.clever.com:443", payload.LoadEnv, payload.ServiceName, podID)
	case "correctness":
		port := "443"
		if payload.Port != "" {
			port = payload.Port
		}
		if payload.ExperimentEnv == "" || payload.ControlEnv == "" {
			return nil, fmt.Errorf("Payload must contain 'experiment_env' and 'control_env' if job_type is correctness")
		}
		if payload.DiffLoc == "" {
			return nil, fmt.Errorf("Payload must contain 'diff_loc' if job_type is correctness")
		}
		if payload.Speed != 0 {
			return nil, fmt.Errorf("Payload can't contain speed if job_type is correctness. Use concurrency")
		}
		podID := ""
		if payload.PodID != "" {
			podID = fmt.Sprintf("--%s", payload.PodID)
		}
		payload.ControlURL = fmt.Sprintf("https://%s--%s%s.int.clever.com:%s", payload.ControlEnv, payload.ServiceName, podID, port)
		payload.ExperimentURL = fmt.Sprintf("https://%s--%s%s.int.clever.com:%s", payload.ExperimentEnv, payload.ServiceName, podID, port)
	default:
		return nil, fmt.Errorf("Payload.job_type must be 'load' or 'correctness', got %s", payload.JobType)
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
	match, err := regexp.MatchString("^[0-9]{4}/[0-9]{2}/[0-9]{2}:[0-9]{2}$", payload.StartBefore)
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

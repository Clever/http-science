package finalize

import (
	"os"

	"github.com/Clever/http-science/config"

	"gopkg.in/Clever/kayvee-go.v3/logger"
)

// LogAndExitIfErr KV logs and exits with code 1 if there is an error
func LogAndExitIfErr(err error, title string, extra interface{}) {
	if err != nil {
		config.KV.ErrorD(title, logger.M{
			"payload": extra,
			"error":   err.Error(),
		})
		os.Exit(1)
	}
}

package gor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Clever/http-science/config"
)

// RunGor runs gor on the file, sending requests to the local server
func RunGor(file string, payload *config.Payload) error {
	args := []string{
		"--verbose",
		"--debug",
		"--input-file", fmt.Sprintf("%s|%d%%", file, payload.Speed),
		"--output-http", "localhost:8000",
	}
	for _, v := range strings.Split(payload.Methods, ",") {
		args = append(args, []string{"--http-allow-method", v}...)
	}
	if payload.DisallowURLRegex != "" {
		for _, v := range strings.Split(payload.DisallowURLRegex, ",") {
			args = append(args, []string{"--http-disallow-url", v}...)
		}
	}

	cmd := exec.Command("gor", args...)
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	// Bit hacky: Gor is done when there is something availble on this pipe.
	buf := make([]byte, 1000)
	stdErr.Read(buf)
	cmd.Process.Kill()
	return nil
}

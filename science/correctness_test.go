package science

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func refreshResults() Results {
	var b []byte
	return Results{
		Reqs:    0,
		Codes:   map[int]map[int]int{},
		Mutex:   &sync.Mutex{},
		Diffs:   0,
		DiffLog: bytes.NewBuffer(b),
	}
}

func compareDiffLog(t *testing.T, scienceServer *httptest.Server, controlResp, expResp string) {
	diff, err := ioutil.ReadAll(Res.DiffLog)
	assert.Nil(t, err)
	split := strings.Split(scienceServer.URL, ":")
	port := split[len(split)-1]
	requestHeaders := fmt.Sprintf("GET / HTTP/1.1\r\nHost: 127.0.0.1:%s\r\nAccept-Encoding: gzip\r\nUser-Agent: Go-http-client/1.1\r\n\r\n", port)
	assert.Equal(t, fmt.Sprintf("=== diff ===\n%s\n---\n%s\n---\n%s\n============\n", requestHeaders, controlResp, expResp), string(diff))
}

func TestCorrectness(t *testing.T) {
	headerResp := "HTTP/1.1 200 OK\r\nConnection: close\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\n"

	controlResp := "control"
	controlRespWithHeaders := fmt.Sprintf(headerResp, controlResp)
	controlHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, controlResp)
		},
	)
	controlServer := httptest.NewTLSServer(controlHandler)
	defer controlServer.Close()

	expResp := "exp"
	expRespWithHeaders := fmt.Sprintf(headerResp, expResp)
	expHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, expResp)
		},
	)
	expServer := httptest.NewTLSServer(expHandler)
	defer expServer.Close()

	// Use same server for both control and exp - no diff
	scienceServer := httptest.NewServer(CorrectnessTest{
		ControlURL:    controlServer.URL,
		ExperimentURL: controlServer.URL,
	})
	Res = refreshResults()

	_, err := http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)
	assert.Equal(t, 0, Res.Diffs)

	// Use different server - expected diff
	scienceServer = httptest.NewServer(CorrectnessTest{
		ControlURL:    controlServer.URL,
		ExperimentURL: expServer.URL,
	})
	Res = refreshResults()

	_, err = http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)
	assert.Equal(t, 1, Res.Diffs)
	assert.Equal(t, map[int]map[int]int{200: map[int]int{200: 1}}, Res.Codes)
	compareDiffLog(t, scienceServer, controlRespWithHeaders, expRespWithHeaders)

	// Send exp to a server that doesn't exist
	scienceServer = httptest.NewServer(CorrectnessTest{
		ControlURL:    controlServer.URL,
		ExperimentURL: "localhost:not_a_port",
	})
	Res = refreshResults()

	_, err = http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)
	assert.Equal(t, 1, Res.Diffs)
	assert.Equal(t, map[int]map[int]int{-1: map[int]int{200: 1}}, Res.Codes)
	compareDiffLog(t, scienceServer, controlRespWithHeaders, errorForwardingExperiment)

	// Send both to a server that doesn't exist
	scienceServer = httptest.NewServer(CorrectnessTest{
		ControlURL:    "localhost:nope",
		ExperimentURL: "localhost:not_a_port",
	})
	Res = refreshResults()

	_, err = http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)
	assert.Equal(t, 1, Res.Diffs)
	assert.Equal(t, map[int]map[int]int{-1: map[int]int{-1: 1}}, Res.Codes)
	compareDiffLog(t, scienceServer, errorForwardingControl, errorForwardingExperiment)

	// Correctly handles GET requests with bodies
	scienceServer = httptest.NewServer(CorrectnessTest{
		ControlURL:    controlServer.URL,
		ExperimentURL: controlServer.URL,
	})
	Res = refreshResults()

	req, err := http.NewRequest("GET", scienceServer.URL, bytes.NewReader([]byte("body")))
	assert.Nil(t, err)
	c := http.Client{}
	_, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)
	assert.Equal(t, 0, Res.Diffs)
}

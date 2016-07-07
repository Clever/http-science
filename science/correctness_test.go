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

func TestCorrectness(t *testing.T) {
	var b []byte
	Res = Results{
		Reqs:    0,
		Codes:   map[int]map[int]int{},
		Mutex:   &sync.Mutex{},
		Diffs:   0,
		DiffLog: bytes.NewBuffer(b),
	}

	controlHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "control")
		},
	)
	controlServer := httptest.NewTLSServer(controlHandler)
	defer controlServer.Close()

	expHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "exp")
		},
	)
	expServer := httptest.NewTLSServer(expHandler)
	defer expServer.Close()

	// Use same server for both control and exp
	scienceServer := httptest.NewServer(CorrectnessTest{
		ControlURL:    controlServer.URL,
		ExperimentURL: controlServer.URL,
	})
	_, err := http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)
	assert.Equal(t, 0, Res.Diffs)

	// Use different server
	scienceServer = httptest.NewServer(CorrectnessTest{
		ControlURL:    controlServer.URL,
		ExperimentURL: expServer.URL,
	})
	_, err = http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 2, Res.Reqs)
	assert.Equal(t, 1, Res.Diffs)
	diff, err := ioutil.ReadAll(Res.DiffLog)
	assert.Nil(t, err)
	split := strings.Split(scienceServer.URL, ":")
	port := split[len(split)-1]
	assert.Equal(t, fmt.Sprintf("=== diff ===\nGET / HTTP/1.1\r\nHost: 127.0.0.1:%s\r\nAccept-Encoding: gzip\r\nUser-Agent: Go-http-client/1.1\r\n\r\n\n---\nHTTP/1.1 200 OK\r\nConnection: close\r\nContent-Type: text/plain; charset=utf-8\r\n\r\ncontrol\n\n---\nHTTP/1.1 200 OK\r\nConnection: close\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nexp\n\n============\n", port), string(diff))
}

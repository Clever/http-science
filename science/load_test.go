package science

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func refreshLoadResults() Results {
	return Results{
		Reqs:  0,
		Mutex: &sync.Mutex{},
	}
}

func TestLoad(t *testing.T) {
	loadHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "not-dead-yet")
		},
	)
	loadServer := httptest.NewTLSServer(loadHandler)
	defer loadServer.Close()

	// Counts request if successful
	Res = refreshLoadResults()
	scienceServer := httptest.NewServer(LoadTest{
		URL: loadServer.URL,
	})
	Res.Reqs = 0
	_, err := http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, Res.Reqs)

	// Doesn't count request if it fails
	Res = refreshLoadResults()
	scienceServer = httptest.NewServer(LoadTest{
		URL: "localhost:not_a_port",
	})
	Res.Reqs = 0
	_, err = http.Get(scienceServer.URL)
	assert.Nil(t, err)
	assert.Equal(t, 0, Res.Reqs)
}

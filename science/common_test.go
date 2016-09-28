package science

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForward(t *testing.T) {
	testResp := "test resp"
	testIn := "test req"
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			buf := make([]byte, len(testIn))
			_, err := io.ReadFull(r.Body, buf)
			assert.Nil(t, err)
			assert.Equal(t, testIn, string(buf))
			fmt.Fprintln(w, testResp)
		},
	)
	server := httptest.NewTLSServer(handler)
	defer server.Close()

	r, err := http.NewRequest("GET", "https://www.example.com", strings.NewReader(testIn))
	assert.Nil(t, err)
	res, err := forwardRequest(r, server.URL, []string{})
	assert.Nil(t, err)

	assert.True(t, strings.Contains(string(res.body), testResp))
	assert.Equal(t, 200, res.code)
}

func TestCleanup(t *testing.T) {
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "")
		},
	)
	server := httptest.NewTLSServer(handler)
	defer server.Close()

	for _, cleanup := range []string{"Content-Length", "Date"} {
		r, err := http.NewRequest("GET", "https://www.example.com", nil)
		assert.Nil(t, err)
		res, err := forwardRequest(r, server.URL, []string{cleanup})
		assert.Nil(t, err)

		assert.False(t, strings.Contains(string(res.body), cleanup))
		assert.Equal(t, 200, res.code)
	}
}

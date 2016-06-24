package science

import (
	"log"
	"net/http"
)

// LoadTest is the interface to run load tests with
type LoadTest struct {
	URL string
}

func (l LoadTest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _, err := forwardRequest(r, l.URL, []string{}, false)

	if err != nil {
		log.Printf("Error forwarding request: %s", err)
		return
	}
}

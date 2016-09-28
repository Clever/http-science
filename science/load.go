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
	_, err := forwardRequest(r, l.URL, []string{})
	if err != nil {
		log.Printf("Error forwarding request: %s", err)
		return
	}
	Res.Mutex.Lock()
	Res.Reqs++
	Res.Mutex.Unlock()
}

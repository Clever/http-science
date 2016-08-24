package science

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

// Results records results from science
type Results struct {
	Reqs int
	// Only used for correctness tests
	Codes   map[int]map[int]int
	Mutex   *sync.Mutex
	Diffs   int
	DiffLog io.ReadWriter
}

// Res represents the outcome of science
var Res Results

// forwardRequest forwards a request to an address and returns the raw HTTP response.
// It lets you pass a slice of headers that you want removed to make it easier to compare
// to other responses.
func forwardRequest(r *http.Request, addr string, cleanup []string) (string, int, error) {
	addr = strings.TrimPrefix(addr, "https://")
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true}) // TODO - get tests to work without this
	if err != nil {
		return "", 0, fmt.Errorf("error establishing tcp connection to %s: %s", addr, err)
	}
	defer conn.Close()
	if err = r.Write(conn); err != nil {
		return "", 0, fmt.Errorf("error writing request to %s: %s", addr, err)
	}
	res, err := http.ReadResponse(bufio.NewReader(conn), r)
	if err != nil {
		return "", 0, fmt.Errorf("error reading response from %s: %s", addr, err)
	}
	defer res.Body.Close()
	for _, val := range cleanup {
		delete(res.Header, val)
		if val == "Content-Length" {
			res.ContentLength = 0
		} else if val == "Transfer-Encoding" {
			res.TransferEncoding = nil
		}
	}
	resDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return "", 0, fmt.Errorf("error dumping response from %s: %s", addr, err)
	}
	return string(resDump), res.StatusCode, nil
}

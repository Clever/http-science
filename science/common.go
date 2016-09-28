package science

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
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

type forwardedRequest struct {
	dump   string
	body   []byte
	header http.Header
	code   int
}

// Res represents the outcome of science
var Res Results

// forwardRequest forwards a request to an address and returns the raw HTTP response.
// It lets you pass a slice of headers that you want removed to make it easier to compare
// to other responses.
func forwardRequest(r *http.Request, addr string, cleanup []string) (*forwardedRequest, error) {

	addr = strings.TrimPrefix(addr, "https://")
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true}) // TODO - get tests to work without this
	if err != nil {
		return &forwardedRequest{}, fmt.Errorf("error establishing tcp connection to %s: %s", addr, err)
	}
	defer conn.Close()
	if err = r.Write(conn); err != nil {
		return &forwardedRequest{}, fmt.Errorf("error writing request to %s: %s", addr, err)
	}
	res, err := http.ReadResponse(bufio.NewReader(conn), r)
	if err != nil {
		return &forwardedRequest{}, fmt.Errorf("error reading response from %s: %s", addr, err)
	}

	cleanupHeaders(res, cleanup)

	dump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return &forwardedRequest{}, fmt.Errorf("error dumping response from %s: %s", addr, err)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &forwardedRequest{}, fmt.Errorf("smt")
	}

	return &forwardedRequest{
		dump:   string(dump),
		body:   buf,
		code:   res.StatusCode,
		header: res.Header,
	}, nil
}

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Science is an http.Handler that forwards requests it receives to two places, logging any
// difference in response.
type Science struct {
	ControlProxy                *httputil.ReverseProxy
	ExperimentProxy             *httputil.ReverseProxy
	ExperimentHTTPMethods       []string
	ExperimentHTTPURLPathRegexp *regexp.Regexp
	ExperimentPercent           float64
	DiffLog                     *log.Logger
}

func in(str string, arr []string) bool {
	for _, val := range arr {
		if val == str {
			return true
		}
	}
	return false
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (s Science) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// determine if we're going to sample this request
	sample :=
		(s.ExperimentHTTPMethods == nil || in(r.Method, s.ExperimentHTTPMethods)) &&
			(s.ExperimentHTTPURLPathRegexp == nil || s.ExperimentHTTPURLPathRegexp.MatchString(r.URL.Path)) &&
			(rand.Float64() < s.ExperimentPercent)

	// if not sampling, proxy straight through
	if !sample {
		s.ControlProxy.ServeHTTP(w, r)
		return
	}

	// save request for potential diff logging
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %s", err)
	}
	req := string(reqDump)

	// proxy the request, but record the response for diffing
	crecorder := httptest.NewRecorder()
	s.ControlProxy.ServeHTTP(crecorder, r)
	copyHeader(w.Header(), crecorder.HeaderMap)
	w.WriteHeader(crecorder.Code)
	controlBodyBytes := crecorder.Body.Bytes() // need to read twice, so pull it into memory
	if _, err := io.Copy(w, bytes.NewReader(controlBodyBytes)); err != nil {
		fmt.Printf("error copying data to response: %s", err)
		return
	}

	// make same request to control
	erecorder := httptest.NewRecorder()
	s.ExperimentProxy.ServeHTTP(erecorder, r)

	// diff response by getting the raw HTTP response
	delete(crecorder.HeaderMap, "Date")
	delete(erecorder.HeaderMap, "Date")
	cresponse, err := httputil.DumpResponse(&http.Response{
		StatusCode: crecorder.Code,
		Header:     crecorder.HeaderMap,
		Body:       ioutil.NopCloser(bytes.NewReader(controlBodyBytes)),
	}, true)
	if err != nil {
		fmt.Printf("error dumping control response: %s", err)
		return
	}
	eresponse, err := httputil.DumpResponse(&http.Response{
		StatusCode: erecorder.Code,
		Header:     erecorder.HeaderMap,
		Body:       ioutil.NopCloser(erecorder.Body),
	}, true)
	if err != nil {
		fmt.Printf("error dumping experiment response: %s", err)
		return
	}

	if string(cresponse) != string(eresponse) {
		s.DiffLog.Printf(`=== diff ===
%s
--- control ---
%s
--- experiment ---
%s
============
`, req, cresponse, eresponse)
	}
}

func main() {
	for _, env := range []string{"CONTROL", "EXPERIMENT"} {
		if os.Getenv(env) == "" {
			log.Fatalf("%s required", env)
		}
	}

	// parse urls
	controlurl, err := url.Parse(os.Getenv("CONTROL"))
	if err != nil {
		log.Fatalf("error parsing CONTROL: %s", err)
	}
	experimenturl, err := url.Parse(os.Getenv("EXPERIMENT"))
	if err != nil {
		log.Fatalf("error parsing EXPERIMENT: %s", err)
	}

	// optional env
	var httpmethods []string
	if methods := os.Getenv("EXPERIMENT_HTTP_METHODS"); methods != "" {
		httpmethods = strings.Split(methods, ",")
	}

	var re *regexp.Regexp
	if restring := os.Getenv("EXPERIMENT_HTTP_URL_PATH_REGEXP"); restring != "" {
		if re, err = regexp.Compile(restring); err != nil {
			log.Fatalf("error parsing EXPERIMENT_HTTP_URL_REGEXP: %s", err)
		}
	}

	percent := 0.0
	if percentstring := os.Getenv("EXPERIMENT_PERCENT"); percentstring != "" {
		if percent, err = strconv.ParseFloat(percentstring, 64); err != nil {
			log.Fatalf("error parsing EXPERIMENT_PERCENT: %s", err)
		}
	}

	log.Fatal(http.ListenAndServe(":80", Science{
		ControlProxy:                httputil.NewSingleHostReverseProxy(controlurl),
		ExperimentProxy:             httputil.NewSingleHostReverseProxy(experimenturl),
		ExperimentHTTPMethods:       httpmethods,
		ExperimentHTTPURLPathRegexp: re,
		ExperimentPercent:           percent,
		DiffLog:                     log.New(os.Stdout, "", 0),
	}))
}

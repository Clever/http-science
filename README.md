# http-science

http-science is an HTTP service that sends requests it receives to two locations, compares the responses, and reports any differences.

## Motivation

Changes to existing code often fall into the bucket of having no user-facing effect, e.g. refactors or rewrites.
In theory, tests should give you 100% confidence in rolling out changes like this.
However, in practice there is often a lot of risk associated with these changes, e.g. the operational risk of the code running in a production environment (performance, etc.), or lack of confidence in the tests.

"[Science](http://zachholman.com/talk/move-fast-break-nothing/)" is a pattern GitHub introduced for deploying changes to code paths that should not change the output of that code path.
`http-science` is a tool for doing the same experimentation at the network level.
Assuming you have two versions of an HTTP service deployed, it takes care of sending requests to both and reporting any difference in the HTTP responses.

## Installation

`go get github.com/Clever/http-science`

## Usage

`http-science` the binary takes no arguments. It is configured via environment variables:

* `CONTROL_URL`: Required. URL to proxy requests to. Responses will be considered as the "control" in the experiment, i.e. any deviation from the response returned from this server will be treated as significant.
* `EXPERIMENT_URL`: Required. Responses from this server will be compared to responses from the `CONTROL` server.
* `EXPERIMENT_HTTP_METHODS`: comma-separated HTTP methods to experiment on, e.g. "GET,OPTIONS". Optional. Default is all methods.
* `EXPERIMENT_HTTP_URL_REGEXP`: regex of URLs to experiment on, e.g. `^/api/.*`. Optional. Default is all URLs.
* `EXPERIMENT_PERCENT`: percent of requests that match the above filters to sample from `CONTROL` and send to the experiment, e.g. to send all requests, set this to "100.0". Optional. Default is 0.

Once launched, `http-science` listens for HTTP requests on port 80, and will proxy all requests it receives to `CONTROL`.
A certain percentage of requests will get sent to `EXPERIMENT` in addition to getting sent to `CONTROL`.
If there's a difference in response between `CONTROL` and `EXPERIMENT`, it will be logged.
`http-science` always responds to requests with the response from `CONTROL`.

## Developing

`make test` runs the tests.

`make build` will build binaries for Linux and Mac OS.

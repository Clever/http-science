# http-science

http-science is an http service that forwards requests it receives to two locations, compares the responses, and reports any differences.

## Motivation

Changes to existing code often fall into the bucket of having no user-facing effect, e.g. refactors or rewrites.
In theory, tests should give you 100% confidence in rolling out changes like this.
However, in practice there is often a lot of risk associated with these changes, e.g. the operational risk of the code running in a production environment (performance, etc.), or lack of confidence in the tests.

"[Science](http://zachholman.com/talk/move-fast-break-nothing/)" is a pattern GitHub introduced for deploying changes to code paths that should not change the output of that code path.
`http-science` is a tool for doing the same experimentation at the network level.
Assuming you have two versions of an HTTP service deployed, it takes care of forwarding requests to both and reporting any difference in the HTTP responses.

## Installation

`go get github.com/Clever/http-science`

## Usage

`http-science` the binary takes no arguments, but expects two environment variables:

* `CONTROL`: address (`DNS or IP`:`port`) of an http server. Responses will be considered as the "control" in the experiment, i.e. any deviation from the response returned from this server will be treated as significant.
* `EXPERIMENT`: address of an http server. Responses from this server will be compared to responses to the `CONTROL` server.

Once launched, `http-science` listens for HTTP requests on port 80, and will forward any request it receives to both `CONTROL` and `EXPERIMENT`.
If there is a difference in responses, it will log a report of the difference.

## Vendoring

Please view the [dev-handbook for instructions](https://github.com/Clever/dev-handbook/blob/master/golang/godep.md).

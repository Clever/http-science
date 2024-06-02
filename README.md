# http-science

http-science is a worker that can perform both load and correctness testing. For Clever specific instructions, see [confluence](https://clever.atlassian.net/wiki/display/ENG/Use+http-science).


## Motivation

Changes to existing code often fall into the bucket of having no user-facing effect, e.g. refactors or rewrites. In theory, tests should give you 100% confidence in rolling out changes like this. However, in practice there is often a lot of risk associated with these changes, e.g. the operational risk of the code running in a production environment (performance, etc.), or lack of confidence in the tests.

"[Science](http://zachholman.com/talk/move-fast-break-nothing/)" is a pattern GitHub introduced for deploying changes to code paths that should not change the output of that code path. `http-science` is a tool for doing the same experimentation at the network level.


## Summary

http-science takes traffic captured with [gor](https://github.com/buger/gor) and replays it at the specified URL(s). It recognizes two job types, 'load' and 'correctness'. When running a load test, traffic is replayed at a single URL and the distribution of response codes are logged. When running a correctness test, traffic is replayed simultaneously to a ExperimentURL and a ControlURL. The responses are compared and differences are logged.

http-science expects files to be located at `s3://<s3_bucket>/<file_prefix>/yyyy/mm/dd/hh/filename.gz`. We plan to support local files soon.


## Compiling
1. Ensure that you have properly setup your `$GOPATH`
2. While in your `$GOPATH`, clone the repository & `cd` into it
3. Run `make install_deps`
4. Run `bin/deps ensure` (`bin/deps status` to view the status of project dependencies)
5. Run `make && make build`
6. Run `go build`


## Running

`./http-science $PAYLOAD`

If http-science is running as a gearman worker, you can post through gearman-admin
`echo $PAYLOAD | http POST <gearman-admin-url>/job/http-science`

The PAYLOAD will depend on which type of test you are running

## Load Testing

Assuming that your target is running at <URL>, start a basic load test with PAYLOAD

```
{
  "job_type": "load", // Required
  "service_name": "<SERVICE_NAME>" // Required
  "load_env": "<ENV>", // Required
}
```

The maximum rate that requests can be replayed appears to be ~100 req/s. If you need more than this, running multiple concurrently is suggested. We have not investigated what the bottleneck of this performance is.

## Correctness Testing

Assuming that your control is running at <ControlURL>, and your experiment at <ExperimentURL>, start a basic correctness test with PAYLOAD.

```
{
  "job_type": "correctness", // Required
  "service_name": "<SERVICE_NAME>" // Required
  "control_env": "<ENV>", // Required
  "experiment_env": "<ENV>", // Required
  "diff_loc": "s3://bucket/prefix/file" // Required, can be s3 or local path
}
```


## Optional Params
The following params can be included in the payload for both load and correctness testing to give more control over the test:
```
{
  "start_before": "2016/05/31:23", // Default 9999/99/99:99
  "speed": 300, // Default 100
  "reqs": 1000, // Default 1000
  "job_number": 1, // Default 1. Required if total_jobs defined
  "total_jobs": 1, // Default 1. Required if job_number defined
  "methods": "GET,POST,PATCH", // Default GET
  "email": address // Email address to send results to once job is done
  "disallow_url_regex": url // URLs to ignore, comma separated if multiple
}
```

* file_prefix: Necessary if there are directories between the bucket and your files
* start_before: Only replay requests recorded before this date. Format is yyyy/mm/dd:hh
* speed: The percentage of recorded speed you want to replay the requests at
* reqs: The minimum number of requests you want replayed. In practice we go slightly over this
* job_number: If running multiple workers in parallel, give each one a unique number < total_jobs
* total_jobs: Number of total jobs running in parallel
* methods: The http methods we will forward
* disallow_url_regex: Urls to ignore when analyzing correctness, comma separated if multiple

## Vendoring

Please view the [dev-handbook for instructions](https://github.com/Clever/dev-handbook/blob/master/golang/godep.md).

## Extensions

These features could be added to make this more useful

* Exit once we have a certain number of diffs
* Let gor specify a rate per second rather than a percentage

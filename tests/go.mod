module github.com/emmrk/sse/v2/tests

go 1.20

require github.com/emmrk/sse/v2 v2.0.0-20230709021301-ab9e60510927

require (
	golang.org/x/net v0.17.0 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
)

replace github.com/emmrk/sse/v2 => ./..

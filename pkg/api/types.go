package api

// Metric contains a metrics for a single Job
type Metric struct {
	Job string

	Builds []Build
}

//BuildResult represents a result of the build
type BuildResult string

// BuildResultSuccess indicates success
var BuildResultSuccess BuildResult = "SUCCESS"

// BuildResultFailure indicates failure
var BuildResultFailure BuildResult = "FAILURE"

// BuildResultUnknown indicates unknown
var BuildResultUnknown BuildResult = "UNKNOWN"

// Build represents a single build data
type Build struct {
	Number   int
	Duration int
	Building bool
	Result   BuildResult
}

package api

// BuildNumbers returns slice of build numbers
func BuildNumbers(m *Metric) []int {
	ret := []int{}
	for _, b := range m.Builds {
		ret = append(ret, b.Number)
	}
	return ret
}

// SuccessBuilds is
func SuccessBuilds(m *Metric) int {
	ret := 0
	for _, b := range m.Builds {
		if b.Result == BuildResultSuccess {
			ret++
		}
	}
	return ret
}

// FailedBuilds is
func FailedBuilds(m *Metric) int {
	ret := 0
	for _, b := range m.Builds {
		if b.Result == BuildResultFailure {
			ret++
		}
	}
	return ret
}

// AverageBuildDurationSeconds is
func AverageBuildDurationSeconds(m *Metric) int {
	ret := 0
	count := 0
	for _, b := range m.Builds {
		ret += b.Duration
		count++
	}
	if count == 0 {
		return 0
	}
	return ret / count
}

// AverageSuccessBuildDurationSeconds is
func AverageSuccessBuildDurationSeconds(m *Metric) int {
	ret := 0
	count := 0
	for _, b := range m.Builds {
		if b.Result == BuildResultSuccess {
			ret += b.Duration
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return ret / count
}

// AverageFailedBuildDurationSeconds is
func AverageFailedBuildDurationSeconds(m *Metric) int {
	ret := 0
	count := 0
	for _, b := range m.Builds {
		if b.Result == BuildResultFailure {
			ret += b.Duration
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return ret / count
}

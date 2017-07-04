package collector

import "testing"

func TestGetTestPullRequestMetrics(t *testing.T) {
	job, err := getJenkinsJob("test_pull_request_origin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	metric, err := getMetrics(job, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("metric: %#v", metric)
}

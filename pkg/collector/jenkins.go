package collector

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/mfojtik/jenkins-prometheus/pkg/api"
	jenkins "github.com/yosida95/golang-jenkins"
)

// getBuildNumbers gets the last build numbers from the job
func getBuildNumbers(jobName string) ([]int, error) {
	buildIDRegexp := regexp.MustCompile(":" + jobName + ":(\\d+)")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(jenkinsURL + "/job/" + jobName + "/rssAll")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	matches := buildIDRegexp.FindAllString(string(body), -1)
	if matches == nil {
		return nil, fmt.Errorf("unable to find build numbers in rss")
	}
	ret := []int{}
	for _, m := range matches {
		i, err := strconv.Atoi(strings.TrimPrefix(m, ":"+jobName+":"))
		if err != nil {
			log.Printf("Failed to parse build number %s: %v (will skip)", m, err)
			continue
		}
		ret = append(ret, i)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ret)))
	return ret, nil
}

func getJenkinsJob(name string) (jenkins.Job, error) {
	j := jenkins.NewJenkins(nil, jenkinsURL)
	return j.GetJob(name)
}

func getMetrics(job jenkins.Job, maxBuilds int) (*api.Metric, error) {
	builds, err := getBuildNumbers(job.Name)
	if err != nil {
		return nil, err
	}
	result := api.Metric{Job: job.Name}
	result.Builds = []api.Build{}
	if len(builds) == 0 {
		return &result, nil
	}
	j := jenkins.NewJenkins(nil, jenkinsURL)
	work := func(n int) {
		build, err := j.GetBuild(job, n)
		if err != nil {
			return
		}
		result.Builds = append(result.Builds, api.Build{
			Number:   build.Number,
			Duration: build.Duration,
			Building: build.Building,
			Result:   parseResult(build.Result),
		})
	}
	wg := sync.WaitGroup{}
	workerCount := 0
	maxWorkers := 5
	for i := 0; i < len(builds); i++ {
		if i > maxBuilds {
			wg.Wait()
			break
		}
		workerCount++
		wg.Add(1)
		go func(n int) {
			defer func() { wg.Done(); workerCount-- }()
			work(n)
		}(builds[i])
		if workerCount > maxWorkers {
			wg.Wait()
		}
	}
	wg.Wait()
	return &result, nil
}

func parseResult(result string) api.BuildResult {
	switch result {
	case "FAILURE":
		return api.BuildResultFailure
	case "SUCCESS":
		return api.BuildResultSuccess
	default:
		return api.BuildResultUnknown
	}
}

package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/jasonlvhit/gocron"

	"github.com/mfojtik/jenkins-prometheus/pkg/api"
)

const (
	collectIntervalSeconds = 30
	maxLastBuilds          = 5

	jenkinsURL = "https://ci.openshift.redhat.com/jenkins"
)

// Manager manages the collection of Jenkins metrics
type Manager struct {
	storage     *db.Col
	storageLock sync.Mutex
}

func (m *Manager) startDatabase() error {
	data, err := db.OpenDB(os.Getenv("COLLECTOR_DB_DIR"))
	if err != nil {
		return err
	}
	initialized := false
	for _, name := range data.AllCols() {
		if name == "metrics" {
			initialized = true
			break
		}
	}
	if !initialized {
		log.Printf("database in %q not initialized, creating metrics collection ...", os.Getenv("COLLECTOR_DB_DIR"))
		if err := data.Create("metrics"); err != nil {
			return err
		}
	}
	m.storage = data.Use("metrics")
	m.storage.Index([]string{"job_name"})
	return nil
}

func (m *Manager) collect() error {
	errors := []string{}
	for _, jobName := range api.OriginJobs {
		job, err := getJenkinsJob(jobName)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		metric, err := getMetrics(job, maxLastBuilds)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		if err := m.store(metric); err != nil {
			errors = append(errors, err.Error())
		}
	}
	return fmt.Errorf(strings.Join(errors, ","))
}

func (m *Manager) store(metric *api.Metric) error {
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+metric.Job+`", "in": ["job_name"]}]`), &query)
	queryResult := make(map[int]struct{})
	if err := db.EvalQuery(query, m.storage, &queryResult); err != nil {
		return err
	}
	if len(queryResult) == 0 {
		log.Printf("Inserting new data for job %q", metric.Job)
		_, err := m.storage.Insert(map[string]interface{}{
			"job_name":                                   metric.Job,
			"builds_count":                               len(metric.Builds),
			"success_builds_count":                       api.SuccessBuilds(metric),
			"failed_builds_count":                        api.FailedBuilds(metric),
			"average_per_build_duration_seconds":         api.AverageBuildDurationSeconds(metric),
			"average_per_success_build_duration_seconds": api.AverageSuccessBuildDurationSeconds(metric),
			"average_per_failed_build_duration_seconds":  api.AverageFailedBuildDurationSeconds(metric),
			"build_numbers":                              api.BuildNumbers(metric),
		})
		return err
	}
	if len(queryResult) > 1 {
		log.Fatalf("Job %q seems to have multiple data, aborting", metric.Job)
	}
	for docID := range queryResult {
		readBack, _ := m.storage.Read(docID)
		log.Printf("Query returned document %#v\n", readBack)
	}
	return nil
}

func (m *Manager) work() {
	log.Printf("Starting worker ...")
	errChan := make(chan error)
	// Do not run two collectors at time. Wait until the collection finishes.
	go func() {
		errChan <- m.collect()
	}()
	select {
	case err := <-errChan:
		log.Printf("Failed to collect metrics: %v", err)
	case <-time.After(time.Second*collectIntervalSeconds - 1):
		log.Printf("Timeout while waiting for a worker to finish")
	}
}

// Run runs the collector
func Run() {
	m := &Manager{}
	if err := m.startDatabase(); err != nil {
		log.Fatalf("Error starting collector database: %v", err)
	}
	s := gocron.NewScheduler()
	s.Every(collectIntervalSeconds).Seconds().Do(m.work)
	startCh := s.Start()
	s.RunAll()
	<-startCh
}

package exporter

import (
	"fmt"
	"sync"

	"github.com/healthcheck-watchdog/cmd/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	config       *model.Config
	jobsCounters map[string]*Counter
	Channel      chan model.TaskStatus
	mutex        sync.Mutex
}

type Counter struct {
	job            model.Job
	status         prometheus.Gauge
	downtime       prometheus.Gauge
	messagesCount  prometheus.GaugeVec
	responseTime   prometheus.Gauge
	failedAttempts prometheus.Gauge
	watchdogAction prometheus.Gauge
}

func NewExporter(config *model.Config) *Exporter {
	exporter := Exporter{
		config:  config,
		Channel: make(chan model.TaskStatus),
	}

	// Register counters
	exporter.jobsCounters = make(map[string]*Counter, len(config.Jobs))
	for _, job := range config.Jobs {
		labels := prometheus.Labels(job.MetricLabels)

		downtime := promauto.NewGauge(prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_downtime", job.MetricName),
			Help:        job.Description,
			ConstLabels: prometheus.Labels(labels),
		})
		status := promauto.NewGauge(prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_status", job.MetricName),
			Help:        fmt.Sprintf("%s is working (0: no, 1: yes)", job.Description),
			ConstLabels: prometheus.Labels(labels),
		})
		messagesCount := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_messages_count", job.MetricName),
			Help:        fmt.Sprintf("%s amount of received messages", job.Description),
			ConstLabels: prometheus.Labels(labels),
		}, []string{"uid"})
		responseTime := promauto.NewGauge(prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_response_time", job.MetricName),
			Help:        fmt.Sprintf("%s response time length", job.Description),
			ConstLabels: prometheus.Labels(labels),
		})
		failedAttempts := promauto.NewGauge(prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_failed_attempts_count", job.MetricName),
			Help:        fmt.Sprintf("%s failed attempts count", job.Description),
			ConstLabels: prometheus.Labels(labels),
		})
		watchdogAction := promauto.NewGauge(prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_watchdog_action_count", job.MetricName),
			Help:        fmt.Sprintf("%s amount of watchdog actions", job.Description),
			ConstLabels: prometheus.Labels(labels),
		})
		exporter.jobsCounters[job.Id] = &Counter{
			job:            job,
			downtime:       downtime,
			status:         status,
			messagesCount:  *messagesCount,
			responseTime:   responseTime,
			failedAttempts: failedAttempts,
			watchdogAction: watchdogAction,
		}

		log.Info(
			fmt.Sprintf(
				"Registered counters with base metric name %s for job %s", job.MetricName, job.Id))
	}

	// Channel for task results
	go exporter.resultProcessor(exporter.Channel)

	return &exporter
}

// Process results from tasks
func (ex *Exporter) resultProcessor(resultChan <-chan model.TaskStatus) {
	for result := range resultChan {
		log.Info(fmt.Sprintf("Exporter: Processed result %v", result))
		ex.setCounters(&result)
	}
}

func (ex *Exporter) setCounters(status *model.TaskStatus) {
	ex.mutex.Lock()
	defer ex.mutex.Unlock()

	counters, found := ex.jobsCounters[status.Id]
	if found {
		if status.Status {
			counters.downtime.Set(0)
		} else {
			counters.downtime.Add(float64(counters.job.Timeout))
			counters.failedAttempts.Inc()
		}
		counters.status.Set(boolToFloat64(status.Status))
		if counters.job.Type == "websocket" {
			if status.Status && status.LastResult.Parameters != nil {
				counters.messagesCount.With(
					prometheus.Labels{"uid": status.LastResult.Parameters["uid"].(string)}).Inc()
			}
		}
		counters.responseTime.Set(float64(status.LastResult.Duration))
		if status.Watchdog {
			counters.watchdogAction.Inc()
		}
	}
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

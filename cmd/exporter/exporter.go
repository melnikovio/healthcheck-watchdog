package exporter

import (
	"errors"
	"fmt"

	"github.com/healthcheck-watchdog/cmd/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	config   *model.Config
	counters map[string]*Counter
}

type Counter struct {
	id             string
	status         prometheus.Gauge
	downtime       prometheus.Gauge
	messagesCount  prometheus.GaugeVec
	responseTime   prometheus.Gauge
	watchdogAction prometheus.Gauge
}

func NewExporter(config *model.Config) *Exporter {
	ex := Exporter{
		config: config,
	}

	if config == nil || config.Jobs == nil {
		err := errors.New("missing monitoring tasks")
		log.Error(fmt.Sprintf("Failed to initialize exporter: %s", err.Error()))

		return &ex
	}

	counters := make(map[string]*Counter, len(config.Jobs))
	for i := 0; i < len(config.Jobs); i++ {
		downtime := promauto.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_downtime", config.Jobs[i].Id),
			Help: config.Jobs[i].Description,
		})
		status := promauto.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_status", config.Jobs[i].Id),
			Help: fmt.Sprintf("%s работает (0: нет, 1: да)", config.Jobs[i].Description),
		})
		messagesCount := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_messages_count", config.Jobs[i].Id),
			Help: fmt.Sprintf("%s количество сообщений", config.Jobs[i].Description),
		}, []string{"uid"})
		responseTime := promauto.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_response_time", config.Jobs[i].Id),
			Help: fmt.Sprintf("%s время ответа", config.Jobs[i].Description),
		})
		watchdogAction := promauto.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_watchdog_action_count", config.Jobs[i].Id),
			Help: fmt.Sprintf("%s количество срабатываний watchdog", config.Jobs[i].Description),
		})
		counters[config.Jobs[i].Id] = &Counter{
			id:             config.Jobs[i].Id,
			downtime:       downtime,
			status:         status,
			messagesCount:  *messagesCount,
			responseTime:   responseTime,
			watchdogAction: watchdogAction,
		}

		log.Info(fmt.Sprintf("Registered counter %s", config.Jobs[i].Id))
	}

	ex.counters = counters

	return &ex
}

//todo config
func (ex *Exporter) IncCounter(id string, param string) {
	counter, found := ex.counters[id]
	if found {
		counter.messagesCount.With(prometheus.Labels{"uid":param}).Inc()
	}
}

func (ex *Exporter) SetGauge(id string, value float64) {
	counter, found := ex.counters[id]
	if found {
		counter.responseTime.Set(value)
	}
}

func (ex *Exporter) SetCounter(id string, online bool) {
	counter, found := ex.counters[id]
	if found {
		var onlineVal float64
		if online {
			onlineVal = 1
		} else {
			onlineVal = 0
		}
		counter.downtime.Set(0)

		counter.status.Set(onlineVal)
	}
}

func (ex *Exporter) AddCounter(id string, value int64) {
	counter, found := ex.counters[id]
	if found {
		counter.downtime.Add(float64(value))

		// todo what is this?
		val2 := float64(0)
		statusMetric := dto.Metric{
			Counter: &dto.Counter{
				Value: &val2,
			},
		}
		err := counter.status.Write(&statusMetric)
		if err != nil {
			log.Error(fmt.Sprintf("Error writing metrics: %s", err.Error()))
		}
	}
}

func (ex *Exporter) IncWatchdogActionCounter(id string) {
	counter, found := ex.counters[id]
	if found {
		counter.watchdogAction.Inc()
	}
}

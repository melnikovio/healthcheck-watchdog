package exporter

import (
	"fmt"
	"github.com/healthcheck-exporter/cmd/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	config   *model.Config
	counters []Counter
}

type Counter struct {
	id       string
	status   prometheus.Gauge
	downtime prometheus.Gauge
	messagesCount prometheus.Gauge
	responseTime prometheus.Gauge
}

func NewExporter(config *model.Config) *Exporter {
	ex := Exporter{
		config: config,
	}

	if config != nil {
		counters := make([]Counter, len(config.Jobs))
		for i := 0; i < len(config.Jobs); i++ {
			downtime := promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_downtime", config.Jobs[i].Id),
				Help: config.Jobs[i].Description,
			})
			status := promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_status", config.Jobs[i].Id),
				Help: fmt.Sprintf("%s работает (0: нет, 1: да)", config.Jobs[i].Description),
			})
			messagesCount := promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_messages_count", config.Jobs[i].Id),
				Help: fmt.Sprintf("%s количество сообщений", config.Jobs[i].Description),
			})
			responseTime := promauto.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_response_time", config.Jobs[i].Id),
				Help: fmt.Sprintf("%s время ответа", config.Jobs[i].Description),
			})
			counters[i] = Counter{
				id:       config.Jobs[i].Id,
				downtime: downtime,
				status:   status,
				messagesCount: messagesCount,
				responseTime: responseTime,
			}

			log.Info(fmt.Sprintf("Registered counter %s", config.Jobs[i].Id))
		}

		ex.counters = counters
	}

	return &ex
}

func (ex *Exporter) IncCounter(id string) {
	for i := 0; i < len(ex.counters); i++ {
		if ex.counters[i].id == id {
			ex.counters[i].messagesCount.Inc()
		}
	}
}

func (ex *Exporter) SetGauge(id string, value float64) {
	for i := 0; i < len(ex.counters); i++ {
		if ex.counters[i].id == id {
			ex.counters[i].responseTime.Set(value)
		}
	}
}

func (ex *Exporter) SetCounter(id string, value int64) {
	for i := 0; i < len(ex.counters); i++ {
		if ex.counters[i].id == id {
			//val := float64(value)
			//downtimeMetric := dto.Metric{
			//	Counter: &dto.Counter{
			//		Value: &val,
			//	},
			//}
			//
			//err := ex.counters[i].downtime.Write(&downtimeMetric)
			//if err != nil {
			//	log.Error(fmt.Sprintf("Error writing metrics: %s", err.Error()))
			//}
			ex.counters[i].downtime.Set(0)

			//val2 := float64(0)
			//statusMetric := dto.Metric{
			//	Gauge: &dto.Gauge{
			//		Value: &val2,
			//	},
			//}
			//err = ex.counters[i].status.Write(&statusMetric)
			ex.counters[i].status.Set(1)
			//if err != nil {
			//	log.Error(fmt.Sprintf("Error writing metrics: %s", err.Error()))
			//}
		}
	}
}

func (ex *Exporter) AddCounter(id string, value int64) {
	for i := 0; i < len(ex.counters); i++ {
		if ex.counters[i].id == id {
			val := float64(value)
			ex.counters[i].downtime.Add(val)

			val2 := float64(0)
			statusMetric := dto.Metric{
				Counter: &dto.Counter{
					Value: &val2,
				},
			}
			err := ex.counters[i].status.Write(&statusMetric)
			if err != nil {
				log.Error(fmt.Sprintf("Error writing metrics: %s", err.Error()))
			}
		}
	}
}

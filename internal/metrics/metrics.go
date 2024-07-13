package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	InfoMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "watchtower_vm_availability",
			Help: "Metric that contains info about vm ",
		},
		[]string{"os_project", "hostname", "name", "project_id", "uuid"}, // Список Labels
	)
)

func init() {
	prometheus.MustRegister(InfoMetric)
}

package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
)

var totalMentionsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "webmentiond_mentions_total",
})
var mentionsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "webmentiond_mentions",
}, []string{"status"})

func init() {
	prometheus.MustRegister(totalMentionsGauge)
	prometheus.MustRegister(mentionsGauge)
	// Register version info collector from common/version package
	prometheus.MustRegister(version.NewCollector("webmentiond"))
}

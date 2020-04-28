package server

import (
	"github.com/prometheus/client_golang/prometheus"
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
}

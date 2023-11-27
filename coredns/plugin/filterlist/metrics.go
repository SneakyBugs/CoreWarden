package filterlist

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	listFetchBackoffs = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: name,
		Name:      "list_fetch_backoffs",
		Help:      "Count of backoffs during list fetching.",
	})
	listFetchFailures = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: name,
		Name:      "list_fetch_failures",
		Help:      "Count of failures during list fetching.",
	})
	listFetchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: name,
		Name:      "list_fetches_total",
		Help:      "Count of list fetches performed including failures.",
	})
	requestsBlocked = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: name,
		Name:      "requests_blocked",
		Help:      "Count of requests blocked by filters.",
	})
	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: name,
		Name:      "requests_total",
		Help:      "Count of requests handled by the plugin.",
	})
)

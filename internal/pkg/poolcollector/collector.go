// Package poolcollector provide method to work with clickhouse prometheus metrics.
package poolcollector

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prometheus/client_golang/prometheus"
)

// StatsGetter is an interface that gets sql.DBStats.
// It's implemented by e.g. *sql.DB or *sqlx.DB.
type StatsGetter interface {
	Stats() driver.Stats
}

// StatsCollector implements the prometheus.Collector interface.
type StatsCollector struct {
	sg StatsGetter

	// descriptions of exported metrics
	maxOpenConns *prometheus.Desc
	openConns    *prometheus.Desc
	maxIdleConns *prometheus.Desc
	idleConns    *prometheus.Desc
}

// NewStatsCollector creates a new StatsCollector.
func NewStatsCollector(namespace, subsystem string, constLabels prometheus.Labels, sg StatsGetter) *StatsCollector {
	return &StatsCollector{
		sg: sg,
		maxOpenConns: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_open_conns"),
			"Maximum number of open connections to the database.",
			nil,
			constLabels,
		),
		openConns: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open_conns"),
			"The number of established connections both in use and idle.",
			nil,
			constLabels,
		),
		maxIdleConns: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_idle_conns"),
			"Maximum number of idle connections to the database.",
			nil,
			constLabels,
		),
		idleConns: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle"),
			"The number of idle connections.",
			nil,
			constLabels,
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (c StatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenConns
	ch <- c.openConns
	ch <- c.idleConns
	ch <- c.maxIdleConns
}

// Collect implements the prometheus.Collector interface.
func (c StatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.sg.Stats()

	ch <- prometheus.MustNewConstMetric(
		c.maxOpenConns,
		prometheus.GaugeValue,
		float64(stats.MaxOpenConns),
	)

	ch <- prometheus.MustNewConstMetric(
		c.openConns,
		prometheus.GaugeValue,
		float64(stats.Open),
	)

	ch <- prometheus.MustNewConstMetric(
		c.maxIdleConns,
		prometheus.GaugeValue,
		float64(stats.MaxIdleConns),
	)

	ch <- prometheus.MustNewConstMetric(
		c.idleConns,
		prometheus.GaugeValue,
		float64(stats.Idle),
	)
}

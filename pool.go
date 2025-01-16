package clickhouse

import (
	"context"
	"crypto/tls"
	"embed"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Masterminds/squirrel"
	"github.com/mkbeh/xclick/internal/pkg/poolcollector"
	"github.com/prometheus/client_golang/prometheus"
)

type Pool struct {
	driver.Conn

	id          string
	cfg         *Config
	logger      *slog.Logger
	qBuilder    squirrel.StatementBuilderType
	compression *clickhouse.Compression
	tls         *tls.Config
	proxy       *url.URL
	namespace   string
	labels      prometheus.Labels
	migrations  []embed.FS
}

func NewPool(opts ...Option) (*Pool, error) {
	p := &Pool{
		cfg:      &Config{},
		logger:   slog.Default(),
		qBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	for _, opt := range opts {
		opt.apply(p)
	}

	p.logger = p.logger.With(slog.String("component", "clickhouse"))

	connOpts := parseConfig(p.cfg)
	connOpts.Compression = p.compression
	connOpts.TLS = p.tls

	conn, err := clickhouse.Open(connOpts)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	p.Conn = conn
	p.exposeMetrics()

	if p.cfg.MigrateEnabled {
		for _, fs := range p.migrations {
			if err := applyMigrations(fs, p.cfg.getDSN(), p.logger); err != nil {
				return nil, err
			}
		}
	}

	metricsCollector := poolcollector.NewStatsCollector(p.namespace, "clickhouse", p.labels, conn)
	prometheus.MustRegister(metricsCollector)

	return p, nil
}

func (p *Pool) QueryBuilder() squirrel.StatementBuilderType {
	return p.qBuilder
}

func (p *Pool) Logger() *slog.Logger {
	return p.logger
}

func (p *Pool) Close() error {
	return p.Conn.Close()
}

func (p *Pool) getID() string {
	if p.id == "" {
		return GenerateUUID()
	}
	return p.id
}

func (p *Pool) exposeMetrics() {
	if p.labels == nil {
		p.labels = make(prometheus.Labels)
	}

	p.labels["client_id"] = p.getID()
	p.labels["db"] = p.cfg.DB
	p.labels["shard_id"] = strconv.Itoa(p.cfg.ShardID)
}

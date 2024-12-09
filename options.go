package clickhouse

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// An Option lets you add opts to pberrors interceptors using With* funcs.
type Option interface {
	apply(p *Pool)
}

type optionFunc func(p *Pool)

func (f optionFunc) apply(p *Pool) {
	f(p)
}

func WithLogger(l *slog.Logger) Option {
	return optionFunc(func(p *Pool) {
		if l != nil {
			p.logger = l
		}
	})
}

func WithConfig(config *Config) Option {
	return optionFunc(func(p *Pool) {
		if config != nil {
			p.cfg = config
		}
	})
}

func WithClientID(id string) Option {
	return optionFunc(func(p *Pool) {
		if id != "" {
			p.id = fmt.Sprintf("%s-%s", id, GenerateUUID())
		}
	})
}

func WithMigrations(migrations ...embed.FS) Option {
	return optionFunc(func(p *Pool) {
		if len(migrations) > 0 {
			p.migrations = migrations
		}
	})
}

func WithCompression(compression *clickhouse.Compression) Option {
	return optionFunc(func(p *Pool) {
		p.compression = compression
	})
}

func WithTLS(cfg *tls.Config) Option {
	return optionFunc(func(p *Pool) {
		if cfg != nil {
			p.tls = cfg
		}
	})
}

func WithHTTPProxy(proxy *url.URL) Option {
	return optionFunc(func(p *Pool) {
		p.proxy = proxy
	})
}

// --- metrics ---

func WithMetricsNamespace(ns string) Option {
	return optionFunc(func(p *Pool) {
		if ns != "" {
			p.namespace = ns
		}
	})
}

type Config struct {
	ShardID  int    `envconfig:"CLICKHOUSE_SHARD_ID"`
	Hosts    string `envconfig:"CLICKHOUSE_HOSTS" required:"true"`
	User     string `envconfig:"CLICKHOUSE_USER" required:"true"`
	Password string `envconfig:"CLICKHOUSE_PASSWORD" required:"true"`
	DB       string `envconfig:"CLICKHOUSE_DB" required:"true"`

	MaxOpenConns         int           `envconfig:"CLICKHOUSE_MAX_OPEN_CONNS"`
	MaxIdleConns         int           `envconfig:"CLICKHOUSE_MAX_IDLE_CONNS"`
	ConnMaxLifetime      time.Duration `envconfig:"CLICKHOUSE_CONN_MAX_LIFETIME"`
	DialTimeout          time.Duration `envconfig:"CLICKHOUSE_DIAL_TIMEOUT"`
	ReadTimeout          time.Duration `envconfig:"CLICKHOUSE_READ_TIMEOUT"`
	Debug                bool          `envconfig:"CLICKHOUSE_DEBUG"`
	FreeBufOnConnRelease bool          `envconfig:"CLICKHOUSE_FREE_BUFFER_ON_CONN_RELEASE"`
	InsecureSkipVerify   bool          `envconfig:"CLICKHOUSE_INSECURE_SKIP_VERIFY"`
	BlockBufferSize      uint8         `envconfig:"CLICKHOUSE_BLOCK_BUFFER_SIZE"`
	MaxCompressionBuffer int           `envconfig:"CLICKHOUSE_MAX_COMPRESSION_BUFFER"`

	HttpHeaders map[string]string `envconfig:"CLICKHOUSE_HTTP_HEADERS"`
	HttpUrlPath string            `envconfig:"CLICKHOUSE_HTTP_URL_PATH"`

	// ConnOpenStrategy available strategies: in_order, round_robin, random
	ConnOpenStrategy string         `envconfig:"CLICKHOUSE_CONN_OPEN_STRATEGY"`
	Settings         map[string]any `envconfig:"CLICKHOUSE_SETTINGS"`

	MigrateEnabled bool   `envconfig:"CLICKHOUSE_MIGRATE_ENABLED"`
	MigrateArgs    string `envconfig:"CLICKHOUSE_MIGRATE_ARGS"`
}

func (c *Config) getDSN() []string {
	return formatDSN(c.Hosts, c.User, c.Password, c.DB, c.MigrateArgs)
}

// formatDSN returns clickhouse hosts.
// Template: <dialect>://<user>:<password>@<host>:<port>/<database>?<migrate_args>
// EX.: clickhouse://user:password@localhost:8001/test
func formatDSN(hosts, user, pass, db, args string) []string {
	conns := strings.Split(hosts, ",")
	dsn := make([]string, 0, len(conns))
	for _, conn := range conns {
		dsn = append(dsn, fmt.Sprintf("%s://%s:%s@%s/%s?%s", "clickhouse", user, pass, conn, db, args))
	}
	return dsn
}

func parseConfig(cfg *Config) *clickhouse.Options {
	opts := &clickhouse.Options{
		Addr: strings.Split(cfg.Hosts, ","),
		Auth: clickhouse.Auth{
			Database: cfg.DB,
			Username: cfg.User,
			Password: cfg.Password,
		},
		MaxOpenConns:         32,
		MaxIdleConns:         8,
		ConnMaxLifetime:      time.Hour * 1,
		DialTimeout:          time.Second * 10,
		ReadTimeout:          time.Second * 10,
		Debug:                cfg.Debug,
		FreeBufOnConnRelease: cfg.FreeBufOnConnRelease,
		BlockBufferSize:      2,
		MaxCompressionBuffer: 10485760,
		HttpHeaders:          cfg.HttpHeaders,
		HttpUrlPath:          cfg.HttpUrlPath,
		ConnOpenStrategy:     getConnOpenStrategy(cfg.ConnOpenStrategy),
	}

	if cfg.MaxOpenConns > 0 {
		opts.MaxOpenConns = cfg.MaxOpenConns
	}

	if cfg.MaxIdleConns > 0 {
		opts.MaxIdleConns = cfg.MaxIdleConns
	}

	if cfg.ConnMaxLifetime > 0 {
		opts.ConnMaxLifetime = cfg.ConnMaxLifetime
	}

	if cfg.DialTimeout > 0 {
		opts.DialTimeout = cfg.DialTimeout
	}

	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = cfg.ReadTimeout
	}

	if cfg.BlockBufferSize > 0 {
		opts.BlockBufferSize = cfg.BlockBufferSize
	}

	if cfg.MaxCompressionBuffer > 0 {
		opts.MaxCompressionBuffer = cfg.MaxCompressionBuffer
	}

	if cfg.Settings != nil {
		opts.Settings = cfg.Settings
	}

	// tls

	if cfg.InsecureSkipVerify {
		opts.TLS = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // skip
	}

	return opts
}

const (
	ConnOpenInOrder    = "in_order"
	ConnOpenRoundRobin = "round_robin"
	ConnOpenRandom     = "random"
)

func getConnOpenStrategy(strategy string) clickhouse.ConnOpenStrategy {
	switch strategy {
	case ConnOpenInOrder:
		return clickhouse.ConnOpenInOrder
	case ConnOpenRoundRobin:
		return clickhouse.ConnOpenRoundRobin
	case ConnOpenRandom:
		return clickhouse.ConnOpenRandom
	default:
		return clickhouse.ConnOpenInOrder
	}
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mkbeh/xclick"
	"github.com/mkbeh/xclick/examples/sample/migrations"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var pool *clickhouse.Pool

var (
	host string
	user string
	pass string
	db   string
	args string
)

type Task struct {
	ID          int64     `ch:"id"`
	Description string    `ch:"description"`
	CreatedAt   time.Time `ch:"created_at"`
}

func init() {
	host = os.Getenv("CLICKHOUSE_HOSTS")
	user = os.Getenv("CLICKHOUSE_USER")
	pass = os.Getenv("CLICKHOUSE_PASSWORD")
	db = os.Getenv("CLICKHOUSE_DB")
	args = os.Getenv("CLICKHOUSE_MIGRATE_ARGS")
}

func getTasksHandler(w http.ResponseWriter, req *http.Request) {
	var resp Task

	query := `SELECT id, created_at, description FROM tasks LIMIT 1`

	err := pool.QueryRow(req.Context(), query).Scan(&resp.ID, &resp.CreatedAt, &resp.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func createTasksHandler(w http.ResponseWriter, req *http.Request) {
	query := `INSERT INTO tasks`
	batch, err := pool.PrepareBatch(req.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := batch.AppendStruct(&Task{ID: 1, Description: "test", CreatedAt: time.Now()}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := batch.Send(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func main() {
	var err error

	cfg := &clickhouse.Config{
		Hosts:          host,
		User:           user,
		Password:       pass,
		DB:             db,
		MigrateEnabled: true,
		MigrateArgs:    args,
	}

	pool, err = clickhouse.NewPool(
		clickhouse.WithConfig(cfg),
		clickhouse.WithClientID("test-client"),
		clickhouse.WithMigrations(migrations.FS),
	)
	if err != nil {
		log.Fatalln(err)
	}

	defer pool.Close()

	http.HandleFunc("/get", getTasksHandler)
	http.HandleFunc("/create", createTasksHandler)
	http.Handle("/metrics", promhttp.Handler())

	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatalln("Unable to start web server:", err)
	}
}

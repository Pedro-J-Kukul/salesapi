package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
)

const version = "v0.3.0"

// Server configuration settings
type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	cors struct {
		trustedOrigins []string
	}
	rateLimit struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type app struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	cfg := loadConfig()

	logger := setupLogger(cfg)

	db, err := openDB(cfg)
	if err != nil {
		logger.Error("Error opening database connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database connection pool established")

	app := &app{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.Error("Error starting server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func loadConfig() config {
	var cfg config

	// Parsing environment variables
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL database connection string")
	flag.Float64Var(&cfg.rateLimit.rps, "rate-limit-rps", 5, "Requests per second")
	flag.IntVar(&cfg.rateLimit.burst, "rate-limit-burst", 10, "Burst limit")
	flag.BoolVar(&cfg.rateLimit.enabled, "rate-limit-enabled", false, "Enable rate limiting")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = append(cfg.cors.trustedOrigins, val)
		return nil
	})
	flag.Parse()

	return cfg
}

func setupLogger(cfg config) *slog.Logger {
	var logger *slog.Logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	// Output loaded configuration settings
	logger.Info("Starting server",
		slog.String("version", version),
		slog.String("env", cfg.env),
		slog.Int("port", cfg.port),
		slog.Float64("rateLimitRPS", cfg.rateLimit.rps),
		slog.Int("rateLimitBurst", cfg.rateLimit.burst),
		slog.Bool("rateLimitEnabled", cfg.rateLimit.enabled),
	)

	return logger
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

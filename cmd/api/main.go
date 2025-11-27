package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/mailer"
)

const version = "v0.3.0"

// Server configuration settings
type config struct {
	port int    // server port
	env  string // environment (development, staging, production)
	db   struct {
		dsn          string        // database source name
		maxOpenConns int           // maximum number of open connections
		maxIdleConns int           // maximum number of idle connections
		maxIdleTime  time.Duration // maximum idle time for connections
	}
	cors struct {
		trustedOrigins []string // list of trusted CORS origins
	}
	limiter struct {
		rps     float64 // requests per second
		burst   int     // burst size
		enabled bool    // whether the limiter is enabled
	}
	smtp struct {
		host     string // SMTP host
		port     int    // SMTP port
		username string // SMTP username
		password string // SMTP password
		sender   string // SMTP sender address
	}
}

type app struct {
	config config         // application configuration settings
	logger *slog.Logger   // logger for structured logging
	wg     sync.WaitGroup // wait group for managing goroutines
	models data.Models
	mailer *mailer.Mailer
}

func main() {
	// For application setup
	cfg := loadConfig()            // load the application configuration
	logger := setUpLogger(cfg.env) // set up the logger
	db, err := openDB(cfg)         // open the database connection
	if err != nil {
		logger.Error("unable to connect to database", slog.Any("error", err)) // log any error connecting to the database
		os.Exit(1)                                                            // exit if there is a database connection error
	}
	defer db.Close()                                    // ensure the database connection is closed when main() exits
	logger.Info("database connection pool established") // log successful database connection

	// For metrics
	expvar.NewString("version").Set(version) // publish the application version
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine() // publish the number of active goroutines
	}))
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats() // publish database connection pool statistics
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix() // publish the current Unix timestamp
	}))

	// Initialize the application dependencies
	app := &app{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	if cfg.smtp.host != "" && cfg.smtp.sender != "" {
		app.mailer = mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)
	}

	err = app.serve() // start the HTTP server
	if err != nil {
		logger.Error("error starting server", slog.Any("error", err)) // log any error starting the server
		os.Exit(1)                                                    // exit if there is a server error
	}
}

func loadConfig() config {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")                                        // server port
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)") // environment

	// Database settings
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")                                                   // database source name
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")                 // max open connections
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")                 // max idle connections
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", time.Minute, "PostgreSQL max connection idle time") // max idle time

	// CORS settings
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s) // split the input string by spaces and assign to trustedOrigins
		return nil
	})

	// Rate limiter settings
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second") // requests per second
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")               // burst size
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")              // whether the limiter is enabled

	// SMTP settings
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")                             // SMTP host
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")                                              // SMTP port
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP username")                                 // SMTP username
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")                                 // SMTP password
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Training <noreply@example.com>", "SMTP sender address") // SMTP sender address

	flag.Parse() // parse the command-line flags

	// Print out all the flag values for debugging
	if cfg.env == "development" {
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("Flag %s: %v\n", f.Name, f.Value)
		})
	}

	// Regex cfg.smtp.sender and convert the first # to < and last # to >
	re := regexp.MustCompile(`#(.*?)#`)
	cfg.smtp.sender = re.ReplaceAllString(cfg.smtp.sender, "<$1>")

	if cfg.db.dsn == "" {
		cfg.db.dsn = os.Getenv("DB_DSN")
	}
	if cfg.db.dsn == "" {
		panic("db-dsn must be provided via flag or DB_DSN environment variable")
	}

	if len(cfg.cors.trustedOrigins) == 0 {
		if origins := strings.Fields(os.Getenv("CORS_TRUSTED_ORIGINS")); len(origins) > 0 {
			cfg.cors.trustedOrigins = origins
		}
	}

	if cfg.smtp.host == "" {
		cfg.smtp.host = os.Getenv("SMTP_HOST")
	}
	if cfg.smtp.username == "" {
		cfg.smtp.username = os.Getenv("SMTP_USERNAME")
	}
	if cfg.smtp.password == "" {
		cfg.smtp.password = os.Getenv("SMTP_PASSWORD")
	}
	if cfg.smtp.sender == "Training <noreply@example.com>" {
		if sender := os.Getenv("SMTP_SENDER"); sender != "" {
			cfg.smtp.sender = sender
		}
	}

	return cfg // return the populated configuration
}

func setUpLogger(env string) *slog.Logger {
	var logger *slog.Logger                                  // declare a logger variable
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))   // default to text handler
	logger = logger.With("app_version", version, "env", env) // add default fields to the logger
	return logger                                            // return the configured logger
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

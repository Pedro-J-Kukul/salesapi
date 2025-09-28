package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Function to start the HTTP server
func (app *app) serve() error {

	// Server configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error) // channel to receive shutdown errors
	// Goroutine to handle graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)                      // channel to receive OS signals
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // listen for interrupt and terminate signals
		s := <-quit                                          // block until a signal is received
		app.logger.Info("shutting down server", "signal", s) // log the shutdown signal

		// 30 second timeout for the shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()                                   // run until the timeout expires
		app.logger.Error("attempting graceful shutdown") // log the shutdown attempt
		shutdownError <- srv.Shutdown(ctx)               // attempt to gracefully shutdown the server
	}()

	// log the server start
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Start the server
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		app.logger.Error("server error", "error", err) // log any errors starting the server
		return err
	}
	// Wait for the shutdown goroutine to complete
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("server stopped", "addr", srv.Addr) // log the server stop

	return nil
}

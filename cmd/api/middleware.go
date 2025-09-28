package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Middleware to close connection when an unexpected panic occurs
func (a *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use defer to ensure this runs at the end of the function
		defer func() {
			err := recover() // recover from panic, returns nil if no panic occurred
			if err != nil {
				w.Header().Set("Connection", "close")              // close the connection
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err)) // log the error and send a 500 response
			}
		}()
		next.ServeHTTP(w, r) // call the next handler in the chain
	})
}

// Middleware for enabling cors
func (a *app) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Setting default CORS headers
		w.Header().Add("Vary", "Origin")                          // Indicate that the response varies based on the Origin header
		w.Header().Add("Vary", "Access-Control-Requested-Method") // Indicate that the response varies based on the Access-Control-Request-Method header

		origin := r.Header.Get("Origin") // Get the Origin header from the request

		// If the origin is in the trusted list, set the appropriate CORS headers
		if origin != "" {
			// Iterate over the trusted origins
			for i := range a.config.cors.trustedOrigins {
				// If the origin matches a trusted origin, set the Access-Control-Allow-Origin header
				if origin == a.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin) // Set the allowed origin to the request's origin

					// If the request method is OPTIONS and has Access-Control-Request-Method header, it's a preflight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE") // Set the Methods allowed for CORS
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type") // Set the Headers allowed for CORS
						w.WriteHeader(http.StatusOK)                                                  // Respond with 200 OK for preflight request
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r) // Call the next handler in the chain
	})
}

// Middleware to limit the rate of requests a client can make
func (a *app) rateLimit(next http.Handler) http.Handler {
	// Define a struct to hold the client information
	type client struct {
		limiter  *rate.Limiter // Rate limiter for the client
		lastSeen time.Time     // Last seen time for the client
	}

	var mu sync.Mutex                   // Mutex to protect the clients map
	clients := make(map[string]*client) // Map to hold the clients

	// Launch a goroutine to clean up old clients every minute
	go func() {
		for {
			time.Sleep(time.Minute) // Run every minute
			mu.Lock()               // Lock the mutex

			// Iterate over the clients map and delete old clients
			for ip, client := range clients {
				// If the client hasn't been seen for more than 3 minutes, delete it
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock() // Unlock the mutex after cleaning up
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.config.rateLimit.enabled {

			// Get the client's IP address returns an error if it fails
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}
			mu.Lock()               // Lock the mutex to safely access the clients map
			_, found := clients[ip] // Check if the client already exists

			// Create a new rate limiter for the client if it doesn't exist
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(a.config.rateLimit.rps), a.config.rateLimit.burst)}
			}
			clients[ip].lastSeen = time.Now() // Update the last seen time

			// Check if the request is allowed by the rate limiter
			if !clients[ip].limiter.Allow() {
				mu.Unlock() // Unlock the mutex before returning
				a.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock() // Unlock the mutex so other requests can access the clients map
		}
		next.ServeHTTP(w, r) // Call the next handler in the chain
	})
}

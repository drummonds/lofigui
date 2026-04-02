package lofigui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// ListenAndServe starts an HTTP server on the given address.
// If the port is in use, it tries up to 10 consecutive ports.
// The address must be in the form ":port" or "host:port".
//
// When handler is nil (using DefaultServeMux), a default /favicon.ico
// handler is registered automatically unless one is already registered.
func ListenAndServe(addr string, handler http.Handler) error {
	if handler == nil {
		registerDefaultFavicon()
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address %q: %w", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	for i := 0; i < 10; i++ {
		tryAddr := net.JoinHostPort(host, strconv.Itoa(port+i))
		ln, err := net.Listen("tcp", tryAddr)
		if err != nil {
			if isAddrInUse(err) {
				log.Printf("Port %d in use, trying %d", port+i, port+i+1)
				continue
			}
			return err
		}
		log.Printf("Starting server on http://localhost:%d", port+i)
		return http.Serve(ln, handler)
	}
	return fmt.Errorf("all ports %d-%d in use", port, port+9)
}

func isAddrInUse(err error) bool {
	return strings.Contains(err.Error(), "address already in use")
}

// ListenAndServe starts an HTTP server on the given address, with graceful
// shutdown when the app's model completes (via Handle).
//
// Use this instead of the package-level ListenAndServe when using app.Handle,
// so the server exits automatically when the model finishes.
func (app *App) ListenAndServe(addr string, handler http.Handler) error {
	if handler == nil {
		registerDefaultFavicon()
	}

	app.mu.Lock()
	app.done = make(chan struct{})
	app.mu.Unlock()

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address %q: %w", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	for i := 0; i < 10; i++ {
		tryAddr := net.JoinHostPort(host, strconv.Itoa(port+i))
		ln, err := net.Listen("tcp", tryAddr)
		if err != nil {
			if isAddrInUse(err) {
				log.Printf("Port %d in use, trying %d", port+i, port+i+1)
				continue
			}
			return err
		}
		log.Printf("Starting server on http://localhost:%d", port+i)

		srv := &http.Server{Handler: handler}
		app.mu.Lock()
		app.server = srv
		app.mu.Unlock()

		go func() {
			<-app.done
			srv.Shutdown(context.Background())
		}()

		err = srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			app.mu.RLock()
			cancelled := app.cancelled
			app.mu.RUnlock()
			if cancelled {
				return ErrCancelled
			}
			return nil // graceful shutdown
		}
		return err
	}
	return fmt.Errorf("all ports %d-%d in use", port, port+9)
}

// signalDone closes the done channel to trigger graceful shutdown.
// No-op if done channel is nil (using package-level ListenAndServe).
func (app *App) signalDone() {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.done != nil {
		select {
		case <-app.done:
		default:
			close(app.done)
		}
	}
}

// registerDefaultFavicon registers ServeFavicon on DefaultServeMux
// if /favicon.ico is not already registered. Safe to call multiple times.
func registerDefaultFavicon() {
	defer func() { recover() }() // HandleFunc panics on duplicate pattern
	http.HandleFunc("/favicon.ico", ServeFavicon)
}

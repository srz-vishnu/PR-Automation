package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	// DefaultReadHeaderTimeOut is to read deadline just after connection is accepted
	// The default is 5 seconds
	DefaultReadHeaderTimeOut = 5 * time.Second

	// DefaultReadTimeOut is to read deadline just after connection is accepted
	// The default is 120 seconds
	DefaultReadTimeOut = 120 * time.Second

	// DefaultWriteTimeOut is the write deadline,
	// default is 0 second
	DefaultWriteTimeOut = 0 * time.Second

	// DefaultIdleTimeOut is the write deadline,
	DefaultIdleTimeOut = 60 * time.Second

	// Shutdown timeout
	ShutdownTimeout = 5 * time.Second
)

func Start(r chi.Router) {

	server := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: DefaultReadHeaderTimeOut,
		ReadTimeout:       DefaultReadTimeOut,
		WriteTimeout:      DefaultWriteTimeOut,
		IdleTimeout:       DefaultIdleTimeOut,
		Handler:           r,
	}
	StartHTTPServer(&server)

}
func StartHTTPServer(s *http.Server) {
	shutdownComplete := make(chan struct{})

	// handle SIGINT and perform graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
			os.Exit(1)
		}
		close(shutdownComplete)
	}()

	log.Printf("HTTP server listening to %v", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-shutdownComplete
}

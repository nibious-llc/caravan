package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/nibious-llc/caravan/internal/server"
	s "github.com/nibious-llc/caravan/internal/server"
	"github.com/rs/zerolog/log"
)

// Global variables that are necessary for all methods to access
var machines_signed_in map[uuid.UUID](bool)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  256,
	WriteBufferSize: 256,
	WriteBufferPool: &sync.Pool{},
}

func process(con *websocket.Conn) {
	// Build management structure
	hub := c.NewClientHub()

	hub.MetricsPort = c.HandleMetrics(hub)

	// Start threads to process incoming/outgoing data, including error handling
	go c.Error(con, hub)
	go c.Reader(con, hub)
	go c.Writer(con, hub)

	// Start processing the messages themselves
	go s.HandleClient(hub)

	// Start metrics server
}

func handle_connection(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("[handler][upgrade]")
		return
	}

	// Process connection in a new goroutine
	go process(c)
}

func main() {

	log.Info().Msg("[Nibious Caravan Server] Starting up...")

	server.ConnectToK8s()

	log.Info().Msg("[Nibious Caravan Server] ✅ Ready")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	log.Info().Msg(fmt.Sprintf("Serving at localhost:%s...", port))

	var DefaultServeMux = http.NewServeMux()

	DefaultServeMux.HandleFunc("/ws/", handle_connection)

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		log.Info().Msg("Waiting for signal to quit...")

		<-sigint

		// We received an interrupt signal, shut down.
		close(idleConnsClosed)
	}()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: DefaultServeMux,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		},
	}

	//go func () {
	//	if err := srv.ListenAndServeTLS("/certs/tls.crt", "/certs/tls.key"); err != http.ErrServerClosed {
	//		// Error starting or closing listener:
	//		log.Error().Err(err).Msg("HTTP server ListenAndServe")
	//	}
	//}()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Error().Err(err).Msg("HTTP server ListenAndServe")
		}
	}()

	<-idleConnsClosed

}

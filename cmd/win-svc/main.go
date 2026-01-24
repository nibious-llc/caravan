package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/gorilla/websocket"
	"github.com/nibious-llc/caravan/internal/client"
	"github.com/nibious-llc/caravan/internal/common"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

// EventLogHook implements the zerolog.Hook interface
type EventLogHook struct {
	// Add fields for your event log client here
	// eventLogClient *eventlog.Log
}

func (h *EventLogHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// Extract data from the event e if needed
	// You can also use reflection on the event's internal buffer (less performant)

	// Call the specific API of your event log system
	// switch level {
	// case zerolog.InfoLevel:
	//     h.eventLogClient.Info(1, msg) // Example call
	// case zerolog.ErrorLevel:
	//     h.eventLogClient.Error(2, msg) // Example call
	// }
}

func main() {

	// Register with Windows Event Log
	source := "Nibious-Caravan"
	elog, err := eventlog.Open(source)
	if err != nil {
		// Try to install the source if it doesn't exist (requires elevated privileges)
		err = eventlog.InstallAsEventCreate(source, eventlog.Info|eventlog.Warning|eventlog.Error)
		if err != nil {
			log.Error().Err(err).Msg("Could not open or install event log source")
		}
		elog, err = eventlog.Open(source)
		if err != nil {
			log.Error().Err(err).Msg("Could not open event log source")
		}
	}
	defer elog.Close()

	// Now redirect all log messages to windows event log
	// Initialize the hook
	logHook := &EventLogHook{}

	// Create a new logger and register the hook
	log.Logger = zerolog.New(nil).Hook(logHook) // nil writer as the hook handles the output

	log.Print("[Nibious-Iunctio/Remote Access Client] Starting up...")

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "443"
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		panic("ADDRESS must be set in the environment")
	}

	scheme := os.Getenv("SCHEME")
	if scheme == "" {
		scheme = "wss"
	}

	uri := fmt.Sprintf("%s:%s", address, port)

	u := url.URL{Scheme: scheme, Host: uri, Path: "/ws/"}
	log.Info().Msg(fmt.Sprintf("connecting to %s", u.String()))
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error().Err(err).Msg("dial error")
		return
	}
	defer c.Close()

	// Build management structure
	hub := common.NewClientHub()

	// Start threads to process incoming/outgoing data, including error handling
	go common.Error(c, hub)
	go common.Reader(c, hub)
	go common.Writer(c, hub)

	// Start processing the messages themselves
	go client.HandleConnection(hub)

	signal.Notify(hub.CloseChan, syscall.SIGTERM, syscall.SIGINT)
	sig := <-hub.CloseChan
	log.Printf("Caught signal %v", sig)

}

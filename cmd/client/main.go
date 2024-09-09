package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/gorilla/websocket"
	"github.com/nibious-llc/caravan/internal/client"
	"github.com/nibious-llc/caravan/internal/common"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/rs/zerolog/log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func main() {

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

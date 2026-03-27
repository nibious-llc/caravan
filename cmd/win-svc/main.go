package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nibious-llc/caravan/internal/client"
	"github.com/nibious-llc/caravan/internal/common"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"net/url"
	"os"
)

const (
	serviceName = "Nibious-Caravan"
)

// EventLogHook implements the zerolog.Hook interface
type EventLogHook struct {
	// Add fields for your event log client here
	WinLog *eventlog.Log
}

func (h *EventLogHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	switch level {
	case zerolog.InfoLevel:
		h.WinLog.Info(3, msg)
	case zerolog.WarnLevel:
		h.WinLog.Warning(2, msg)
	case zerolog.ErrorLevel:
		h.WinLog.Error(1, msg)
	default:
		h.WinLog.Info(4, msg)
	}
}

type CaravanSVC struct{}

func main() {

	isAnInteractiveSession, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Error().Err(err).Msg("Failed to determine if interactive session")
		return
	}

	if isAnInteractiveSession {
		// It was requested to run as a CLI
		log.Error().Msg("This is only designed to be a windows service. Please use another binary to run as an interactive client")
		return
	}

	// Register with Windows Event Log
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		// Try to install the source if it doesn't exist (requires elevated privileges)
		err = eventlog.InstallAsEventCreate(serviceName, eventlog.Info|eventlog.Warning|eventlog.Error)
		if err != nil {
			log.Error().Err(err).Msg("Could not open or install event log source")
		}
		elog, err = eventlog.Open(serviceName)
		if err != nil {
			log.Error().Err(err).Msg("Could not open event log source")
		}
	}
	defer elog.Close()

	// Now redirect all log messages to windows event log
	// Initialize the hook
	logHook := &EventLogHook{WinLog: elog}

	// Create a new logger and register the hook
	log.Logger = zerolog.New(nil).Hook(logHook) // nil writer as the hook handles the output

	err = svc.Run(serviceName, &CaravanSVC{})
	if err != nil {
		log.Printf("Service failed: %v\n", serviceName, err)
	}
	log.Printf("Service %s stopped\n", serviceName)

}
func (m *CaravanSVC) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {

	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// Check that the environment is setup correctly
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
	hub, c := runServer(u)

	if hub == nil {
		log.Fatal().Msg("Could not start hub")
		return
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case changeReq := <-r:
			switch changeReq.Cmd {
			case svc.Interrogate:
				changes <- changeReq.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Info().Msg("Shutdown requested by service manager")
				goto EXIT
			default:
				log.Printf("Unexpected service control request #%d", changeReq.Cmd)
			}
		case <-hub.CloseChan:
			log.Info().Msg("Recieved closed channel signal")
			goto EXIT
		}
	}

EXIT:
	//Report that things are shutting down
	log.Info().Msg("Shutting down...")
	changes <- svc.Status{State: svc.StopPending}

	client.StopClient(hub)
	c.Close()

	return
}

func runServer(u url.URL) (*common.ClientHub, *websocket.Conn) {
	log.Print("Starting server...")

	log.Info().Msg(fmt.Sprintf("connecting to %s", u.String()))
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error().Err(err).Msg("dial error")
		return nil, nil
	}

	// Build management structure
	hub := common.NewClientHub()

	// Start threads to process incoming/outgoing data, including error handling
	go common.Error(c, hub)
	go common.Reader(c, hub)
	go common.Writer(c, hub)

	// Start processing the messages themselves
	go client.HandleConnection(hub)

	return hub, c

}

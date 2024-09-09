package server

import (
	"fmt"
	"github.com/google/uuid"
	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
)

// Setup constants for the reverse proxy ports
const (
	CONN_HOST = "0.0.0.0" // available on any IP address
	CONN_PORT = "0"       // 0 means random available port
	CONN_TYPE = "tcp"     // We are only focused on TCP connections
)

func setup_tunnel_listeners(client *c.ClientHub, tunnels []c.TunnelInit) bool {

	for _, t := range tunnels {
		var tunnel c.Tunnel
		// Create a listener for each tunnel request
		l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		if err != nil {
			log.Error().Err(err).Msg("Error listening:")
			return false
		}
		tunnel.Listener = l
		tunnel.ListenerPort = tunnel.Listener.Addr().(*net.TCPAddr).Port
		tunnel.IPAddr = t.IPAddr
		tunnel.Port = t.Port
		tunnel.Control = make(chan int)
		tunnel.SessionID = strconv.Itoa(t.Port)
		client.Tunnels = append(client.Tunnels, tunnel)
	}
	return true
}

func start_tunnel_listener_threads(client *c.ClientHub) {

	// loop over the tunnels and wait to accept. Need to start new threads.
	log.Debug().Msg("Starting tunnel listeners")
	for _, tunnel := range client.Tunnels {
		go func(tunnel c.Tunnel) {

			// TOOD: Add switch here for control channel.

			// Close the listener when the method closes.
			defer tunnel.Listener.Close()
			log.Info().Msg(fmt.Sprintf("[start_tunnel_listener_threads] Listening on %s:%d", CONN_HOST, tunnel.ListenerPort))
			for {

				// Listen for an incoming connection.
				conn, err := tunnel.Listener.Accept()
				log.Info().Msg("[start_tunnel_listener_threads] New connection")
				if err != nil {
					log.Error().Err(err).Msg("[start_tunnel_listener_threads] Error accepting connection")
					break
				}

				// Handle connections in a new goroutine.
				// New Tunnel Instance,
				tunnel_instance := c.NewIunctioActiveTunnel(conn)

				go handle_tunnel_request(client, tunnel, tunnel_instance)
			}
			log.Info().Msg(fmt.Sprintf("[start_tunnel_listener_threads] Closing socket on %s:%d", CONN_HOST, tunnel.ListenerPort))
		}(tunnel)
	}
}

func generateTunnelStartMsg(tunnel c.Tunnel, tunnel_instance *c.IunctioActiveTunnel) []byte {

	var tunnelStart c.TunnelStart
	tunnelStart.IPAddr = tunnel.IPAddr
	tunnelStart.Port = tunnel.Port
	tunnelStart.Id = tunnel_instance.Id

	data_to_send, err := c.MarshalObject(tunnelStart)
	if err != nil {
		log.Error().Err(err)
		return nil
	}

	return c.ConvertToMessage(c.RequestTunnelInitMsgType, data_to_send)
}

// Handles incoming requests.
func handle_tunnel_request(client *c.ClientHub, tunnel c.Tunnel, tunnel_instance *c.IunctioActiveTunnel) {

	// We need to setup a new channel per request... So we can send data to the
	// right request

	tunnel_instance.Id = uuid.New()
	log.Info().Msg(fmt.Sprintf("[handle_tunnel_request] Creating Tunnel Request: %s", tunnel_instance.Id))

	// Add to the client notes
	client.ActiveConns[tunnel_instance.Id] = *tunnel_instance

	// Ask the other end if we are all good to start a tunnel
	client.Writer <- generateTunnelStartMsg(tunnel, tunnel_instance)
}

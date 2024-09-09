package server

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/prometheus/client_model/go"
	"github.com/rs/zerolog/log"
	"time"
)

type IAuthProvider interface {
	IsLoginValid(clientID uuid.UUID, secret string) (c.IunctioClient, bool)
}

var machines_signed_in map[uuid.UUID](*c.ClientHub)
var AuthProvider IAuthProvider

func HandleClient(client *c.ClientHub) {
	// Start heartbeat ticker
	ticker := time.NewTicker(time.Second * 90)
	defer ticker.Stop()
	defer closeClient(client)

	// First ask for authentication from client
	authenticated := false
	client.Writer <- generateRequestAuthMsg()
	alive_counter := 2

	for {
		select {
		case incoming := <-client.Reader:
			var msg c.Message
			err := c.UnmarshalObject(incoming, &msg)
			if err != nil {
				log.Error().Err(err).Msg("Could not unmarshal object")
			}
			// determine the kind of message from the client and act upon it
			switch msg.Type {

			case c.ResponseCredentialsMsgType:
				log.Debug().Msg("Received credentials from client")

				if verifyCreds(msg.Content, client) == false {
					return
				}

				authenticated = true

				client.Writer <- generateRequestAuthSuccessMsg()

				// Ask for tunnel details
				client.Writer <- generateTunnelDescriptionRequest()

			case c.ResponseTunnelDescriptionsMsgType:
				// Setup the tunnels that have been requested
				if !authenticated {
					log.Error().Msg("Client is not authenticated")
					return
				}

				log.Debug().Msg("Setting up tunnels...")
				setupTunnels(client, msg.Content)

			case c.TunnelMessageType:
				if !authenticated {
					return
				}
				log.Info().Msg("Tunnel Message Received")

			case c.AdminMessageType:
				if !authenticated {
					return
				}

				log.Info().Msg("Admin Message Received")

			case c.ResponseTunnelInitMsgType:
				if !authenticated {
					return
				}

				log.Info().Msg("Tunnel Request Accepted. Starting handler")
				handleTunnel(client, msg.Content)

			case c.TunnelDataMsgType:
				if !authenticated {
					return
				}
				c.HandleDataReceive(client, msg.Content)

			case c.ClientGoodbyeMessageType:
				log.Debug().Msg("Client said goodbye")
				return

			case c.RequestTunnelCloseMsgType:
				if !authenticated {
					return
				}
				log.Debug().Msg("Tunnel needs to be closed")
				c.CloseTunnel(client, msg.Content)

			case c.ResponseMetricsReportMsgType:
				handleMetricsReport(client, msg.Content)

			case c.PingDataMsgType:
				client.Writer <- c.GeneratePong()
				alive_counter++
			}

			// Auth message

			// Data message -> pass along to the proper queue for this client.
		case <-ticker.C:
			if alive_counter < 0 {
				log.Error().Msg(fmt.Sprintf("Client (%s) did not respond within keepalive interval. Closing Connection", client.LoginID))
				c.CloseTunnel(client, []byte{})
			}
			client.Writer <- c.GeneratePing()
			alive_counter--

		}
	}

}

func setupTunnels(client *c.ClientHub, content []byte) bool {

	var tunnels []c.TunnelInit
	err := c.UnmarshalObject(content, &tunnels)
	if err != nil {
		log.Error().Err(err).Msg("Could not unmarshal object")
		return false
	}

	// Setup normal tunnels
	if !setup_tunnel_listeners(client, tunnels) {
		log.Error().Msg("Could not setup tunnels")
		return false
	}

	start_tunnel_listener_threads(client)

	// Create services based on the tunnels
	create_service(client)
	return true
}

func closeClient(client *c.ClientHub) {

	if client.MetricsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		client.MetricsServer.Shutdown(ctx)
	}

	delete_service(client)

	for _, tunnel := range client.ActiveConns {
		log.Info().Msg("Closing tunnels")
		tunnel.Control <- c.ConvertToMessage(c.RequestTunnelCloseMsgType, []byte("o"))
	}

	for _, tunnel := range client.Tunnels {
		log.Info().Msg("Removing tunnels")
		tunnel.Listener.Close()
	}

	close(client.Reader)
	close(client.Writer)
	close(client.Control)

	log.Info().Msg("Client has been closed")

}

func generateRequestAuthMsg() []byte {

	msg := c.Message{
		Type:    c.RequestAuthenticationMsgType,
		Content: nil,
	}

	d, err := c.MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d

}

func verifyCreds(msg []byte, client *c.ClientHub) bool {

	var login_data c.IunctioClientLogin
	err := c.UnmarshalObject(msg, &login_data)
	if err != nil {
		return false
	}

	// Check the login
	db_login_data, ok := AuthProvider.IsLoginValid(login_data.ClientID, login_data.Secret)

	if !ok {
		log.Info().Msg("Invalid login: " + login_data.ClientID.String())
		return false
	}

	log.Info().Msg(fmt.Sprintf("[OnEvent][login] New client logged in: %s (%s)", db_login_data.ClientID, db_login_data.Hostname))

	client.LoginID = db_login_data.ClientID
	client.Hostname = db_login_data.Hostname
	client.Namespace = db_login_data.Namespace

	return true
}

func generateRequestAuthSuccessMsg() []byte {

	msg := c.Message{
		Type:    c.ResponseAuthenticationSuccessMsgType,
		Content: nil,
	}

	d, err := c.MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d

}

func generateTunnelDescriptionRequest() []byte {

	msg := c.Message{
		Type:    c.RequestTunnelDescriptionsMsgType,
		Content: nil,
	}

	d, err := c.MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d

}

// Server side, handle the tunnel
func handleTunnel(client *c.ClientHub, msg []byte) {

	var tunnel c.TunnelStart
	err := c.UnmarshalObject(msg, &tunnel)
	if err != nil {
		log.Error().Err(err).Msg("[handleTunnel] Could not unmarshal object")
		return
	}

	// Start threads to process incoming/outgoing data, including error handling
	go c.TunnelError(client, tunnel.Id)
	go c.TunnelReader(client, tunnel.Id)
	go c.TunnelWriter(client, tunnel.Id)

}

func handleMetricsReport(client *c.ClientHub, msg []byte) {
	var data []*io_prometheus_client.MetricFamily
	err := c.UnmarshalObject(msg, &data)
	if err != nil {
		log.Error().Err(err).Msg("[handleTunnel] Could not unmarshal object")
		return
	}

	client.MetricsChan <- data
}

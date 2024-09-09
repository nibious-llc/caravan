package client

import (
	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/rs/zerolog/log"
	"syscall"
	"time"
)

func StopClient(client *c.ClientHub) {

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

	client.CloseChan <- syscall.SIGTERM

	log.Info().Msg("Client has been closed")

}

func HandleConnection(client *c.ClientHub) {

	// Start heartbeat ticker
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	defer StopClient(client)

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

			case c.TunnelMessageType:
				log.Info().Msg("Tunnel Message Received")

			case c.AdminMessageType:
				log.Info().Msg("Admin Message Received")

			case c.RequestAuthenticationMsgType:
				// Send Auth details
				log.Debug().Msg("Sending Login Details")
				client.Writer <- generateLoginCredsMsg()

			case c.ResponseAuthenticationSuccessMsgType:
				log.Debug().Msg("Auth successful")

				log.Debug().Msg("Starting metrics")
				initMetricsReport()

			case c.PingDataMsgType:
				client.Writer <- c.GeneratePong()
				alive_counter++

			case c.RequestTunnelDescriptionsMsgType:
				log.Debug().Msg("Tunnel Descriptions Requested")
				client.Writer <- generateTunnelDescriptions()

			case c.RequestTunnelInitMsgType:
				log.Debug().Msg("Tunnel Requested")
				client.Writer <- generateTunnelInitAck(client, msg.Content)

			case c.TunnelDataMsgType:
				c.HandleDataReceive(client, msg.Content)

			case c.RequestTunnelCloseMsgType:
				c.CloseTunnel(client, msg.Content)

			case c.RequestMetricsMsgType:
				client.Writer <- generateMetricsReportMsg()

			case c.ClientGoodbyeMessageType:
				log.Debug().Msg("Server said goodbye")
				return
			}
		case <-ticker.C:
			if alive_counter < 0 {
				log.Error().Msg("Server did not respond within keepalive interval. Quitting...")
				return
			}
			client.Writer <- c.GeneratePing()
			alive_counter--
		}
	}
}

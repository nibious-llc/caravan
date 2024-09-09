package client

import (
	"fmt"
	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/rs/zerolog/log"
	"net"
	"net/netip"
)

// This is for the client to send login credentials
func generateTunnelDescriptions() []byte {

	// Send information for tunnelling
	var tunnels []c.TunnelInit

	tunnels = append(tunnels, c.TunnelInit{netip.MustParseAddr("127.0.0.1"), 22})

	obj, err := c.MarshalObject(tunnels)
	if err != nil {
		log.Err(err).Msg("Could not marshal []tunnel object")
		return nil
	}

	return c.ConvertToMessage(c.ResponseTunnelDescriptionsMsgType, obj)
}

// This is client side -- we need to setup a TCP Dial object and check it
func generateTunnelInitAck(client *c.ClientHub, content []byte) []byte {
	// Default to yes, start the tunnel!

	// Need to connect to the service. If successful, then we can send back a
	// message stating that success.

	var tunnelStart c.TunnelStart
	err := c.UnmarshalObject(content, &tunnelStart)
	if err != nil {
		log.Error().Err(err).Msg("[generateTunnelInitAck] Could not unmarshal object")

		return c.ConvertToMessage(c.RequestTunnelCloseMsgType, content)
	}

	// connect to the specified tunnel
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tunnelStart.IPAddr.String(), tunnelStart.Port))
	if err != nil {
		log.Error().Err(err).Msg("[OnEvent:setup_tunnel] Connection to service error")
		log.Debug().Msg(fmt.Sprintf("%s", tunnelStart.Id))
		return c.GenerateCloseTunnelMsg(tunnelStart.Id)
	}

	// Add active connection to client structure
	tunnel_instance := c.NewIunctioActiveTunnel(conn)
	client.ActiveConns[tunnelStart.Id] = *tunnel_instance

	// Go start reading/writing/error
	go c.TunnelError(client, tunnelStart.Id)
	go c.TunnelReader(client, tunnelStart.Id)
	go c.TunnelWriter(client, tunnelStart.Id)

	return c.ConvertToMessage(c.ResponseTunnelInitMsgType, content)
}

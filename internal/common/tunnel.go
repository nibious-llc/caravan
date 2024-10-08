package common

import (
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func CloseTunnel(client *ClientHub, content []byte) {
	var tunnel TunnelData
	err := UnmarshalObject(content, &tunnel)
	if err != nil {
		log.Error().Err(err).Msg("[closeTunnel] Could not unmarshal object")
		return
	}
	conn, ok := client.ActiveConns[tunnel.Id]

	if ok {
		if !conn.ThreadsRunning {
			conn.Conn.Close()
		} else {
			conn.Control <- ConvertToMessage(RequestTunnelCloseMsgType, content)
		}
	}
}

func HandleDataReceive(client *ClientHub, msg []byte) {

	var data TunnelData
	err := UnmarshalObject(msg, &data)
	if err != nil {
		log.Error().Err(err).Msg("[handleTunnel] Could not unmarshal object")
		return
	}

	client.ActiveConns[data.Id].Reader <- msg

}

func TunnelError(client *ClientHub, connID uuid.UUID) {

	tunnel := client.ActiveConns[connID]
	tunnel.ThreadsRunning = true
	log.Info().Msg("Waiting for control")
	<-tunnel.Control
	log.Info().Msg("Error happening!")

	// Notify the other end that we are closing the connection
	client.Writer <- GenerateCloseTunnelMsg(connID)

	tunnel.Conn.Close()

	close(tunnel.Reader)

	log.Info().Msg("Waiting for reader")
	<-tunnel.Control //Reader

	log.Info().Msg("Closing the tunnel")
	close(tunnel.Control)

	delete(client.ActiveConns, connID)
	log.Debug().Msg("Tunnel has been deleted")
}

// Take information from the TCP connection specified by connID and send it to
// the client application via websockets. This is "reading" from the TCP
// connection and dropping the information into the channels
func TunnelReader(client *ClientHub, connID uuid.UUID) {

	defer func() {
		if r := recover(); r != nil {
			log.Info().Msg("TunnelReader recovered")
		}
	}()

	tunnel := client.ActiveConns[connID]
	conn := tunnel.Conn

	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		reqLen = reqLen
		if err != nil {
			if err.Error() != "EOF" {
				log.Error().Err(err).Msg("[Reader][Tunnel] Error happened")
			}
			log.Info().Msg("[Reader][Tunnel] Closing Connection")

			// Tell this side to close as well.
			tunnel.Control <- GenerateCloseTunnelMsg(connID)

			return
		}

		str := base64.StdEncoding.EncodeToString(buf[:reqLen])

		data := TunnelData{str, connID}

		data_to_send, err := MarshalObject(data)
		if err != nil {
			log.Error().Err(err)
			return
		}

		client.Writer <- ConvertToMessage(TunnelDataMsgType, data_to_send)
	}
}

func TunnelWriter(client *ClientHub, connID uuid.UUID) {

	tunnel := client.ActiveConns[connID]
	conn := tunnel.Conn

	for {
		content, ok := <-tunnel.Reader

		if !ok {
			tunnel.Control <- GenerateCloseTunnelMsg(connID)
			return
		}

		var data TunnelData
		err := UnmarshalObject(content, &data)
		if err != nil {
			log.Error().Err(err).Msg("[handleTunnel] Could not unmarshal object")
			tunnel.Control <- GenerateCloseTunnelMsg(connID)
			return
		}

		bytes_to_send, base64_err := base64.StdEncoding.DecodeString(data.Data)
		if base64_err != nil {
			log.Error().Err(base64_err).Msg("[Writer] Could not decode string")
			tunnel.Control <- GenerateCloseTunnelMsg(connID)
			return
		}

		_, err_write := conn.Write(bytes_to_send)
		if err_write != nil {
			log.Error().Err(err_write).Msg("[Writer]")
			tunnel.Control <- GenerateCloseTunnelMsg(connID)
			return
		}
	}
}

func GenerateCloseTunnelMsg(id uuid.UUID) []byte {
	var closeMsg TunnelData
	closeMsg.Id = id
	data_to_send, err := MarshalObject(closeMsg)
	if err != nil {
		log.Error().Err(err)
		return ConvertToMessage(RequestTunnelCloseMsgType, []byte("Really bad error..."))
	}
	return ConvertToMessage(RequestTunnelCloseMsgType, data_to_send)
}

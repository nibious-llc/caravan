package common

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_model/go"
	"github.com/rs/zerolog/log"
	//	"k8s.io/client-go/rest"
	"net/http"
	//	"nibious.com/iunctio/pkg/clientset/v1alpha1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"os"
)

func MarshalObject(o any) ([]byte, error) {

	data, err := json.Marshal(o)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func UnmarshalObject(data []byte, v any) error {
	return json.Unmarshal(data, &v)
}

// This is a struct to hold information about the client that is connected to
// the server. This should have information about the client connection itself
// and any tunnels/connections that are available. This allows the server to
// safely discard resources in the case of a client disconnect.
type ClientHub struct {
	Reader  chan []byte
	Writer  chan []byte
	Control chan int

	Tunnels     []Tunnel                            // Tunnels the client is sending to the server
	ActiveConns map[uuid.UUID](IunctioActiveTunnel) // Active network connections the
	LoginID     uuid.UUID                           // The UUID used for logging in
	Hostname    string                              // The hostname
	Namespace   string                              // What k8s namespace to put this in
	CloseChan   chan os.Signal                      // Used to close client

	// Metrics
	MetricsChan   chan []*io_prometheus_client.MetricFamily
	MetricsServer *http.Server
	MetricsPort   int
}

func NewClientHub() *ClientHub {
	return &ClientHub{
		Reader:      make(chan []byte),
		Writer:      make(chan []byte),
		Control:     make(chan int),
		ActiveConns: make(map[uuid.UUID](IunctioActiveTunnel)),
		CloseChan:   make(chan os.Signal, 1),
		MetricsChan: make(chan []*io_prometheus_client.MetricFamily),
	}
}

func Reader(c *websocket.Conn, hub *ClientHub) {

	defer func() {
		if recover() != nil {
			return
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("[process][read]")
			hub.Control <- 1
			return
		}
		hub.Reader <- message
	}
}

func Writer(c *websocket.Conn, hub *ClientHub) {
	defer func() {
		if recover() != nil {
			return
		}
	}()
	for {
		data, ok := <-hub.Writer

		if !ok {
			// Channel was closed already
			return
		}

		err := c.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Error().Err(err).Msg("[process][Writer]")
			hub.Control <- 1
			return
		}
	}
}

func Error(c *websocket.Conn, hub *ClientHub) {
	_, ok := <-hub.Control

	if !ok {
		log.Info().Msg("Closing Connections")
		c.Close()
		return
	}

	log.Info().Msg("Closing Connections")
	hub.Reader <- ConvertToMessage(ClientGoodbyeMessageType, []byte(""))
	c.Close()
}

func ConvertToMessage(mt MessageType, content []byte) []byte {
	msg := Message{
		Type:    mt,
		Content: content,
	}

	d, err := MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}
	return d
}

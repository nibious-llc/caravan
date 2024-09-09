package common

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net"
	"net/netip"
)

type MessageType int

const (
	TunnelMessageType MessageType = iota
	AdminMessageType

	RequestAuthenticationMsgType
	ResponseCredentialsMsgType
	ResponseAuthenticationSuccessMsgType

	RequestTunnelDescriptionsMsgType
	ResponseTunnelDescriptionsMsgType

	RequestTunnelInitMsgType
	ResponseTunnelInitMsgType // Basically yeah or nah

	RequestTunnelCloseMsgType
	ResponseTunnelCloseMsgType

	TunnelDataMsgType // Either direction

	PingDataMsgType
	PongDataMsgType
	RequestMetricsMsgType        // For metrics, Server asking client
	ResponseMetricsReportMsgType // For metrics, client responding to server

	ClientGoodbyeMessageType
)

type Message struct {
	Type    MessageType
	Content []byte
}

type TunnelInit struct {
	IPAddr netip.Addr `json:"ip_addr"`
	Port   int        `json:"port"`
}

type Tunnel struct {
	IPAddr       netip.Addr
	Port         int
	SessionID    string
	Listener     net.Listener
	ListenerPort int
	Control      chan int
}

func NewIunctioActiveTunnel(conn net.Conn) *IunctioActiveTunnel {
	return &IunctioActiveTunnel{
		Reader:         make(chan []byte),
		Control:        make(chan []byte),
		Conn:           conn,
		ThreadsRunning: false,
	}
}

type TunnelStart struct {
	IPAddr netip.Addr
	Port   int
	Id     uuid.UUID
}

type TunnelData struct {
	Data string //base64 encoded
	Id   uuid.UUID
}

// This struct contains members that describe an active tunnel connection and
// associated data. This includes the net.Conn structure, required channels, and
// more.
type IunctioActiveTunnel struct {
	Id             uuid.UUID   // The UUID of the tunnel
	Conn           net.Conn    // The network connection
	Reader         chan []byte // The channel that will send data to the other end of this tcp socket
	Control        chan []byte
	ThreadsRunning bool
	// Writer is not needed because we can use the clienthub writer
}

type IunctioClientLogin struct {
	ClientID uuid.UUID // The ID to sign in with
	Secret   string    // The secret ID for the client
	Hostname string    // The hostname
}

type IunctioClient struct {
	gorm.Model
	IunctioClientLogin `gorm:"embedded"`
	Namespace          string // What k8s namespace to put this in
}

type Client struct {
	ClientID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Secret      string    `gorm:"size:64;primaryKey"`
	Type        string
	DisplayName string
}

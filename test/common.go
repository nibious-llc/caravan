package test

import (
	"github.com/google/uuid"
	"github.com/nibious-llc/caravan/internal/client"
	"github.com/nibious-llc/caravan/internal/common"
	"github.com/nibious-llc/caravan/internal/server"
	"github.com/rs/zerolog/log"
)

// Setup fake auth provider
type FakeAuthProvider struct {
}

func (ap FakeAuthProvider) IsLoginValid(clientID uuid.UUID, secret string) (common.IunctioClient, bool) {
	fakeClient := common.IunctioClient{}
	return fakeClient, true
}

func setupTest() (serverSide *common.ClientHub, clientSide *common.ClientHub) {

	// Create both sides of the connection
	clientSide = common.NewClientHub()
	serverSide = common.NewClientHub()

	server.AuthProvider = FakeAuthProvider{}

	// Start the appropriate threads for managing the connections, but don't worry about the reader/writer/threads yet
	go server.HandleClient(serverSide)
	go client.HandleConnection(clientSide)

	// Setup communication between the two instances. This is replacing the websocket connection and just passing messages between the two channels
	//go transferMessage(clientSide, serverSide)
	//go transferMessage(serverSide, clientSide)

	//go error(clientSide)
	//go error(serverSide)

	return
}

func transferMessage(hub1 *common.ClientHub, hub2 *common.ClientHub) {

	defer func() {
		if recover() != nil {
			return
		}
	}()

	for {
		msg, ok := <-hub1.Writer
		if !ok {
			log.Debug().Msg("[process][transferMessage] closed")
			hub1.Control <- 1
			return
		}
		hub2.Reader <- msg
	}
}

func error(hub *common.ClientHub) {
	_, ok := <-hub.Control

	if !ok {
		log.Info().Msg("Closing Connections")
		return
	}

	log.Info().Msg("Closing Connections")
	hub.Reader <- common.ConvertToMessage(common.ClientGoodbyeMessageType, []byte(""))
}

func cleanTestForHub(hub *common.ClientHub) {
	hub.Control <- 1
}

func cleanTest(hub1 *common.ClientHub, hub2 *common.ClientHub) {
	cleanTestForHub(hub1)
	cleanTestForHub(hub2)
}

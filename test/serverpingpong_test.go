package test

import (
	"bytes"
	"testing"

	"github.com/nibious-llc/caravan/internal/common"
)

func TestServerPingPong(t *testing.T) {

	serverSide, _ := setupTest()

	// Eat the auth request. If we don't the reader below will block as well.
	<-serverSide.Writer

	// Insert ping message to reader
	serverSide.Reader <- common.GeneratePing()

	// Check that the server gets a pong
	msg, ok := <-serverSide.Writer

	if !ok {
		t.Fatalf("Could not get message from serverSide.Writer")
	}

	if !bytes.Equal(msg, common.GeneratePong()) {
		t.Fatalf("Did not recieve pong from server")
	}
}

package test

import (
	"bytes"
	"github.com/nibious-llc/caravan/internal/common"
	"testing"
)

// This may occasionally fail because of the timer that automatically
// sends a ping every 30 seconds.
func TestClientPingPong(t *testing.T) {

	_, clientSide := setupTest()

	// Insert ping message to reader
	clientSide.Reader <- common.GeneratePing()

	// Check that the server gets a pong
	msg, ok := <-clientSide.Writer

	if !ok {
		t.Fatalf("Could not get message from clientSide.Writer")
	}

	if !bytes.Equal(msg, common.GeneratePong()) {
		t.Fatalf("Did not recieve pong from client")
	}
}

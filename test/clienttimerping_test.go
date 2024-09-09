package test

import (
	"bytes"
	"github.com/nibious-llc/caravan/internal/common"
	"testing"
	"time"
)

func TestClientTimerPing(t *testing.T) {

	timeout := time.After(34 * time.Second)
	done := make(chan int)
	go func() {

		_, clientSide := setupTest()

		// Check that the server gets a pong
		msg, ok := <-clientSide.Writer

		if !ok {
			done <- 1

		}

		if !bytes.Equal(msg, common.GeneratePing()) {
			done <- 2

		}
		done <- 0
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case r := <-done:
		switch r {
		case 0:
			break
		case 1:
			t.Fatalf("Could not get message from clientSide.Writer")
		case 2:
			t.Fatalf("Did not recieve pong from client")
		}
	}
}

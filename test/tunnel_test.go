package test

import (
	"os"
	"testing"

	c "github.com/nibious-llc/caravan/internal/common"
)

func TestStartStop(t *testing.T) {

	os.Setenv("CLIENTID", "2ba76aae-e7f2-4e0b-9e1b-0012be0d45e0")
	os.Setenv("SECRET", "cGskIYKEeorZjyo3KwzUsAIM2kjkoQ9utkhoCFlfelrNZsRmeehcRPPfUxrg0kxJ")

	serverSide, clientSide := setupTest()

	// Setup goroutines for all of the comms
	go transferMessage(clientSide, serverSide)
	go transferMessage(serverSide, clientSide)

	go error(clientSide)
	go error(serverSide)

	// Tell both hubs to terminate
	serverSide.Reader <- c.ConvertToMessage(c.ClientGoodbyeMessageType, []byte(""))
	clientSide.Reader <- c.ConvertToMessage(c.ClientGoodbyeMessageType, []byte(""))
}

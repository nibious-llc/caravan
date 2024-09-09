package client

import (
	"github.com/google/uuid"
	c "github.com/nibious-llc/caravan/internal/common"
	"github.com/rs/zerolog/log"
	"os"
)

//TODO: Grab the login info from the environment

// This is for the client to send login credentials
func generateLoginCredsMsg() []byte {

	// Start the login process
	var login_data c.IunctioClientLogin

	CLIENTID := os.Getenv("CLIENTID")
	if CLIENTID == "" {
		log.Panic().Msg("env var 'CLIENTID' must be set")
	}

	SECRET := os.Getenv("SECRET")
	if SECRET == "" {
		log.Panic().Msg("env var 'SECRET' must be set")
	}

	login_data.ClientID = uuid.MustParse(CLIENTID)
	login_data.Secret = SECRET

	data_to_send, err := c.MarshalObject(login_data)
	if err != nil {
		log.Error().Err(err)
		return nil
	}

	return c.ConvertToMessage(c.ResponseCredentialsMsgType, data_to_send)
}

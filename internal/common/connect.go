package common

import (
	"github.com/rs/zerolog/log"
)

func GeneratePong() []byte {

	msg := Message{
		Type:    PongDataMsgType,
		Content: nil,
	}

	d, err := MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d

}

func GeneratePing() []byte {

	msg := Message{
		Type:    PingDataMsgType,
		Content: nil,
	}

	d, err := MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d

}

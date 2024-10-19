package main

import (
	"github.com/rs/zerolog/log"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

// HandlerLoop blocks until the handler has finished. The result of the execution is given by Handler.Result().
func HandlerLoop(myid party.ID, h protocol.Handler, comm *Comm) {
	log.Debug().Msgf("HanderLoop started")
	for {
		select {

		// outgoing messages
		case msg, ok := <-h.Listen():
			if !ok {
				<-comm.Done()
				// the channel was closed, indicating that the protocol is done executing.
				return
			}
			go comm.Send(msg)

		// incoming messages
		case msg := <-comm.Next():
			h.Accept(msg)
		}
	}
}

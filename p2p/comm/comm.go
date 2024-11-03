package comm

import (
	"context"
	"io"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/rs/zerolog/log"
	"github.com/taurusgroup/multi-party-sig/p2p/utils"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

const ProtocolID = "/multipartysig/1.0.0"

// Comm the p2p communicator for passing messages between the parties
type Comm struct {
	parties   party.IDSlice
	myPartyId party.ID
	p2pHost   host.Host
	inMsg     chan *protocol.Message
}

// NewComm constructs a new Comm; myPartyId is the party ID of the current party
// and parties is the list of all party IDs;
// It creates a libp2p network and one stream to each party
// PartyID == libp2p Host ID
// streams are uni-directional only for sending messages
func NewComm(myPartyId party.ID, parties party.IDSlice, host host.Host) *Comm {
	sort.Sort(parties)
	N := len(parties)
	comm := &Comm{
		myPartyId: myPartyId,
		parties:   parties,
		p2pHost:   host,
		inMsg:     make(chan *protocol.Message, 10*N),
	}
	host.SetStreamHandler(ProtocolID, comm.commStreamHandler)
	return comm
}

// Next returns the channel that gives incoming message
func (comm *Comm) Next() <-chan *protocol.Message {

	return comm.inMsg
}

func (comm *Comm) Send(msg *protocol.Message) {
	log.Debug().Msgf("sending message(%s) to %s: data len(%d)", msg.Protocol, msg.To, len(msg.Data))
	if msg.From != comm.myPartyId {
		log.Error().Msgf("cannot send; from mismatch; expected %s, got %s", comm.myPartyId, msg.From)
		return
	}

	for _, party := range comm.parties {
		go func() {
			if !msg.IsFor(party) {
				return
			}
			pid, err := utils.PartyIDToPeerID(party)
			stream, err := comm.p2pHost.NewStream(context.Background(), pid, ProtocolID)
			if err != nil {
				log.Error().Msgf("failed to open stream to %s: %v", msg.To, err)
				spew.Dump(msg)
				return
			}
			defer stream.Close()

			stream.CloseRead() // write only stream
			bz, err := msg.MarshalBinary()
			if err != nil {
				log.Error().Msgf("failed to marshal message: %v", err)
				return
			}
			_, err = stream.Write(bz)
			if err != nil {
				log.Error().Msgf("failed to write to stream: %v", err)
				return
			}
			log.Debug().Msgf("msg content: round(%d), protocol(%s), broadcast(%v)", msg.RoundNumber, msg.Protocol, msg.Broadcast)
			log.Debug().Msgf("sent msg(%d)", len(bz))
			//spew.Dump("sent msg", msg)

		}()
	}
}

func (comm *Comm) Done() <-chan struct{} {
	log.Debug().Msgf("done(%s)", comm.myPartyId)
	done := make(chan struct{})
	close(done)
	return done
}

// handling incoming msgs
func (comm *Comm) commStreamHandler(inStream network.Stream) {
	log.Debug().Msgf("incoming new stream from %s", inStream.Conn().RemotePeer())
	defer inStream.Close()

	buf, err := io.ReadAll(inStream)
	if err != nil {

		log.Error().Msgf("failed to read from stream: %v", err)
		return
	}
	msg := &protocol.Message{}
	log.Debug().Msgf("raw recv msg(%d)", len(buf))
	err = msg.UnmarshalBinary(buf)
	if err != nil {
		log.Error().Msgf("failed to unmarshal message: %v", err)
		return
	}
	log.Debug().Msgf("received message(%s) from %s: data len(%d): round(%d)", msg.Protocol, msg.From, len(msg.Data), msg.RoundNumber)
	comm.inMsg <- msg
}

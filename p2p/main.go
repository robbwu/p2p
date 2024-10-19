package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"golang.org/x/crypto/sha3"
)

const keyFileName = "peer_id.key"

func loadOrCreateIdentity() (crypto.PrivKey, error) {
	if _, err := os.Stat(keyFileName); err == nil {
		// Key file exists, load it
		data, err := ioutil.ReadFile(keyFileName)
		if err != nil {
			return nil, err
		}
		keyBytes, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, err
		}
		privKey, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("Loaded existing peer ID")
		return privKey, nil
	}

	// Key file does not exist, create a new one
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	keyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(keyFileName, []byte(base64.StdEncoding.EncodeToString(keyBytes)), 0600)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Created new peer ID and saved to file")
	return privKey, nil
}

func computeSessionID() int64 {
	// Get the current Unix timestamp
	now := time.Now().Unix()

	// Truncate to a 1000s window by removing the seconds part
	sessionID := now / 1000 * 1000

	return sessionID
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	privKey, err := loadOrCreateIdentity()
	if err != nil {
		panic(err)
	}

	limits := rcmgr.DefaultLimits
	limits.StreamBaseLimit.Streams = 128
	limits.StreamBaseLimit.StreamsInbound = 64
	limits.StreamBaseLimit.StreamsOutbound = 64
	limitConf := limits.AutoScale()
	limiter := rcmgr.NewFixedLimiter(limitConf)
	rmgr, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		panic(err)
	}

	host, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.Security("/tls", libp2ptls.New),
		libp2p.Security("/noise", noise.New),
		libp2p.ResourceManager(rmgr),
	)

	log.Info().Msgf("My ID is %s", host.ID())

	if err != nil {
		panic(err)
	}
	kademliaDHT, err := dht.New(context.Background(), host, dht.Mode(dht.ModeServer))
	if err = kademliaDHT.Bootstrap(context.Background()); err != nil {
		panic(err)
	}

	for _, addr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
		if err := host.Connect(context.Background(), *peerinfo); err != nil {
			log.Debug().Err(err).Msg("Connection failed")
		} else {
			log.Info().Msgf("Connection established with bootstrap node: %v", *peerinfo)
		}
	}

	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	ns := fmt.Sprintf("multipartysig-test-%d", computeSessionID())
	log.Info().Msgf("Announcing ourselves with session ID %s...", ns)
	dutil.Advertise(context.Background(), routingDiscovery, ns)

	//pingService := ping.NewPingService(host)

	var parties []party.ID
	myPartyId, err := PeerIDToPartyID(host.ID())
	if err != nil {
		panic(err)
	}
	parties = append(parties, myPartyId)
	foundPeer := false
	for range time.NewTicker(5 * time.Second).C {
		if foundPeer {
			break
		}
		log.Info().Msg("Searching for other peers...")
		peerChan, err := routingDiscovery.FindPeers(context.Background(), ns)
		if err != nil {
			panic(err)
		}
		for peer := range peerChan {
			if peer.ID == host.ID() {
				log.Debug().Msg("Found self")
				continue
			}
			log.Info().Msgf("Found peer: %s! Connecting...", peer.ID)
			if err = host.Connect(context.Background(), peer); err != nil {
				log.Error().Err(err).Msg("Connecting peer failed")
				continue
			}
			log.Info().Msgf("OK: Connected to peer: %s!", peer.ID)
			conns := host.Network().ConnsToPeer(peer.ID)
			for _, conn := range conns {
				spew.Dump("local multiaddr", conn.LocalMultiaddr())
				spew.Dump("remote multiaddr", conn.RemoteMultiaddr())
				log.Info().Msgf("connection security %v", conn.ConnState())
			}
			foundPeer = true
			partyId, err := PeerIDToPartyID(peer.ID)
			if err != nil {
				panic(err)
			}
			parties = append(parties, partyId)
			// attept to create a stream--note this can fail if the peer hasn't register a handler for
			// the protocol

		}
	}

	log.Info().Msgf("Parties connected; total: %d", len(parties))
	log.Info().Msgf("myPartyID: %s", myPartyId)
	log.Info().Msgf("parties: %s", parties)

	comm := NewComm(myPartyId, parties, host)

	// wait until all peers have registered the protocolID
	log.Info().Msgf("Waiting for all peers to register protocolID %s...", protocolID)
	var wg sync.WaitGroup
	for _, party := range parties {
		if party == myPartyId {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				pid, err := PartyIDToPeerID(party)
				if err != nil {
					panic(err)
				}
				stream, err := host.NewStream(context.Background(), pid, protocolID)
				if err != nil {
					log.Warn().Err(err).Msg("failed to create stream; retrying...")
				} else {
					stream.Close()
					return
				}
				time.Sleep(3 * time.Second)
			}
		}()
	}
	wg.Wait()

	threshold := 1
	partiesSlice := party.NewIDSlice(parties)
	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, myPartyId, partiesSlice, threshold, nil), nil)
	if err != nil {
		panic(err)
	}
	s := time.Now()
	HandlerLoop(myPartyId, h, comm)

	r, err := h.Result()
	if err != nil {
		panic(err)
	}

	config, ok := r.(*cmp.Config)
	if !ok {
		panic("unexpected type")
	}
	//spew.Dump(config)
	log.Info().Msgf("Keygen success!: parties(%d), threshold(%d)", len(config.PartyIDs()), config.Threshold)
	log.Info().Msgf("Keygen takes %s", time.Since(s))
	pkbz, err := config.PublicPoint().MarshalBinary()

	pk, err := secp256k1.ParsePubKey(pkbz)
	if err != nil {
		panic(err)
	}
	uncompressed := pk.SerializeUncompressed()
	hash := sha3.NewLegacyKeccak256()
	hash.Write(uncompressed)
	hashedPublicKey := hash.Sum(nil)
	ethAddr := hashedPublicKey[12:]
	log.Info().Msgf("Public: Ethereum address: %x", ethAddr)

	{
		hash := sha3.NewLegacyKeccak256()
		hash.Write([]byte("hello multisig"))
		m := hash.Sum(nil)
		h, err := protocol.NewMultiHandler(cmp.Sign(config, parties, m, nil), nil)
		if err != nil {
			panic(err)
		}
		s := time.Now()
		HandlerLoop(myPartyId, h, comm)
		log.Info().Msgf("Keysign takes %s", time.Since(s))
		signResult, err := h.Result()
		if err != nil {
			panic(err)
		}
		signature := signResult.(*ecdsa.Signature)
		if !signature.Verify(config.PublicPoint(), m) {
			panic(err)
		}
		log.Info().Msg("Keysign success: Signature verified!")
	}

}

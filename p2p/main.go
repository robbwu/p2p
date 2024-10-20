package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/taurusgroup/multi-party-sig/p2p/cmd"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cmd.Execute()
	//
	//N := flag.Int("n", 3, "keygen parties")
	//threshold := flag.Int("t", 1, "keygen threshold")
	//dir := flag.String("d", "", "directory to store peer ID")
	//flag.Parse()
	//
	//
	//privKey, err := loadOrCreateIdentity(*dir)
	//if err != nil {
	//	panic(err)
	//}
	//
	////pl := pool.NewPool(0)
	//pl := pool.NewPool(1)
	//
	//limits := rcmgr.DefaultLimits
	//limits.StreamBaseLimit.Streams = 128
	//limits.StreamBaseLimit.StreamsInbound = 64
	//limits.StreamBaseLimit.StreamsOutbound = 64
	//limitConf := limits.AutoScale()
	//limiter := rcmgr.NewFixedLimiter(limitConf)
	//rmgr, err := rcmgr.NewResourceManager(limiter)
	//if err != nil {
	//	panic(err)
	//}
	//
	//host, err := libp2p.New(
	//	libp2p.Identity(privKey),
	//	libp2p.Security("/tls", libp2ptls.New),
	//	libp2p.Security("/noise", noise.New),
	//	libp2p.ResourceManager(rmgr),
	//)
	//
	//log.Info().Msgf("My ID is %s", host.ID())
	//
	//if err != nil {
	//	panic(err)
	//}
	//kademliaDHT, err := dht.New(context.Background(), host, dht.Mode(dht.ModeServer))
	//if err = kademliaDHT.Bootstrap(context.Background()); err != nil {
	//	panic(err)
	//}
	//
	//for _, addr := range dht.DefaultBootstrapPeers {
	//	peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
	//	if err := host.Connect(context.Background(), *peerinfo); err != nil {
	//		log.Debug().Err(err).Msg("Connection failed")
	//	} else {
	//		log.Info().Msgf("Connection established with bootstrap node: %v", *peerinfo)
	//	}
	//}
	//
	//routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	//ns := fmt.Sprintf("multipartysig-test-%d", computeSessionID())
	//log.Info().Msgf("Announcing ourselves with session ID %s...", ns)
	//dutil.Advertise(context.Background(), routingDiscovery, ns)
	//
	////pingService := ping.NewPingService(host)
	//
	//var parties []party.ID
	//myPartyId, err := PeerIDToPartyID(host.ID())
	//if err != nil {
	//	panic(err)
	//}
	//parties = append(parties, myPartyId)
	//
	//var connectedPeers []peer.ID
	//for range time.NewTicker(5 * time.Second).C {
	//	if len(parties) == *N {
	//		break
	//	}
	//	log.Info().Msg("Searching for other peers...")
	//	peerChan, err := routingDiscovery.FindPeers(context.Background(), ns)
	//	if err != nil {
	//		panic(err)
	//	}
	//	for peer := range peerChan {
	//		if peer.ID == host.ID() {
	//			log.Debug().Msg("Found self")
	//			continue
	//		}
	//		if slices.Contains(connectedPeers, peer.ID) {
	//			log.Debug().Msg("Already connected")
	//			continue
	//		}
	//		log.Info().Msgf("Found peer: %s! Connecting...", peer.ID)
	//		if err = host.Connect(context.Background(), peer); err != nil {
	//			log.Error().Err(err).Msg("Connecting peer failed")
	//			continue
	//		}
	//		log.Info().Msgf("OK: Connected to peer: %s!", peer.ID)
	//		conns := host.Network().ConnsToPeer(peer.ID)
	//		for _, conn := range conns {
	//			spew.Dump("local multiaddr", conn.LocalMultiaddr())
	//			spew.Dump("remote multiaddr", conn.RemoteMultiaddr())
	//			log.Info().Msgf("connection security %v", conn.ConnState())
	//		}
	//
	//		partyId, err := PeerIDToPartyID(peer.ID)
	//		if err != nil {
	//			panic(err)
	//		}
	//		parties = append(parties, partyId)
	//		connectedPeers = append(connectedPeers, peer.ID)
	//	}
	//}
	//
	//log.Info().Msgf("Parties connected; total: %d", len(parties))
	//log.Info().Msgf("myPartyID: %s", myPartyId)
	//log.Info().Msgf("parties: %s", parties)
	//
	//comm := NewComm(myPartyId, parties, host)
	//
	//// wait until all peers have registered the protocolID
	//log.Info().Msgf("Waiting for all peers to register protocolID %s...", protocolID)
	//var wg sync.WaitGroup
	//for _, party := range parties {
	//	if party == myPartyId {
	//		continue
	//	}
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		for {
	//			pid, err := PartyIDToPeerID(party)
	//			if err != nil {
	//				panic(err)
	//			}
	//			stream, err := host.NewStream(context.Background(), pid, protocolID)
	//			if err != nil {
	//				log.Warn().Err(err).Msg("failed to create stream; retrying...")
	//			} else {
	//				stream.Close()
	//				return
	//			}
	//			time.Sleep(3 * time.Second)
	//		}
	//	}()
	//}
	//wg.Wait()
	//log.Info().Msg("All peers have registered protocolID")
	//go func() {
	//	for _, peer := range connectedPeers {
	//		go func() {
	//			for {
	//				rtt := <-ping.Ping(context.Background(), host, peer)
	//				log.Info().Msgf("RTT to %s: %s", peer, rtt)
	//				time.Sleep(2 * time.Second)
	//			}
	//		}()
	//	}
	//}()
	//
	//partiesSlice := party.NewIDSlice(parties)
	//h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, myPartyId, partiesSlice, *threshold, pl), nil)
	//if err != nil {
	//	panic(err)
	//}
	//s := time.Now()
	//HandlerLoop(myPartyId, h, comm)
	//
	//r, err := h.Result()
	//if err != nil {
	//	panic(err)
	//}
	//
	//config, ok := r.(*cmp.Config)
	//if !ok {
	//	panic("unexpected type")
	//}
	////spew.Dump(config)
	//log.Info().Msgf("Keygen success!: parties(%d), threshold(%d)", len(config.PartyIDs()), config.Threshold)
	//log.Info().Msgf("Keygen takes %s", time.Since(s))
	//pkbz, err := config.PublicPoint().MarshalBinary()
	//
	//pk, err := secp256k1.ParsePubKey(pkbz)
	//if err != nil {
	//	panic(err)
	//}
	//uncompressed := pk.SerializeUncompressed()
	//hash := sha3.NewLegacyKeccak256()
	//hash.Write(uncompressed)
	//hashedPublicKey := hash.Sum(nil)
	//ethAddr := hashedPublicKey[12:]
	//log.Info().Msgf("Public: Ethereum address: %x", ethAddr)
	//
	//{
	//	removeParty := comm.parties[0]
	//	comm.parties = comm.parties[1:]
	//	if removeParty == myPartyId {
	//		log.Info().Msgf("out of signers; skipping")
	//		return
	//	}
	//	hash := sha3.NewLegacyKeccak256()
	//	hash.Write([]byte("hello multisig"))
	//	m := hash.Sum(nil)
	//	h, err := protocol.NewMultiHandler(cmp.Sign(config, comm.parties, m, pl), nil)
	//	if err != nil {
	//		panic(err)
	//	}
	//	s := time.Now()
	//
	//	HandlerLoop(myPartyId, h, comm)
	//	log.Info().Msgf("Keysign takes %s", time.Since(s))
	//	signResult, err := h.Result()
	//	if err != nil {
	//		panic(err)
	//	}
	//	signature := signResult.(*ecdsa.Signature)
	//	if !signature.Verify(config.PublicPoint(), m) {
	//		panic(err)
	//	}
	//	log.Info().Msgf("Keysign success (%d/%d): Signature verified!", len(comm.parties), N)
	//}

}

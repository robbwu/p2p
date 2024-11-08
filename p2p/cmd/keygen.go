/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/taurusgroup/multi-party-sig/p2p/encryption"

	//"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	myp2p "github.com/taurusgroup/multi-party-sig/p2p/comm"
	"github.com/taurusgroup/multi-party-sig/p2p/handler"
	"github.com/taurusgroup/multi-party-sig/p2p/utils"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"golang.org/x/crypto/sha3"

	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	commpkg "github.com/taurusgroup/multi-party-sig/p2p/comm"
)

var N int
var threshold int

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "generate a ECDSA private key",
	Long:  `ECDSA distributed keygen`,
	Run: func(cmd *cobra.Command, args []string) {

		if password == "" {
			log.Info().Msgf("No password via CLI arguments; reading from stdin...")
			pw, err := utils.GetPassword("Enter password: ")
			if err != nil {
				panic(err)
			}
			password = pw
		}
		fmt.Println("keygen called")
		if cfgDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			cfgDir = path.Join(homeDir, ".p2p")
		}
		log.Info().Msgf("Config Directory: %s", cfgDir)
		log.Info().Msgf("Vault: %s", vault)
		vaultDir := path.Join(cfgDir, vault)

		privKey, err := utils.LoadOrCreateIdentity(vaultDir)
		if err != nil {
			panic(err)
		}

		//pl := pool.NewPool(0)
		pl := pool.NewPool(0)

		host, err := libp2p.New(
			libp2p.Identity(privKey),
			utils.P2POptions(),
		)
		if err != nil {
			panic(err)
		}

		log.Info().Msgf("My ID is %s", host.ID())
		log.Info().Msgf("my address is %s", host.Addrs())

		if token == "" {
			log.Error().Msgf("Session token not provided")
			return
		}
		ns := fmt.Sprintf("%s-%d", token, utils.ComputeSessionID())
		comm, connectedPeers, parties := MustConnectWithEnoughPeers(host, N, nil, ns)
		go func() {
			for _, peer := range connectedPeers {
				go func() {
					for {
						rtt := <-ping.Ping(context.Background(), host, peer)
						log.Info().Msgf("RTT to %s: %s", peer, rtt)
						time.Sleep(2 * time.Second)
					}
				}()
			}
		}()

		myPartyId, err := utils.PeerIDToPartyID(host.ID())
		if err != nil {
			panic(err)
		}
		partiesSlice := party.NewIDSlice(parties)
		h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, myPartyId, partiesSlice, threshold, pl), nil)
		if err != nil {
			panic(err)
		}
		s := time.Now()
		handler.HandlerLoop(myPartyId, h, comm)

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

		rawbz, err := config.MarshalBinary()
		if err != nil {
			panic(err)
		}
		bz, err := encryption.Encrypt(rawbz, []byte(password))
		if err != nil {
			panic(err)
		}

		configPath := path.Join(vaultDir, "keygen_config.json")
		if _, err := os.Stat(configPath); err == nil {
			log.Warn().Msgf("Config file already exists; making a backup...")
			err = os.Rename(configPath, fmt.Sprintf("%s.%d", configPath, time.Now().Unix()))
			if err != nil {
				log.Error().Err(err).Msg("failed to make backup; saving current keygen to a temporary file...")

				os.WriteFile(configPath+".tmp", bz, 0600)
				return
			}
		}

		os.WriteFile(configPath, bz, 0600)
		log.Info().Msgf("Config saved to %s", configPath)
	},
}

// this is blocking until the specific amount of peers are found
func MustConnectWithEnoughPeers(host host.Host, numPeers int, whitelist []peer.ID, ns string) (*commpkg.Comm, []peer.ID, party.IDSlice) {
	if len(whitelist) > 0 && len(whitelist) < numPeers {
		panic("whitelist must be empty or at least the same size as numPeers")
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
			log.Debug().Msgf("Connection established with bootstrap node: %v", *peerinfo)
		}
	}

	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	log.Info().Msgf("Announcing ourselves with session ID %s...", ns)
	dutil.Advertise(context.Background(), routingDiscovery, ns)

	//pingService := ping.NewPingService(host)

	var parties []party.ID
	myPartyId, err := utils.PeerIDToPartyID(host.ID())
	if err != nil {
		panic(err)
	}
	parties = append(parties, myPartyId)

	var connectedPeers []peer.ID
	for range time.NewTicker(5 * time.Second).C {
		if len(parties) == numPeers {
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
			if slices.Contains(connectedPeers, peer.ID) {
				log.Debug().Msg("Already connected")
				continue
			}
			if len(whitelist) > 0 && !slices.Contains(whitelist, peer.ID) {
				log.Warn().Msgf("Peer %s not in whitelist; do not connecte!", peer.ID)
				continue
			}
			log.Info().Msgf("Found peer: %s! Connecting...", peer.ID)

			cont := false
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel() // Ensure the context is canceled to avoid resource leaks
				if err = host.Connect(ctx, peer); err != nil {
					log.Error().Err(err).Msg("Connecting peer failed")
					cont = true
				}
				log.Info().Msgf("OK: Connected to peer: %s!", peer.ID)
			}()
			if cont {
				continue
			}
			conns := host.Network().ConnsToPeer(peer.ID)
			for _, conn := range conns {
				log.Info().Msgf("connection security %v", conn.ConnState())
			}

			partyId, err := utils.PeerIDToPartyID(peer.ID)
			if err != nil {
				panic(err)
			}
			parties = append(parties, partyId)
			connectedPeers = append(connectedPeers, peer.ID)
		}
	}

	log.Info().Msgf("Parties connected; total: %d", len(parties))
	log.Info().Msgf("myPartyID: %s", myPartyId)
	log.Info().Msgf("parties: %s", parties)

	comm := myp2p.NewComm(myPartyId, parties, host)

	// wait until all peers have registered the protocolID
	log.Info().Msgf("Waiting for all peers to register protocolID %s...", commpkg.ProtocolID)
	var wg sync.WaitGroup
	for _, party := range parties {
		if party == myPartyId {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				pid, err := utils.PartyIDToPeerID(party)
				if err != nil {
					panic(err)
				}
				stream, err := host.NewStream(context.Background(), pid, commpkg.ProtocolID)
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
	log.Info().Msg("All peers have registered protocolID")
	return comm, connectedPeers, parties
}

func PointToPubkeyUncompressed65B(p curve.Point) []byte {
	bz, err := p.MarshalBinary() // 33B compressed pubkey
	if err != nil {
		panic(err)
	}
	pk, err := secp256k1.ParsePubKey(bz)
	if err != nil {
		panic(err)
	}
	uncompressed := pk.SerializeUncompressed()
	return uncompressed

}

func UncompressedToEthAddr(uncompressed []byte) []byte {
	if len(uncompressed) != 65 {
		panic("invalid uncompressed pubkey length")
	}
	hash := sha3.NewLegacyKeccak256()
	hash.Write(uncompressed)
	hashedPublicKey := hash.Sum(nil)
	ethAddr := hashedPublicKey[12:]
	return ethAddr
}

func init() {
	rootCmd.AddCommand(keygenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// keygenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// keygenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	keygenCmd.Flags().IntVar(&N, "n", 3, "keygen parties (default to 3)")
	keygenCmd.Flags().IntVar(&threshold, "t", 1, "keygen threshold; need t+1 to sign (default to 1)")

}

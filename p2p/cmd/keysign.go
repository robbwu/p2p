/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/p2p/handler"
	"github.com/taurusgroup/multi-party-sig/p2p/utils"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
)

var messageHashHex string

// keysignCmd represents the keysign command
var keysignCmd = &cobra.Command{
	Use:   "keysign",
	Short: "ECDSA keysign a message (hash), producing signature",
	Long:  `ECDSA keysign a message (hash), producing signature`,
	Run: func(cmd *cobra.Command, args []string) {
		msghash, err := hex.DecodeString(messageHashHex)
		if err != nil {
			panic(err)
		}
		if len(msghash) != 32 {
			panic("message hash must be 32 bytes")
		}

		// read keygen config
		config, vaultDir := readConfig()
		log.Info().Msgf("N %d, threshold %d", len(config.PartyIDs()), config.Threshold)

		privKey, err := utils.LoadOrCreateIdentity(vaultDir)
		if err != nil {
			panic(err)
		}

		pl := pool.NewPool(0)

		host, err := libp2p.New(
			libp2p.Identity(privKey),
			utils.P2POptions(),
		)

		log.Info().Msgf("My ID is %s", host.ID())

		myPartyId, err := utils.PeerIDToPartyID(host.ID())
		if err != nil {
			panic(err)
		}
		var peers []peer.ID
		for _, party := range config.PartyIDs() {
			pid, err := utils.PartyIDToPeerID(party)
			if err != nil {
				panic(err)
			}
			peers = append(peers, pid)
		}
		if token == "" {
			log.Error().Msgf("Session token not provided")
			return
		}
		ns := fmt.Sprintf("%s-%d", token, utils.ComputeSessionID())
		comm, _, parties := MustConnectWithEnoughPeers(host, config.Threshold+1, peers, ns)

		partiesSlice := party.NewIDSlice(parties)
		h, err := protocol.NewMultiHandler(cmp.Sign(config, partiesSlice, msghash, pl), nil)
		if err != nil {
			panic(err)
		}
		s := time.Now()
		handler.HandlerLoop(myPartyId, h, comm)
		log.Info().Msgf("Keysign takes %s", time.Since(s))

		signResult, err := h.Result()
		if err != nil {
			panic(err)
		}
		signature := signResult.(*ecdsa.Signature)
		if !signature.Verify(config.PublicPoint(), msghash) {
			panic(err)
		}
		log.Info().Msgf("Keysign success (%d/%d): Signature verified!", len(partiesSlice), len(config.PartyIDs()))
		sig, err := signature.SigEthereum()
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("Signature(Ethereum)(%d): %x", len(sig), sig)
		ethAddr := UncompressedToEthAddr(PointToPubkeyUncompressed65B(config.PublicPoint()))
		log.Info().Msgf("Ethereum address 0x%x", ethAddr)
	},
}

func init() {
	rootCmd.AddCommand(keysignCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// keysignCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// keysignCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	keysignCmd.Flags().StringVar(&messageHashHex, "msg-hash", "", "message hash in hex format: 32B, no leading 0x")
}

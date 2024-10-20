/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/hex"
	"os"
	"path"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/p2p/handler"
	"github.com/taurusgroup/multi-party-sig/p2p/utils"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
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
		if cfgDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			cfgDir = path.Join(homeDir, ".p2p")
		}
		vaultDir := path.Join(cfgDir, vault)
		log.Info().Msgf("describe %s", vaultDir)
		configPath := path.Join(vaultDir, "keygen_config.json")
		// read file
		data, err := os.ReadFile(configPath)
		if err != nil {
			panic(err)
		}
		log.Debug().Msgf("File size %d", len(data))
		config := cmp.EmptyConfig(curve.Secp256k1{})
		err = config.UnmarshalBinary(data)
		if err != nil {
			panic(err)
		}
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
		comm, _, parties := MustConnectWithEnoughPeers(host, config.Threshold+1)

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

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/p2p/encryption"
	"github.com/taurusgroup/multi-party-sig/p2p/utils"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
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
		rawdata, err := os.ReadFile(configPath)
		if err != nil {
			panic(err)
		}
		if password == "" {
			log.Info().Msgf("No password via CLI arguments; reading from stdin...")
			pw, err := utils.GetPassword("Enter password: ")
			if err != nil {
				panic(err)
			}
			password = pw
		}
		data, err := encryption.Decrypt(rawdata, []byte(password))
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
		log.Info().Msgf("my party ID: %s", config.ID)
		log.Info().Msgf("parties: ")
		for _, id := range config.PartyIDs() {
			log.Info().Msgf("  %s", id)
		}

		ethAddr := UncompressedToEthAddr(PointToPubkeyUncompressed65B(config.PublicPoint()))
		log.Info().Msgf("Ethereum address 0x%x", ethAddr)
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// describeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// describeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

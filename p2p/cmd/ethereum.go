/*
Copyright © 2024 brewmaster012
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

// ethereumCmd represents the ethereum command
var ethereumCmd = &cobra.Command{
	Use:   "ethereum",
	Short: "Build an ethereum transaction",
	Long:  `Interactive command to build an ethereum transaction`,
	Run: func(cmd *cobra.Command, args []string) {
		// read keygen config
		config, _ := readConfig()

		log.Info().Msgf("N %d, threshold %d", len(config.PartyIDs()), config.Threshold)

		ethAddr := UncompressedToEthAddr(PointToPubkeyUncompressed65B(config.PublicPoint()))
		log.Info().Msgf("Ethereum address 0x%x", ethAddr)
	},
}

func readConfig() (*cmp.Config, string) {
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
	log.Debug().Msgf("File size %d", len(rawdata))
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

	config := cmp.EmptyConfig(curve.Secp256k1{})
	err = config.UnmarshalBinary(data)
	if err != nil {
		panic(err)
	}
	return config, vaultDir
}

func init() {
	rootCmd.AddCommand(ethereumCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ethereumCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ethereumCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

/*
Copyright © 2024 brewmaster012
*/
package cmd

import (
	"context"
	"os"
	"path"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/p2p/encryption"
	"github.com/taurusgroup/multi-party-sig/p2p/utils"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
)

var rpcUrl string

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

		client, err := ethclient.Dial(rpcUrl)
		if err != nil {
			panic(err)
		}
		block, err := client.BlockByNumber(context.Background(), nil)
		t := time.Unix(int64(block.Time()), 0)
		log.Info().Msgf("Latest block number %d; timestamp %s", block.Number(), t.String())

		from := ethcommon.BytesToAddress(ethAddr)
		bal, err := client.BalanceAt(context.Background(), from, nil)
		if err != nil {
			panic(err)
		}
		balF, _ := bal.Float64()
		log.Info().Msgf("Balance of %s is %.6f", from.Hex(), balF/params.Ether)
		nonce, err := client.NonceAt(context.Background(), from, nil)
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("Nonce of %s is %d", from.Hex(), nonce)
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			panic(err)
		}
		gasPriceF, _ := gasPrice.Float64()
		log.Info().Msgf("Suggested gas price is %.6fgwei", gasPriceF/1.0e9)
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

	ethereumCmd.Flags().StringVar(&rpcUrl, "rpc", "wss://ethereum-rpc.publicnode.com", "Ethereum JSON RPC endpoint;")
}

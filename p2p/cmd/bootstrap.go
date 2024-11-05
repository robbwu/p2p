/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/spf13/cobra"
)

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bootstrap called")
		ip := "100.88.27.113"
		port := "8686"
		addr := fmt.Sprintf("/ip4/%s/tcp/%s", ip, port)

		host, err := libp2p.New(
			libp2p.ListenAddrStrings(addr),
		)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		fmt.Println("Host created; host ID ", host.ID())
		fmt.Println("Host address ", host.Addrs())

		kdht, err := dht.New(context.Background(), host, dht.Mode(dht.ModeServer))
		if err != nil {
			fmt.Println("dht New Error: ", err)
			return
		}
		if err = kdht.Bootstrap(context.Background()); err != nil {
			fmt.Println("Bootstrap Error: ", err)
			return
		}
		fmt.Println("Bootstrap done")
		select {}
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bootstrapCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bootstrapCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

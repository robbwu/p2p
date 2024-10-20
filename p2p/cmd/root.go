/*
Copyright © 2024 brewmaster012
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgDir string
var vault string
var token string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "p2p",
	Short: "p2p multi-party-sig protocol",
	Long:  `Sign messages using multi-party-sig protocol`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.p2p.yaml)")
	rootCmd.PersistentFlags().StringVar(&cfgDir, "dir", "", "config directory (default is $HOME/.p2p)")
	rootCmd.PersistentFlags().StringVar(&vault, "vault", "default", "vault name (default default)")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "session token--make sure each party uses the same token; don't reuse the same token")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   string
	Commit    string
	BuildTime string
	Branch    string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows version and build time",
	Long:  `Shows version commit and the build time of the binary`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", Version)
		fmt.Printf("%s\n", Commit)
		fmt.Printf("%s\n", BuildTime)
		fmt.Printf("%s\n", Branch)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

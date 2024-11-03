/*
Copyright © 2024 brewmaster012
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var words *string

// dropCmd represents the drop command
var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Send/Recv file over p2p network ",
	Long: `To send a file, pipe to the stdin like so
cat file | p2p drop

It's going to generate four words as password. And to
receive the file, the receving side does

p2p drop --recv word1-word2-word3-word4
The received content will be to stdout; redirect to save
into a file. 
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("drop called")
		fmt.Printf("recv flags? %s\n", *words)
		if words == nil { // drop send mode
			// step 1: generate 4 random words from the bip wordlist

			// step 2: derive a session id and connect to p2p peer
			// step 3: encrypt and communicate ciphertext
		} else { // recv mode
			// step 1: from the 4 words derive the session id
			// step 2: connect to the peer
			// step 3: receive and decrypt the message
		}
	},
}

func init() {
	rootCmd.AddCommand(dropCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dropCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dropCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	words = dropCmd.Flags().String("recv", "", "the 4 words password")
}

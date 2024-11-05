/*
Copyright © 2024 brewmaster012
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/p2p/encryption"
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
		if *words == "" { // drop send mode
			fmt.Println("send mode")
			// step 1: generate 4 random words from the bip wordlist
			words, err := encryption.GenerateRandomWords()
			if err != nil {
				fmt.Println("Error generating random words")
				return
			}
			wordsDash := strings.Join(words, "-")
			fmt.Println("words: ", wordsDash)
			fmt.Println("to receive this file, type the following command")
			fmt.Printf("p2p drop --recv %s\n", wordsDash)
			// step 2: derive a session id and connect to p2p peer
			sessionID := SessionIDFromWords(words)
			fmt.Printf("recv side should print the same session id: %s\n", sessionID)
			// step 3: encrypt and communicate ciphertext
			// read from stdin and encrypt
			// step 3: encrypt and communicate ciphertext
			plaintext, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println("Error reading from stdin:", err)
				return
			}

			// Convert words to integers
			keyBytes, err := encryption.WordsToBytes(words)
			if err != nil {
				fmt.Println("Error converting integers to bytes:", err)
				return
			}

			// Encrypt the plaintext using the derived key
			ciphertext, err := encryption.Encrypt(plaintext, keyBytes)
			if err != nil {
				fmt.Println("Error encrypting data:", err)
				return
			}

			// Here you would send the ciphertext to the peer
			fmt.Printf("Encrypted %d bytes of data\n", len(ciphertext))
			fmt.Printf("Sending ciphertext to peer...\n")

			host, err := libp2p.New()
			if err != nil {
				fmt.Println("Error creating host:", err)
				return
			}
			fmt.Printf("host id: %s\n", host.ID())
			kdht, err := dht.New(context.Background(), host, dht.Mode(dht.ModeServer))
			if err = kdht.Bootstrap(context.Background()); err != nil {
				fmt.Println("Error bootstrapping DHT:", err)
				return
			}
			for _, addr := range dht.DefaultBootstrapPeers {
				peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
				if err := host.Connect(context.Background(), *peerinfo); err != nil {
					fmt.Println("Error: connecting to bootstrap node:", err)
				} else {
					fmt.Println("OK: Connected to bootstrap node:", addr)
				}
			}
			routingDiscovery := routing.NewRoutingDiscovery(kdht)
			// Advertise the session ID
			dutil.Advertise(context.Background(), routingDiscovery, sessionID)
			// Find the peer with the session ID
			done := make(chan struct{})
			streamHandler := func(s network.Stream) {
				defer func() {
					s.Close()
					done <- struct{}{}
				}()
				fmt.Println("Got a new stream!")
				fmt.Println("From peer:", s.Conn().RemotePeer())
				fmt.Println("peer addr:", s.Conn().RemoteMultiaddr())
				s.CloseRead() // write only end
				// send the ciphertext
				n, err := s.Write(ciphertext)
				if err != nil {
					fmt.Println("Error sending ciphertext:", err)
					return
				}
				fmt.Printf("Sent %d bytes of ciphertext\n", n)
				fmt.Printf("Closing stream. Drop done\n")

			}
			host.SetStreamHandler("/p2p/drop", streamHandler)
			// wait for the stream to close
			fmt.Println("waiting for peer to connect...")
			<-done
			fmt.Printf("Exiting drop\n")
		} else { // recv mode
			fmt.Println("recv mode")
			// step 1: from the 4 words derive the session id
			words := strings.Split(*words, "-")
			sessionID := SessionIDFromWords(words)
			fmt.Printf("session id: %s\n", sessionID)

			// step 2: connect to the peer
			host, err := libp2p.New()
			if err != nil {
				fmt.Println("Error creating host:", err)
				return
			}
			fmt.Printf("host id: %s\n", host.ID())
			kdht, err := dht.New(context.Background(), host, dht.Mode(dht.ModeClient))
			if err = kdht.Bootstrap(context.Background()); err != nil {
				fmt.Println("Error bootstrapping DHT:", err)
				return
			}
			for _, addr := range dht.DefaultBootstrapPeers {
				peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
				if err := host.Connect(context.Background(), *peerinfo); err != nil {
					fmt.Println("Error: connecting to bootstrap node:", err)
				} else {
					fmt.Println("OK: Connected to bootstrap node:", addr)
				}
			}
			routingDiscovery := routing.NewRoutingDiscovery(kdht)
			// Find the peer with the session ID
			peerChan, err := routingDiscovery.FindPeers(context.Background(), sessionID)
			if err != nil {
				fmt.Println("Error finding peer:", err)
				return
			}
			found := false
			var dropSenderPeerID peer.ID
			for range time.NewTicker(5 * time.Second).C {
				fmt.Println("One tick querying peer info redenzvous...")
				if found {
					fmt.Println("found peer; exit loop")
					break
				}
				for peer := range peerChan {
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
						defer cancel()
						if err = host.Connect(ctx, peer); err != nil {
							fmt.Println("Error connecting to peer:", err)
							return
						}
						fmt.Println("Connected to peer ID:", peer.ID)
						fmt.Println(" at address: ", peer.Addrs)
						found = true
						dropSenderPeerID = peer.ID
						fmt.Println("Found peer with session ID; making stream to request data...")
					}()
				}
			}
			// step 3: receive and decrypt the message
			s, err := host.NewStream(context.Background(), dropSenderPeerID, "/p2p/drop")
			if err != nil {
				fmt.Println("Error opening stream:", err)
				return
			}
			defer s.Close()
			s.CloseWrite() // read only end
			ciphertext, err := io.ReadAll(s)
			if err != nil {
				fmt.Println("Error reading ciphertext:", err)
				return
			}
			fmt.Printf("Received %d bytes of ciphertext\n", len(ciphertext))
			// decrypt the ciphertext
			ints, err := encryption.WordsToInts(words)
			if err != nil {
				fmt.Println("Error converting words to ints:", err)
				return
			}
			keyBytes, err := encryption.IntsToBytes(ints)
			if err != nil {
				fmt.Println("Error converting ints to bytes:", err)
				return
			}
			plaintext, err := encryption.Decrypt(ciphertext, keyBytes)
			if err != nil {
				fmt.Println("Error decrypting ciphertext:", err)
				return
			}
			fmt.Printf("Decrypted message: len %d\n", len(plaintext))
			// write the decrypted message to stdout
			n, err := os.Stdout.Write(plaintext)
			if err != nil {
				fmt.Println("Error writing to stdout:", err)
				return
			}
			fmt.Printf("Wrote %d bytes to stdout\n", n)
			fmt.Printf("Exiting drop\n")
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

// words plus current time in the 5min bucket; used as rendezvous point in peer discovery
func SessionIDFromWords(words []string) string {
	t := time.Now()
	t = t.Truncate(5 * time.Minute)
	var sb strings.Builder
	for _, w := range words {
		sb.WriteString(w)
	}
	sb.WriteString(fmt.Sprintf("%d", t.Unix()))
	return sb.String()

}

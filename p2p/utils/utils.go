package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
)

const keyFileName = "peer_id.key"

func LoadOrCreateIdentity(dir string) (crypto.PrivKey, error) {
	keyPath := path.Join(dir, keyFileName)
	if _, err := os.Stat(keyPath); err == nil {
		// Key file exists, load it
		data, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		keyBytes, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, err
		}
		privKey, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("Loaded existing peer ID")
		return privKey, nil
	}

	// Key file does not exist, create a new one
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	keyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(keyPath, []byte(base64.StdEncoding.EncodeToString(keyBytes)), 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to save peer ID to file: %w", err)
	}
	log.Info().Msg("Created new peer ID and saved to file")
	return privKey, nil
}

func ComputeSessionID() int64 {
	// Get the current Unix timestamp
	now := time.Now().Unix()

	// Truncate to a 1000s window by removing the seconds part
	sessionID := now / 1000 * 1000

	return sessionID
}

// party.ID is a string but really is 32Bytes binary []byte("string")
// NOTE: only for ed25519 public keys derived peer ID
func PeerIDToPartyID(pids peer.ID) (party.ID, error) {
	pk, err := pids.ExtractPublicKey()
	if err != nil {
		return "", err
	}
	bz, err := pk.Raw()
	if err != nil {
		return "", err
	}
	return party.ID(bz), nil
}

// NOTE: only for ed25519 public keys derived peer ID
func PartyIDToPeerID(pid party.ID) (peer.ID, error) {
	bz := []byte(pid)

	if len(bz) != 32 {
		return "", fmt.Errorf("invalid party ID length %d", len(bz))
	}
	pk, err := crypto.UnmarshalEd25519PublicKey(bz)
	if err != nil {
		return "", err
	}

	return peer.IDFromPublicKey(pk)
}

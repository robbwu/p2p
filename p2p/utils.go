package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
)

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

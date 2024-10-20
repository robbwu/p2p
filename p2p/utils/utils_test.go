package utils

import (
	"encoding/hex"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

func TestID(t *testing.T) {
	peerID := "12D3KooWEZZZQrjL82FdsbTb69V3RMACpUT3qVUzr1E3iUfpn55N"
	pid, err := peer.Decode(peerID)
	if err != nil {
		t.Fatalf("failed to decode peer ID: %v", err)
	}

	party, err := PeerIDToPartyID(pid)
	if err != nil {
		t.Fatalf("failed to convert peer ID to party ID: %v", err)
	}

	peer, err := PartyIDToPeerID(party)
	if err != nil {
		t.Fatalf("failed to convert party ID to peer ID: %v", err)
	}

	if peer != pid {
		t.Fatalf("peer ID mismatch: expected %s, got %s", pid, peer)
	}

}

func TestMsgMarshal(t *testing.T) {

	msgHex := "a86453534944584096fa59ab2525a34bee542a184883f0f5a82fe3ef3b96d90fd2300309214d3d3fb1f96697771d0609b53aa1c518e46011869e710c3659d1a2f216e96c2f8360e36446726f6d7820467f9a25d2b487831af514330ae1fcccb267105aff74b2af5b9e456714e2cd8562546f606850726f746f636f6c74636d702f6b657967656e2d7468726573686f6c646b526f756e644e756d626572026444617461584ea16a436f6d6d69746d656e745840061221ec45ea7bb2e4451865b378f81ada29e3f842d96ff3baa2a420ad946bd96e7f33590558c6473eb36b0e04178099fbda3639ffbf64cd8db8c319ba6412536942726f616463617374f57542726f616463617374566572696669636174696f6ef6"
	bz, err := hex.DecodeString(msgHex)
	if err != nil {
		t.Fatalf("failed to decode message: %v", err)
	}
	assert.Equal(t, len(bz), 277)
	msg := protocol.Message{}
	//err = cbor.UnmarshalBinary(bz, &msg)
	err = msg.UnmarshalBinary(bz)
	if err != nil {
		t.Logf("failed to unmarshal message: %v", err)
	}
	t.Logf("message: %x", msg.Data)
	//spew.Dump(msg)
	t.Fatalf("TODO")
}

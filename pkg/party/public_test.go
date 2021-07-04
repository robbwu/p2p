package party

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taurusgroup/cmp-ecdsa/pkg/math/curve"
	"github.com/taurusgroup/cmp-ecdsa/pkg/math/sample"
	"github.com/taurusgroup/cmp-ecdsa/pkg/paillier"
	"github.com/taurusgroup/cmp-ecdsa/pkg/params"
	"github.com/taurusgroup/cmp-ecdsa/pkg/pedersen"
)

func TestPublic_Validate(t *testing.T) {
	sk := paillier.NewSecretKey()
	p := sk.PublicKey
	N := p.N
	ped, _ := sk.GeneratePedersen()

	_, X := sample.ScalarPointPair()
	N2 := big.NewInt(1)
	N2.Add(N2, N)
	p2 := paillier.NewPublicKey(N2)
	type fields struct {
		ID       ID
		ECDSA    *curve.Point
		Paillier *paillier.PublicKey
		Pedersen *pedersen.Parameters
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"all ok",
			fields{
				"blabla",
				X,
				p,
				ped},
			false,
		},
		{"no ID",
			fields{
				"",
				X,
				p,
				ped},
			true,
		},
		{"no ped",
			fields{
				"",
				X,
				p,
				nil},
			true,
		},
		{"no paillier",
			fields{
				"",
				X,
				nil,
				ped},
			true,
		},
		{"missing S",
			fields{
				"",
				X,
				p,
				&pedersen.Parameters{
					N: N,
					S: nil,
					T: ped.T,
				}},
			true,
		},
		{"missing T",
			fields{
				"",
				X,
				p,
				&pedersen.Parameters{
					N: N,
					S: ped.S,
					T: nil,
				}},
			true,
		},
		{"different N",
			fields{
				"",
				X,
				p2,
				ped},
			true,
		},
		{"small ssid",
			fields{
				"",
				X,
				p,
				ped},
			true,
		},
		{"no ssid",
			fields{
				"",
				X,
				p,
				ped},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Public{
				ID:       tt.fields.ID,
				ECDSA:    tt.fields.ECDSA,
				Paillier: tt.fields.Paillier,
				Pedersen: tt.fields.Pedersen,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPublic_MarshalJSON(t *testing.T) {
	ssid := make([]byte, params.HashBytes)
	_, _ = rand.Read(ssid)
	sk := paillier.NewSecretKey()
	pk := sk.PublicKey
	ped, _ := sk.GeneratePedersen()
	p := Public{
		ID:       RandomIDs(1)[0],
		ECDSA:    curve.NewIdentityPoint().ScalarBaseMult(sample.Scalar()),
		Paillier: pk,
		Pedersen: ped,
	}

	data, err := json.Marshal(p)
	require.NoError(t, err, "marshalling failed")
	p2 := Public{}
	err = json.Unmarshal(data, &p2)
	require.NoError(t, err, "unmarshalling failed")
	require.Equal(t, p, p2, "unmarshalling gave different result")
	fmt.Println(string(data))
}
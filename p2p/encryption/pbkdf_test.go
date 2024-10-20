package encryption

import (
	"bytes"
	"testing"
)

func Test1(t *testing.T) {
	password := []byte("helo")
	//pw, err := utils.GetPassword("password: ")
	//password := []byte(pw)
	//if err != nil {
	//	t.Fatal(err)
	//}
	plaintext := []byte("what a world we are in!")

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("size of cyphertext: %d", len(encrypted))

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(plaintext, decrypted) != 0 {
		t.Fatalf("decrypted plaintext does not match original plaintext")
	}
}

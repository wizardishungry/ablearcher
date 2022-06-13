package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSign(t *testing.T) {

	const (
		keyPub  = "d4d24cbf31543ae5bb2061bd4f0b6fc25ac89f8a440eb051e3a56a1708ed2023"
		keyPriv = "4f80f59c3d9d90432ef727a971e801068994cd8ccec13428990de517658c964cd4d24cbf31543ae5bb2061bd4f0b6fc25ac89f8a440eb051e3a56a1708ed2023"
		msg     = "the owls are not what they seem"
	)

	kPub, err := hex.DecodeString(keyPub)
	if err != nil {
		t.Fatalf("hex.DecodeString(keyPub) %v", err)
	}

	kPriv, err := hex.DecodeString(keyPriv)
	if err != nil {
		t.Fatalf("hex.DecodeString(keyPriv) %v", err)
	}

	sig := ed25519.Sign(kPriv, []byte(msg))
	fmt.Println(hex.EncodeToString(sig))
	ok := ed25519.Verify(kPub, []byte(msg), sig)
	fmt.Println(ok)
}

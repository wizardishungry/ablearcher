package ablearcher

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"
)

const (
	keyPub      = "d4d24cbf31543ae5bb2061bd4f0b6fc25ac89f8a440eb051e3a56a1708ed2023"
	keyPriv     = "4f80f59c3d9d90432ef727a971e801068994cd8ccec13428990de517658c964cd4d24cbf31543ae5bb2061bd4f0b6fc25ac89f8a440eb051e3a56a1708ed2023"
	msg         = "the owls are not what they seem"
	expectedSig = "6f0a3d5ef5f9573cf90e47326d18fd57ccce51bd4a833cd13e4a4b13446dabec3c375d19025b969ef21e2b3aad5279200b9285c4c10084899408cdebe6bcc40d"
)

func TestSign(t *testing.T) {

	kPub, err := hex.DecodeString(keyPub)
	if err != nil {
		t.Fatalf("hex.DecodeString(keyPub) %v", err)
	}

	kPriv, err := hex.DecodeString(keyPriv)
	if err != nil {
		t.Fatalf("hex.DecodeString(keyPriv) %v", err)
	}

	sig := ed25519.Sign(kPriv, []byte(msg))
	sigAsHex := hex.EncodeToString(sig)
	fmt.Println(sigAsHex)
	if sigAsHex != expectedSig {
		t.Fatalf("sigAsHex!= expectedSig")
	}
	ok := ed25519.Verify(kPub, []byte(msg), sig)
	fmt.Println(ok)

	authHeader := "Spring-83 Signature=" + hex.EncodeToString(sig)

	auth, err := parseAuthHeader(authHeader)
	if err != nil {
		t.Fatalf("parseAuthHeader %v", err)
	}
	fmt.Println(sig)
	fmt.Println(auth)

}

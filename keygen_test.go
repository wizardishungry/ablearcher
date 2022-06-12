package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"regexp"
	"testing"
	"time"
)

var keyRegexp = regexp.MustCompile(`ed[0-9]{4}$`)

func TestKeygen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping keygen")
	}
	const goodKey = "1c6ffef2825b294274478bad8c80a7a610d38245a9fded18cd004c4a67ed2023"
	if !(keyRegexp.MatchString(goodKey)) {
		t.Fatalf("bad regexp")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	pub, priv, tries, err := genKey(ctx, [4]rune{'2', '0', '2', '3'}, 0)
	if err != nil {
		t.Fatalf("genKey %v %d tries", err, tries)
	}

	_, _ = pub, priv
	asHex := hex.EncodeToString(pub)
	fmt.Println(tries, "tries", asHex)
	if !(keyRegexp.MatchString(asHex)) {
		t.Fatalf("regexp does not match")
	}
}

package ablearcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	ctx := context.Background()
	s := &service{}
	server := httptest.NewServer(s.handler())
	defer server.Close()
	fmt.Println(server.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/"+keyPub, bytes.NewBufferString(msg))
	req.Header.Set(headerVersion, "83")
	req.Header.Set("Content-Type", "text/html;charset=utf-8")
	zero := time.Time{}
	req.Header.Set(headerIfUnmodifiedSince, zero.Format(http.TimeFormat))
	req.Header.Set(headerAuthorization, fmt.Sprintf("Spring-83 Signature=%s", expectedSig)) // borrow const

	if err != nil {
		t.Fatalf("http.NewRequestWithContext %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.DefaultClient.Do %v", err)
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		t.Fatalf("io.Copy %v", err)
	}
	fmt.Println("status", resp.Status)
	fmt.Println("body", string(b.Bytes()))
}

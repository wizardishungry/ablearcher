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
	s := &service{
		storageProvider: chainedStorage{
			&memoryStore{},
		},
	}
	server := httptest.NewServer(s.handler())
	defer server.Close()
	fmt.Println(server.URL)

	keyURL := server.URL + "/" + keyPub

	{
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, keyURL, bytes.NewBufferString(msg))
		if err != nil {
			t.Fatalf("http.NewRequestWithContext %v", err)
		}

		req.Header.Set(headerVersion, "83")
		req.Header.Set("Content-Type", "text/html;charset=utf-8")
		zero := time.Time{}
		req.Header.Set(headerIfUnmodifiedSince, zero.Format(http.TimeFormat))
		req.Header.Set(headerAuthorization, fmt.Sprintf("Spring-83 Signature=%s", expectedSig)) // borrow const

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("http.DefaultClient.Do %v", err)
		}
		resp.Body.Close()
		fmt.Println("put status", resp.Status)
	}

	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, keyURL, nil)
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
		resp.Body.Close()
		fmt.Println("get status", resp.Status)
		fmt.Println("body", string(b.Bytes()))
	}
}

func BenchmarkPut(b *testing.B) {
	ctx := context.Background()
	s := &service{
		storageProvider: &memoryStore{},
	}
	server := httptest.NewServer(s.handler())
	defer server.Close()
	fmt.Println(server.URL)

	keyURL := server.URL + "/" + keyPub

	for i := 0; i < b.N; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, keyURL, bytes.NewBufferString(msg))
		req.Header.Set(headerVersion, "83")
		req.Header.Set("Content-Type", "text/html;charset=utf-8")
		zero := time.Time{}
		req.Header.Set(headerIfUnmodifiedSince, zero.Format(http.TimeFormat))
		req.Header.Set(headerAuthorization, fmt.Sprintf("Spring-83 Signature=%s", expectedSig)) // borrow const

		if err != nil {
			b.Fatalf("http.NewRequestWithContext %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("http.DefaultClient.Do %v", err)
		}
		defer resp.Body.Close()
	}
}

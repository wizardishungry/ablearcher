package ablearcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGossip(t *testing.T) {
	ctx := context.Background()
	var svcs []*service
	var urls []string

	for i := 0; i < 2; i++ {
		s := &service{
			storageProvider: chainedStorage{
				&memoryStore{},
				newGossip(ctx),
			},
		}
		server := httptest.NewServer(s.handler())
		defer server.Close()
		fmt.Println(server.URL)
		svcs = append(svcs, s)
		urls = append(urls, server.URL)
	}

	err := svcs[0].AddPeer(ctx, urls[1])
	if err != nil {
		t.Fatalf("AddPeer %v", err)
	}

	{
		keyURL := urls[0] + "/" + keyPub
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, keyURL, bytes.NewBufferString(msg))
		if err != nil {
			t.Fatalf("http.NewRequestWithContext %v", err)
		}
		authBytes, err := hex.DecodeString(expectedSig)
		if err != nil {
			t.Fatalf("hex.DecodeString %v", err)
		}

		item := &storageRecordWithKey{
			storageRecord: &storageRecord{
				body: []byte(msg),
				auth: authBytes,
			},
		}
		setupPutRequest(req, item)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("http.DefaultClient.Do %v", err)
		}
		resp.Body.Close()
		fmt.Println("put status", resp.Status)
	}
	time.Sleep(time.Second) // propagate
	{
		keyURL := urls[1] + "/" + keyPub
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

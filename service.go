package ablearcher

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type service struct {
	storageProvider
}

type storageProvider interface{}

func (s *service) handler() http.Handler {
	return http.HandlerFunc(s.handleFunc)
}

func (s *service) handleFunc(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	if u.Path == "/" {
		s.handleRoot(w, r)
		return
	}
	key := u.Path[1:]
	switch r.Method {
	case http.MethodPut:
		s.handlePut(w, r, key)
		return
	case http.MethodGet:
		s.handleGet(w, r, key)
		return
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func (s *service) handleRoot(w http.ResponseWriter, r *http.Request) {
	u := r.URL

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(u.Path))
}

func (s *service) handleGet(w http.ResponseWriter, r *http.Request, key string) {

}

type spring83Post struct {
	key               []byte
	version           int
	ifUnmodifiedSince time.Time
	authorization     []byte
	body              []byte
}

const (
	headerVersion           = "Spring-Version"
	headerIfUnmodifiedSince = "If-Unmodified-Since"
	headerAuthorization     = "Authorization"
)

func parseSpring83PostFromRequest(r *http.Request) (*spring83Post, error) {

	ver, err := strconv.ParseInt(r.Header.Get(headerVersion), 10, 64)
	if err != nil {
		return nil, err
	}

	ifUnmodifiedSince, err := http.ParseTime(r.Header.Get(headerIfUnmodifiedSince))
	if err != nil {
		return nil, err
	}

	auth, err := parseAuthHeader(r.Header.Get(headerAuthorization))
	if err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(r.URL.Path[1:])
	if err != nil {
		return nil, err
	}

	s83p := &spring83Post{
		key:               key,
		version:           int(ver),
		ifUnmodifiedSince: ifUnmodifiedSince,
		authorization:     auth,
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r.Body)
	if err != nil {
		return nil, err
	}
	s83p.body = buf.Bytes()

	return s83p, nil
}

func (s83p *spring83Post) Verify() bool {
	return ed25519.Verify(ed25519.PublicKey(s83p.key), s83p.body, s83p.authorization)
}

var ErrMissingAuth = fmt.Errorf("missing Spring-83 Authorization header")

func parseAuthHeader(ah string) ([]byte, error) {
	var sigBytes []byte
	if ah == "" {
		return nil, ErrMissingAuth
	}
	n, err := fmt.Sscanf(ah, "Spring-83 Signature=%x", &sigBytes)
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, fmt.Errorf("auth failed to parse")
	}
	return sigBytes, nil
}

func (s *service) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	s83p, err := parseSpring83PostFromRequest(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+" "+err.Error(), http.StatusBadRequest)
		return
	}
	if !s83p.Verify() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("%+v", s83p)))
}

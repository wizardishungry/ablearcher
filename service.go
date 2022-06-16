package ablearcher

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type service struct {
	storageProvider
}

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
	ctx := r.Context()

	keyArr, err := parseKeyFromRequestURL(r)
	if err != nil {

	}
	sr, err := s.storageProvider.Get(ctx, keyArr)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(sr.body)
}

type spring83Post struct {
	key               storageKey
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

func parseKeyFromRequestURL(r *http.Request) (s storageKey, err error) {
	key, err := hex.DecodeString(r.URL.Path[1:])
	if err != nil {
		return s, err
	}
	return sliceToStorageKey(key)
}

func parseSpring83PostFromRequest(r *http.Request) (*spring83Post, error) {
	keyArr, err := parseKeyFromRequestURL(r)
	if err != nil {
		return nil, err
	}

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

	s83p := &spring83Post{
		key:               keyArr,
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

func (s83p *spring83Post) Validate() error {
	return nil
}

func (s83p *spring83Post) Verify() bool {
	key := s83p.key[0:len(s83p.key)] // TODO move to helper
	return ed25519.Verify(ed25519.PublicKey(key), s83p.body, s83p.authorization)
}

var ErrMissingAuth = fmt.Errorf("missing Spring-83 Authorization header")

func parseAuthHeader(ah string) ([]byte, error) {
	var sigBytes []byte
	if ah == "" {
		return nil, ErrMissingAuth
	}
	fmt.Println("auth is", ah)
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
	ctx := r.Context()
	s83p, err := parseSpring83PostFromRequest(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+" "+err.Error(), http.StatusBadRequest)
		return
	}
	if !s83p.Verify() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	sr := &storageRecord{
		body: s83p.body,
		auth: s83p.authorization,
		time: s83p.ifUnmodifiedSince, // TODO needs to come out of sig & be validated
	}

	err = s.storageProvider.Put(ctx, s83p.key, sr)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("%+v", s83p)))
}

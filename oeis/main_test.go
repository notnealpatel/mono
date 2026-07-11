package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOeisRemoteWireFormat(t *testing.T) {
	var gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		io.WriteString(w, "ok")
	}))
	defer srv.Close()
	t.Setenv("SUPERMARKET_URL", srv.URL)

	if err := oeisRemote("show", "A000045"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/oeis" {
		t.Errorf("path = %q, want /oeis", gotPath)
	}
	var req struct {
		Subcommand string `json:"subcommand"`
		Query      string `json:"query"`
	}
	if err := json.Unmarshal(gotBody, &req); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if req.Subcommand != "show" || req.Query != "A000045" {
		t.Errorf("request = %+v, want {show, A000045}", req)
	}
}

func TestOeisRemoteError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()
	t.Setenv("SUPERMARKET_URL", srv.URL)

	if err := oeisRemote("match", "1,1,2,3,5"); err == nil {
		t.Fatal("want error on non-200 response")
	}
}

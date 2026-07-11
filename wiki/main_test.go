package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWikiRemoteWireFormat(t *testing.T) {
	var gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		io.WriteString(w, "ok")
	}))
	defer srv.Close()
	t.Setenv("SUPERMARKET_URL", srv.URL)

	if err := wikiRemote("article", "Erdos number"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/wikipedia" {
		t.Errorf("path = %q, want /wikipedia", gotPath)
	}
	var req struct {
		Subcommand string `json:"subcommand"`
		Query      string `json:"query"`
	}
	if err := json.Unmarshal(gotBody, &req); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if req.Subcommand != "article" || req.Query != "Erdos number" {
		t.Errorf("request = %+v, want {article, Erdos number}", req)
	}
}

func TestWikiRemoteError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()
	t.Setenv("SUPERMARKET_URL", srv.URL)

	if err := wikiRemote("search", "x"); err == nil {
		t.Fatal("want error on non-200 response")
	}
}

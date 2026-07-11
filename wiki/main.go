// Package main implements a thin client around the
// supermarket retrieval endpoint for Wikipedia articles.
//
// All commands POST the retrieval wire format
// ({"subcommand": ..., "query": ...}) to /wikipedia on
// the supermarket:9001 tailscale node and write the
// response body to stdout.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprint(os.Stdout, usage)
		os.Exit(0)
	}

	var err error
	switch os.Args[1] {
	case "article", "links", "search":
		err = wikiRemote(os.Args[1], strings.Join(os.Args[2:], " "))
	default:
		fmt.Fprint(os.Stdout, usage)
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "goof-wiki: %v\n", err)
		os.Exit(1)
	}
}

const usage = `usage: wiki <command> [args]

  article <title>    article text as markdown
  links <title>      outgoing links
  search <query>     search titles
`

func wikiRemote(subcommand, query string) error {
	base := os.Getenv("SUPERMARKET_URL")
	if base == "" {
		base = "http://supermarket:9001"
	}
	body, err := json.Marshal(struct {
		Subcommand string `json:"subcommand"`
		Query      string `json:"query"`
	}{subcommand, query})
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(base+"/wikipedia", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supermarket returned %s: %s", resp.Status, bytes.TrimSpace(msg))
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

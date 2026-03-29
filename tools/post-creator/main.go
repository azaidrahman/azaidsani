package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := flag.Int("port", 3333, "server port")
	project := flag.String("project", "../..", "path to Hugo project root")
	flag.Parse()

	projectRoot, err := filepath.Abs(*project)
	if err != nil {
		log.Fatalf("invalid project path: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectRoot, "hugo.toml")); err != nil {
		log.Fatalf("no hugo.toml found at %s — is --project pointing to your Hugo project root?", projectRoot)
	}

	srv, err := NewServer(projectRoot)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Post Creator running at http://localhost:%d\n", *port)
	fmt.Printf("Hugo project: %s\n", projectRoot)
	log.Fatal(http.ListenAndServe(addr, srv.Router()))
}

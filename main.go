package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"localnetwork/network"
)

//go:embed web/*
var webFS embed.FS

var server *http.Server

func main() {
	mux := http.NewServeMux()

	// Static files (HTML/JS/CSS)
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	// Interfaces endpoint
	mux.HandleFunc("/api/interfaces", func(w http.ResponseWriter, r *http.Request) {
		ifaces, err := network.ListInterfaces()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, ifaces)
	})

	// Health check (ping) endpoint
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		target := strings.TrimSpace(r.URL.Query().Get("target"))
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}
		count := 3
		timeout := 5 * time.Second
		result := network.PingOnce(target, count, timeout)
		writeJSON(w, result)
	})

	// Network discovery (ARP table)
	mux.HandleFunc("/api/discover", func(w http.ResponseWriter, r *http.Request) {
		results, err := network.DiscoverLAN()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, results)
	})

	// Internet probe (device-local)
	mux.HandleFunc("/api/internet", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
		defer cancel()
		result := network.ProbeInternet(ctx)
		writeJSON(w, result)
	})

	// Shutdown endpoint
	mux.HandleFunc("/api/shutdown", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "shutting down"})
		go func() {
			time.Sleep(100 * time.Millisecond) // Give time for response to be sent
			log.Println("Shutdown requested via web interface")
			if server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := server.Shutdown(ctx); err != nil {
					log.Printf("Server shutdown error: %v", err)
				}
			}
			os.Exit(0)
		}()
	})

	addr := "127.0.0.1:8080"
	server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("Local Network Monitor listening on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode: %v", err), http.StatusInternalServerError)
	}
}



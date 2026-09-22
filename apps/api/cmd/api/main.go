package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(healthResponse{
			Status: "ok", Service: "sodienthoai-api", Version: "0.1.0",
		})
	})

	server := &http.Server{Addr: ":" + port, Handler: mux}
	log.Printf("sodienthoai api listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

package main

import (
	"checkers-solver/api"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", api.HandleHealth)
	mux.HandleFunc("/api/solve", api.HandleSolve)

	handler := api.CORSMiddleware(mux)

	addr := ":8080"
	log.Printf("Checkers solver backend starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

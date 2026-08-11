package main

import (
	"fmt"
	"net/http"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func main() {
	http.HandleFunc("/healthz", healthzHandler)
	fmt.Println("Server starting on :8040")
	if err := http.ListenAndServe(":8040", nil); err != nil {
		panic(err)
	}
}

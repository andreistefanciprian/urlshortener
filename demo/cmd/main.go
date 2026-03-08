// demo is a minimal HTTP server for testing release-please.
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	log.Println("demo: server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
